package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/project"
	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/services/ai"
	"github.com/redis/go-redis/v9"

	. "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	"github.com/google/uuid"
)

type AdminQuestionnaireService struct {
	repo        *questionnaire.QuestionnaireRepository
	aiService   ai.AIService
	redisClient *redis.Client
}

func NewAdminQuestionnaireService(repo *questionnaire.QuestionnaireRepository, aiService ai.AIService, redisClient *redis.Client) *AdminQuestionnaireService {
	return &AdminQuestionnaireService{
		repo:        repo,
		aiService:   aiService,
		redisClient: redisClient,
	}
}

// Target Role CRUD Operations
func (s *AdminQuestionnaireService) CreateTargetRole(req CreateTargetRoleRequest) (*TargetRoleResponse, error) {
	targetRole := &project.TargetRole{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.repo.CreateTargetRole(targetRole)
	if err != nil {
		return nil, errors.New("failed to create target role")
	}

	return s.mapTargetRoleToResponse(targetRole), nil
}

func (s *AdminQuestionnaireService) GetTargetRoles(page, limit int) (*PaginatedTargetRolesResponse, error) {
	log.Printf("DEBUG: Service GetTargetRoles called with page=%d, limit=%d", page, limit)

	offset := (page - 1) * limit
	roles, total, err := s.repo.GetTargetRoles(offset, limit)
	if err != nil {
		log.Printf("DEBUG: Repository error: %v", err)
		return nil, err
	}

	log.Printf("DEBUG: Service received %d roles, total=%d", len(roles), total)

	roleResponses := make([]TargetRoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = *s.mapTargetRoleToResponse(&role)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedTargetRolesResponse{
		Data:       roleResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminQuestionnaireService) GetTargetRoleByID(id uuid.UUID) (*TargetRoleResponse, error) {
	role, err := s.repo.GetTargetRoleByID(id)
	if err != nil {
		return nil, errors.New("target role not found")
	}

	return s.mapTargetRoleToResponse(role), nil
}

func (s *AdminQuestionnaireService) UpdateTargetRole(id uuid.UUID, req UpdateTargetRoleRequest) (*TargetRoleResponse, error) {
	role, err := s.repo.GetTargetRoleByID(id)
	if err != nil {
		return nil, errors.New("target role not found")
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Category != nil {
		role.Category = *req.Category
	}
	if req.Active != nil {
		role.Active = *req.Active
	}
	role.UpdatedAt = time.Now()

	err = s.repo.UpdateTargetRole(role)
	if err != nil {
		return nil, errors.New("failed to update target role")
	}

	return s.mapTargetRoleToResponse(role), nil
}

func (s *AdminQuestionnaireService) DeleteTargetRole(id uuid.UUID) error {
	return s.repo.DeleteTargetRole(id)
}

// AI-powered questionnaire generation
func (s *AdminQuestionnaireService) GenerateQuestionnaire(req GenerateQuestionnaireRequest) (*QuestionnaireGenerationResponse, error) {
	// Automatically get ALL active target roles from database
	targetRoles, err := s.getAllActiveTargetRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get active target roles: %w", err)
	}

	if len(targetRoles) == 0 {
		return nil, errors.New("no active target roles found in database")
	}

	// Extract role names for AI prompt
	targetRoleNames := make([]string, len(targetRoles))
	for i, role := range targetRoles {
		targetRoleNames[i] = role.Name
	}

	questionnaire := &questionnaire.ProfilingQuestionnaire{
		Name:        req.Name,
		GeneratedBy: "ai",
		Version:     1,
		Active:      false,
	}

	err = s.repo.CreateQuestionnaire(questionnaire)
	if err != nil {
		return nil, errors.New("failed to create questionnaire")
	}

	prompt := s.buildQuestionGenerationPrompt(req, targetRoleNames)

	go s.processQuestionGeneration(questionnaire.ID, prompt, req)

	return &QuestionnaireGenerationResponse{
		QuestionnaireID: questionnaire.ID,
		Status:          "processing",
		Progress:        0,
		Message:         "Questionnaire generation started",
	}, nil
}

func (s *AdminQuestionnaireService) GetQuestionnaires(page, limit int) (*PaginatedQuestionnairesResponse, error) {
	offset := (page - 1) * limit
	questionnaires, total, err := s.repo.GetQuestionnairesNew(offset, limit)
	if err != nil {
		return nil, err
	}

	questionnaireResponses := make([]QuestionnaireListResponse, len(questionnaires))
	for i, q := range questionnaires {
		questionnaireResponses[i] = QuestionnaireListResponse{
			ID:          q.ID,
			Name:        q.Name,
			Description: "", // Not available in ProfilingQuestionnaire model
			Version:     fmt.Sprintf("v%d", q.Version),
			TargetRoles: []TargetRoleResponse{}, // TODO: Load target roles if needed
			Active:      q.Active,
			CreatedAt:   q.CreatedAt,
			UpdatedAt:   q.UpdatedAt,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedQuestionnairesResponse{
		Data:       questionnaireResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminQuestionnaireService) GetQuestionnaireDetail(id uuid.UUID) (*QuestionnaireDetailResponse, error) {
	questionnaire, err := s.repo.GetQuestionnaireByID(id)
	if err != nil {
		return nil, errors.New("questionnaire not found")
	}

	// Get target roles (from junction table)
	targetRoles, err := s.repo.GetTargetRolesByQuestionnaireID(id)
	if err != nil {
		return nil, err
	}

	// Map target roles to response format
	targetRoleResponses := make([]TargetRoleResponse, len(targetRoles))
	for i, role := range targetRoles {
		targetRoleResponses[i] = *s.mapTargetRoleToResponse(&role)
	}

	// Get total submissions
	totalSubmissions, err := s.repo.GetQuestionnaireSubmissionCountNew(id)
	if err != nil {
		totalSubmissions = 0 // Default to 0 if error
	}

	// Get questions for the questionnaire
	questions, err := s.repo.GetQuestionsByQuestionnaireID(id)
	if err != nil {
		return nil, err
	}

	// Map questions to response format
	questionResponses := make([]GeneratedQuestionResponse, len(questions))
	for i, q := range questions {
		var options []OptionDTO
		if q.Options != nil && *q.Options != "" {
			var rawOptions []struct {
				Text  string `json:"text"`
				Score int    `json:"score"`
			}
			if err := json.Unmarshal([]byte(*q.Options), &rawOptions); err == nil {
				for _, opt := range rawOptions {
					options = append(options, OptionDTO{
						Text:  opt.Text,
						Value: opt.Text, // Using text as value for compatibility
						Score: opt.Score,
					})
				}
			}
		}

		questionResponses[i] = GeneratedQuestionResponse{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			QuestionType: string(q.QuestionType),
			Category:     q.Category,
			Options:      options,
			MaxScore:     q.MaxScore,
			Order:        q.QuestionOrder, // Use QuestionOrder from model
		}
	}

	return &QuestionnaireDetailResponse{
		ID:               questionnaire.ID,
		Name:             questionnaire.Name,
		Description:      "", // Model doesn't have description field
		Version:          fmt.Sprintf("v%d", questionnaire.Version),
		TargetRoles:      targetRoleResponses,
		Questions:        questionResponses,
		Active:           questionnaire.Active,
		TotalSubmissions: totalSubmissions,
		CreatedAt:        questionnaire.CreatedAt,
		UpdatedAt:        questionnaire.UpdatedAt,
	}, nil
}

func (s *AdminQuestionnaireService) ActivateQuestionnaire(id uuid.UUID, active bool) error {
	if active {
		// Activate the questionnaire (this deactivates all others and activates this one)
		return s.repo.ActivateQuestionnaire(id)
	} else {
		// Deactivate just this questionnaire
		return s.repo.DeactivateAllQuestionnaires() // For now, deactivate all questionnaires
	}
}

func (s *AdminQuestionnaireService) GetQuestionnaireResponses(page, limit int, questionnaireID *uuid.UUID) (*PaginatedResponsesResponse, error) {
	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Get responses from repository
	responses, total, err := s.repo.GetQuestionnaireResponsesNew(offset, limit, questionnaireID)
	if err != nil {
		return nil, fmt.Errorf("failed to get questionnaire responses: %w", err)
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Map responses to DTO
	responseOverviews := make([]QuestionnaireResponseOverview, len(responses))
	for i, response := range responses {
		// Get student profile and user info
		studentProfile, err := s.repo.GetStudentByProfileIDNew(response.StudentProfileID)
		if err != nil {
			log.Printf("Warning: failed to get student profile for response %s: %v", response.ID, err)
			continue
		}

		// Get questionnaire info
		questionnaire, err := s.repo.GetQuestionnaireByID(response.QuestionnaireID)
		if err != nil {
			log.Printf("Warning: failed to get questionnaire for response %s: %v", response.ID, err)
			continue
		}

		// Get user info
		var studentName, studentEmail string
		studentName = studentProfile.Fullname
		studentEmail = "unknown" // Default fallback

		// Try to get email from preloaded User relationship
		if studentProfile.User.Email != "" {
			studentEmail = studentProfile.User.Email
		}

		// Calculate scores
		totalScore := 0
		if response.TotalScore != nil {
			totalScore = *response.TotalScore
		}

		// Calculate max score (assuming we can get this from questionnaire questions)
		maxScore := s.calculateMaxScore(questionnaire.Questions)

		// Calculate percentage
		scorePercentage := 0.0
		if maxScore > 0 {
			scorePercentage = (float64(totalScore) / float64(maxScore)) * 100
		}

		// Get top recommendations (simplified for now)
		topRecommendations := []TopRecommendationDTO{}
		if response.AIRecommendations != nil && *response.AIRecommendations != "" {
			// Parse AI recommendations if needed - for now, return empty array
			// TODO: Implement proper recommendation parsing
		}

		// Determine processing status
		processingStatus := "completed"
		if response.ProcessedAt == nil {
			processingStatus = "pending"
		} else if response.AIAnalysis == nil {
			processingStatus = "processing"
		}

		responseOverviews[i] = QuestionnaireResponseOverview{
			ID:                 response.ID,
			QuestionnaireID:    response.QuestionnaireID,
			QuestionnaireName:  questionnaire.Name,
			StudentName:        studentName,
			StudentEmail:       studentEmail,
			TotalScore:         totalScore,
			MaxScore:           maxScore,
			ScorePercentage:    scorePercentage,
			TopRecommendations: topRecommendations,
			ProcessingStatus:   processingStatus,
			SubmittedAt:        response.SubmittedAt,
		}
	}

	return &PaginatedResponsesResponse{
		Data:       responseOverviews,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// Helper method to calculate max score from questions
func (s *AdminQuestionnaireService) calculateMaxScore(questions []questionnaire.QuestionnaireQuestion) int {
	maxScore := 0
	for _, question := range questions {
		maxScore += question.MaxScore
	}
	return maxScore
}
func (s *AdminQuestionnaireService) GetResponseDetail(id uuid.UUID) (*ResponseDetailResponse, error) {
	// Get the response from repository
	response, err := s.repo.GetQuestionnaireResponseByIDNew(id)
	if err != nil {
		return nil, fmt.Errorf("response not found: %w", err)
	}

	// Get questionnaire info
	questionnaire, err := s.repo.GetQuestionnaireByID(response.QuestionnaireID)
	if err != nil {
		return nil, fmt.Errorf("questionnaire not found: %w", err)
	}

	// Get student profile and user info
	studentProfile, err := s.repo.GetStudentByProfileIDNew(response.StudentProfileID)
	if err != nil {
		return nil, fmt.Errorf("student profile not found: %w", err)
	}

	// Get questions for detailed answers
	questions, err := s.repo.GetQuestionsByQuestionnaireID(response.QuestionnaireID)
	if err != nil {
		return nil, fmt.Errorf("questions not found: %w", err)
	}

	// Parse answers from JSON
	var answers []struct {
		QuestionID string `json:"question_id"`
		Answer     string `json:"answer"`
		Score      int    `json:"score"`
	}
	if err := json.Unmarshal([]byte(response.Answers), &answers); err != nil {
		return nil, fmt.Errorf("failed to parse answers: %w", err)
	}

	// Create question map for easier lookup
	questionMap := make(map[uuid.UUID]QuestionnaireQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// Map answers to detailed format
	var detailedAnswers []DetailedAnswerDTO
	totalScore := 0
	maxScore := 0

	for _, answer := range answers {
		questionID, err := uuid.Parse(answer.QuestionID)
		if err != nil {
			continue
		}

		if question, exists := questionMap[questionID]; exists {
			detailedAnswers = append(detailedAnswers, DetailedAnswerDTO{
				QuestionID:   questionID,
				QuestionText: question.QuestionText,
				Answer:       answer.Answer,
				Score:        answer.Score,
				MaxScore:     question.MaxScore,
				Category:     question.Category,
			})
			totalScore += answer.Score
			maxScore += question.MaxScore
		}
	}

	// Calculate score percentage
	scorePercentage := 0.0
	if maxScore > 0 {
		scorePercentage = (float64(totalScore) / float64(maxScore)) * 100
	}

	// Parse AI recommendations if available
	var recommendations []DetailedRecommendationDTO
	if response.AIRecommendations != nil && *response.AIRecommendations != "" {
		var aiRecs []struct {
			RoleID        string  `json:"role_id"`
			RoleName      string  `json:"role_name"`
			Score         float64 `json:"score"`
			Justification string  `json:"justification"`
			Category      string  `json:"category"`
		}
		if err := json.Unmarshal([]byte(*response.AIRecommendations), &aiRecs); err == nil {
			for _, rec := range aiRecs {
				roleID, _ := uuid.Parse(rec.RoleID)
				recommendations = append(recommendations, DetailedRecommendationDTO{
					ID:            roleID,
					RoleName:      rec.RoleName,
					Score:         rec.Score,
					Justification: rec.Justification,
					Category:      rec.Category,
					Active:        true,
				})
			}
		}
	}

	// Parse AI analysis if available
	var analysis *AnalysisResultDTO
	if response.AIAnalysis != nil && *response.AIAnalysis != "" {
		var aiAnalysis struct {
			PersonalityTraits []string `json:"personality_traits"`
			Interests         []string `json:"interests"`
			Strengths         []string `json:"strengths"`
			WorkStyle         string   `json:"work_style"`
		}
		if err := json.Unmarshal([]byte(*response.AIAnalysis), &aiAnalysis); err == nil {
			analysis = &AnalysisResultDTO{
				PersonalityTraits: aiAnalysis.PersonalityTraits,
				Interests:         aiAnalysis.Interests,
				Strengths:         aiAnalysis.Strengths,
				WorkStyle:         aiAnalysis.WorkStyle,
			}
		}
	}

	// Determine processing status
	processingStatus := "completed"
	if response.ProcessedAt == nil {
		processingStatus = "pending"
	} else if response.AIAnalysis == nil {
		processingStatus = "processing"
	}

	// Get student basic info
	studentInfo := StudentBasicInfo{
		ID:    studentProfile.ID,
		Name:  studentProfile.Fullname,
		Email: studentProfile.User.Email,
		NIM:   studentProfile.NIS, // Using NIS as NIM
	}

	return &ResponseDetailResponse{
		ID:                response.ID,
		QuestionnaireID:   response.QuestionnaireID,
		QuestionnaireName: questionnaire.Name,
		Student:           studentInfo,
		Answers:           detailedAnswers,
		TotalScore:        totalScore,
		MaxScore:          maxScore,
		ScorePercentage:   scorePercentage,
		Recommendations:   recommendations,
		Analysis:          analysis,
		ProcessingStatus:  processingStatus,
		SubmittedAt:       response.SubmittedAt,
	}, nil
}

// Helper methods
func (s *AdminQuestionnaireService) mapTargetRoleToResponse(role *project.TargetRole) *TargetRoleResponse {
	return &TargetRoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Category:    role.Category,
		Active:      role.Active,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
func (s *AdminQuestionnaireService) buildQuestionGenerationPrompt(req GenerateQuestionnaireRequest, targetRoleNames []string) string {
	customInstructions := ""
	if req.CustomInstructions != nil {
		customInstructions = *req.CustomInstructions
	}

	// Get hardcoded focus areas based on target roles
	focusAreas := GetDefaultFocusAreas(targetRoleNames)

	return fmt.Sprintf(`Anda adalah seorang ahli psikologi karir dan teknologi yang akan membuat kuesioner profiling karir untuk siswa SMK.

TUJUAN: Buat TEPAT %d pertanyaan yang dapat mengidentifikasi kecenderungan siswa terhadap peran teknologi tertentu.

TARGET ROLES: %v

LEVEL KESULITAN: %s
- basic: Pertanyaan mudah dipahami siswa SMK, fokus pada minat dasar
- intermediate: Pertanyaan yang menggali preferensi kerja dan gaya belajar  
- advanced: Pertanyaan mendalam tentang problem-solving dan thinking patterns

FOKUS AREA: %v

CUSTOM INSTRUCTIONS: %s

PENTING: HARUS MENGHASILKAN TEPAT %d PERTANYAAN, TIDAK BOLEH KURANG ATAU LEBIH!

FORMAT OUTPUT JSON:
{
  "questions": [
    {
      "question_text": "Pertanyaan yang relevan",
      "question_type": "likert|mcq|case|text",
      "options": [
        {"text": "Option text", "score": 0-5}
      ],
      "category": "technical|behavioral|interest|aptitude"
    }
  ]
}

RULES:
1. WAJIB: Hasilkan TEPAT %d pertanyaan, tidak boleh kurang!
2. Gunakan bahasa Indonesia yang mudah dipahami siswa SMK
3. Variasikan jenis pertanyaan (60%% likert, 25%% mcq, 15%% case/text)
4. Pastikan pertanyaan dapat membedakan antar target roles
5. Fokus pada minat, kepribadian, dan kemampuan teknis dasar
6. Hindari pertanyaan yang bias gender atau latar belakang sosial
7. Untuk mcq, berikan 4-5 opsi yang masuk akal
8. Untuk likert, gunakan skala 1-5 (sangat tidak setuju - sangat setuju)
9. Urutkan output JSON dengan urutan tipe pertanyaan: **mcq → likert → case → text**

VALIDASI AKHIR: Pastikan array "questions" memiliki TEPAT %d elemen!

Mulai generasi sekarang:`,
		req.QuestionCount, targetRoleNames, req.DifficultyLevel, focusAreas, customInstructions,
		req.QuestionCount, req.QuestionCount, req.QuestionCount)
}

func (s *AdminQuestionnaireService) processQuestionGeneration(questionnaireID uuid.UUID, promptUsed string, req GenerateQuestionnaireRequest) {
	ctx := context.Background()

	s.setGenerationStatus(ctx, questionnaireID, "processing", 10, "Initializing AI question generation")
	time.Sleep(2 * time.Second)

	var aiResponse *ai.QuestionGenerationResponse
	var err error

	if s.aiService != nil {
		s.setGenerationStatus(ctx, questionnaireID, "processing", 30, "Generating questions with AI")
		aiResponse, err = s.aiService.GenerateQuestions(ctx, promptUsed)
		if err != nil {
			s.setGenerationStatus(ctx, questionnaireID, "failed", 0, fmt.Sprintf("AI generation failed: %v", err))
			return
		}
	} else {
		s.setGenerationStatus(ctx, questionnaireID, "failed", 0, "AI service not available")
		return
	}

	s.setGenerationStatus(ctx, questionnaireID, "processing", 50, "Validating AI response")

	if len(aiResponse.Questions) != req.QuestionCount {
		s.setGenerationStatus(ctx, questionnaireID, "failed", 0,
			fmt.Sprintf("AI generated %d questions but %d were requested", len(aiResponse.Questions), req.QuestionCount))
		return
	}

	s.setGenerationStatus(ctx, questionnaireID, "processing", 70, "Fetching questionnaire data")
	questionnaireEntity, err := s.repo.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		s.setGenerationStatus(ctx, questionnaireID, "failed", 0, fmt.Sprintf("Error fetching questionnaire: %v", err))
		return
	}

	s.setGenerationStatus(ctx, questionnaireID, "processing", 80, "Processing questions")
	questions := make([]questionnaire.QuestionnaireQuestion, len(aiResponse.Questions))
	for i, aiQ := range aiResponse.Questions {
		var optionsJSON *string
		if len(aiQ.Options) > 0 {
			optionsBytes, _ := json.Marshal(aiQ.Options)
			optionsStr := string(optionsBytes)
			optionsJSON = &optionsStr
		}

		questionType := questionnaire.QuestionType(aiQ.QuestionType)
		maxScore := s.getMaxScoreForType(questionType)

		questions[i] = questionnaire.QuestionnaireQuestion{
			QuestionnaireID: questionnaireID,
			QuestionText:    aiQ.QuestionText,
			QuestionType:    questionType,
			Options:         optionsJSON,
			MaxScore:        maxScore,
			QuestionOrder:   i + 1,
			Category:        aiQ.Category,
		}
	}

	s.setGenerationStatus(ctx, questionnaireID, "processing", 90, "Saving questions to database")
	err = s.repo.AddQuestionsToQuestionnaire(questionnaireID, questions)
	if err != nil {
		s.setGenerationStatus(ctx, questionnaireID, "failed", 0, fmt.Sprintf("Error saving questions: %v", err))
		return
	}

	questionnaireEntity.AIPromptUsed = &promptUsed
	err = s.repo.UpdateQuestionnaire(questionnaireEntity)
	if err != nil {
		s.setGenerationStatus(ctx, questionnaireID, "failed", 0, fmt.Sprintf("Error updating questionnaire: %v", err))
		return
	}

	s.setGenerationStatus(ctx, questionnaireID, "completed", 100, fmt.Sprintf("Successfully generated %d questions", len(questions)))
}

func (s *AdminQuestionnaireService) getMaxScoreForType(questionType questionnaire.QuestionType) int {
	switch questionType {
	case questionnaire.QuestionTypeLikert:
		return 5
	case questionnaire.QuestionTypeMCQ:
		return 1
	case questionnaire.QuestionTypeCase:
		return 5
	case questionnaire.QuestionTypeText:
		return 0
	default:
		return 1
	}
}

func (s *AdminQuestionnaireService) getAllActiveTargetRoles() ([]project.TargetRole, error) {
	log.Printf("Getting all active target roles from database")

	roles, _, err := s.repo.GetTargetRoles(0, 1000)
	if err != nil {
		log.Printf("Failed to get all target roles: %v", err)
		return nil, fmt.Errorf("failed to get all target roles: %w", err)
	}

	// Filter only active roles
	var activeRoles []project.TargetRole
	for _, role := range roles {
		if role.Active {
			activeRoles = append(activeRoles, role)
		}
	}

	log.Printf("Found %d active target roles out of %d total roles", len(activeRoles), len(roles))
	return activeRoles, nil
}

func (s *AdminQuestionnaireService) getTargetRoleNames(roleIDs []string) ([]string, error) {
	var names []string

	log.Printf("Getting target role names for IDs: %+v", roleIDs)

	for _, idStr := range roleIDs {
		// Skip empty strings
		if idStr == "" {
			log.Printf("Skipping empty role ID")
			continue
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Failed to parse UUID: %s, error: %v", idStr, err)
			return nil, fmt.Errorf("invalid role ID: %s", idStr)
		}

		// Get target role from database
		role, err := s.repo.GetTargetRoleByID(id)
		if err != nil {
			log.Printf("Failed to get target role by ID %s: %v", id, err)
			return nil, fmt.Errorf("failed to get target role: %s", err.Error())
		}

		names = append(names, role.Name)
		log.Printf("Found role: %s for ID: %s", role.Name, id)
	}

	if len(names) == 0 {
		log.Printf("No valid role IDs provided")
		return nil, fmt.Errorf("no valid role IDs provided")
	}

	log.Printf("Successfully retrieved role names: %+v", names)
	return names, nil
}

// getDefaultRoleName returns a placeholder role name based on ID
// This would be replaced with actual database lookup in production
func (s *AdminQuestionnaireService) getDefaultRoleName(id uuid.UUID) string {
	// Generate role name based on ID pattern for testing
	idStr := id.String()

	// Simple mapping based on first character of UUID
	switch idStr[0] {
	case '1', '2', '3':
		return "Software Developer"
	case '4', '5', '6':
		return "Data Analyst"
	case '7', '8', '9':
		return "UI/UX Designer"
	case 'a', 'b', 'c':
		return "Project Manager"
	case 'd', 'e', 'f':
		return "Business Analyst"
	default:
		return "Technology Specialist"
	}
}

type GenerationStatus struct {
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Message  string `json:"message"`
}

func (s *AdminQuestionnaireService) setGenerationStatus(ctx context.Context, questionnaireID uuid.UUID, status string, progress int, message string) {
	err := s.repo.UpdateQuestionnaireGenerationStatus(questionnaireID, status, progress, message)
	if err != nil {
		return
	}

	if s.redisClient != nil {
		statusData := GenerationStatus{
			Status:   status,
			Progress: progress,
			Message:  message,
		}

		statusJSON, _ := json.Marshal(statusData)
		key := fmt.Sprintf("questionnaire:generation:%s", questionnaireID.String())
		s.redisClient.Set(ctx, key, string(statusJSON), 5*time.Minute)
	}
}

func (s *AdminQuestionnaireService) GetGenerationStatus(questionnaireID uuid.UUID) (*GenerationStatus, error) {
	ctx := context.Background()
	key := fmt.Sprintf("questionnaire:generation:%s", questionnaireID.String())

	if s.redisClient != nil {
		result, err := s.redisClient.Get(ctx, key).Result()
		if err == nil {
			var status GenerationStatus
			if json.Unmarshal([]byte(result), &status) == nil {
				return &status, nil
			}
		}
	}

	dbQuestionnaire, err := s.repo.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		return &GenerationStatus{
			Status:   "not_found",
			Progress: 0,
			Message:  "Generation status not found",
		}, nil
	}

	status := &GenerationStatus{
		Status:   dbQuestionnaire.GenerationStatus,
		Progress: dbQuestionnaire.GenerationProgress,
		Message:  dbQuestionnaire.GenerationMessage,
	}

	if s.redisClient != nil {
		statusJSON, _ := json.Marshal(status)
		s.redisClient.Set(ctx, key, string(statusJSON), 5*time.Minute)
	}

	return status, nil
}
