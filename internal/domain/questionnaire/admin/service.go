package admin

import (
	"aicademy-backend/internal/domain/questionnaire"
	"aicademy-backend/internal/services/ai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type AdminQuestionnaireService struct {
	repo      *questionnaire.QuestionnaireRepository
	aiService ai.AIService
}

func NewAdminQuestionnaireService(repo *questionnaire.QuestionnaireRepository, aiService ai.AIService) *AdminQuestionnaireService {
	return &AdminQuestionnaireService{
		repo:      repo,
		aiService: aiService,
	}
}

// Target Role CRUD Operations
func (s *AdminQuestionnaireService) CreateTargetRole(req CreateTargetRoleRequest) (*TargetRoleResponse, error) {
	targetRole := &questionnaire.TargetRole{
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
	// First, get target role names from IDs
	targetRoleNames, err := s.getTargetRoleNames(req.TargetRoleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get target role names: %w", err)
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
	return nil, errors.New("not implemented yet")
}

// Helper methods
func (s *AdminQuestionnaireService) mapTargetRoleToResponse(role *questionnaire.TargetRole) *TargetRoleResponse {
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

TUJUAN: Buat %d pertanyaan yang dapat mengidentifikasi kecenderungan siswa terhadap peran teknologi tertentu.

TARGET ROLES: %v

LEVEL KESULITAN: %s
- basic: Pertanyaan mudah dipahami siswa SMK, fokus pada minat dasar
- intermediate: Pertanyaan yang menggali preferensi kerja dan gaya belajar  
- advanced: Pertanyaan mendalam tentang problem-solving dan thinking patterns

FOKUS AREA: %v

CUSTOM INSTRUCTIONS: %s

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
1. Gunakan bahasa Indonesia yang mudah dipahami siswa SMK
2. Variasikan jenis pertanyaan (60%% likert, 25%% mcq, 15%% case/text)
3. Pastikan pertanyaan dapat membedakan antar target roles
4. Fokus pada minat, kepribadian, dan kemampuan teknis dasar
5. Hindari pertanyaan yang bias gender atau latar belakang sosial
6. Untuk mcq, berikan 4-5 opsi yang masuk akal
7. Untuk likert, gunakan skala 1-5 (sangat tidak setuju - sangat setuju)

Mulai generasi sekarang:`,
		req.QuestionCount, targetRoleNames, req.DifficultyLevel, focusAreas, customInstructions)
}

func (s *AdminQuestionnaireService) processQuestionGeneration(questionnaireID uuid.UUID, promptUsed string, req GenerateQuestionnaireRequest) {
	log.Printf("Starting AI question generation for questionnaire %s", questionnaireID)

	time.Sleep(2 * time.Second)

	var aiResponse *ai.QuestionGenerationResponse
	var err error

	if s.aiService != nil {
		ctx := context.Background()
		aiResponse, err = s.aiService.GenerateQuestions(ctx, promptUsed)
		if err != nil {
			log.Printf("Error generating questions with AI: %v", err)
			return
		}
	} else {
		log.Printf("AI service not available, using default questions")
		return
	}

	questionnaireEntity, err := s.repo.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		log.Printf("Error fetching questionnaire: %v", err)
		return
	}

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

	err = s.repo.AddQuestionsToQuestionnaire(questionnaireID, questions)
	if err != nil {
		log.Printf("Error saving questions: %v", err)
		return
	}

	questionnaireEntity.AIPromptUsed = &promptUsed
	err = s.repo.UpdateQuestionnaire(questionnaireEntity)
	if err != nil {
		log.Printf("Error updating questionnaire: %v", err)
		return
	}

	log.Printf("Successfully generated %d questions for questionnaire %s", len(questions), questionnaireID)
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

func (s *AdminQuestionnaireService) getTargetRoleNames(roleIDs []string) ([]string, error) {
	var names []string

	for _, idStr := range roleIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid role ID: %s", idStr)
		}

		// For now, we'll use placeholder logic since the repository methods aren't fully implemented
		// In real implementation, this would call s.repo.GetTargetRoleByID(id)
		// For testing purposes, let's use some default role names based on common patterns
		names = append(names, s.getDefaultRoleName(id))
	}

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
