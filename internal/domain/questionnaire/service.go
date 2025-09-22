package questionnaire

import (
	"aicademy-backend/internal/domain/user"
	"aicademy-backend/internal/services/ai"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type QuestionnaireService struct {
	repo      *QuestionnaireRepository
	aiService ai.AIService
}

func NewQuestionnaireService(repo *QuestionnaireRepository, aiService ai.AIService) *QuestionnaireService {
	return &QuestionnaireService{
		repo:      repo,
		aiService: aiService,
	}
}

func (s *QuestionnaireService) GetActiveQuestionnaire() (*ActiveQuestionnaireResponse, error) {
	questionnaire, err := s.repo.GetActiveQuestionnaire()
	if err != nil {
		return nil, err
	}

	questions := make([]QuestionnaireQuestionDTO, len(questionnaire.Questions))
	for i, q := range questionnaire.Questions {
		var options []string
		if q.Options != nil {
			var optionObjects []QuestionOption
			if json.Unmarshal([]byte(*q.Options), &optionObjects) == nil {
				for _, opt := range optionObjects {
					options = append(options, opt.Text)
				}
			}
		}

		questions[i] = QuestionnaireQuestionDTO{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			QuestionType: q.QuestionType,
			Options:      options,
			MaxScore:     q.MaxScore,
			Category:     q.Category,
			Order:        q.QuestionOrder,
		}
	}

	return &ActiveQuestionnaireResponse{
		ID:        questionnaire.ID,
		Name:      questionnaire.Name,
		Version:   questionnaire.Version,
		Questions: questions,
	}, nil
}

func (s *QuestionnaireService) SubmitQuestionnaire(userID uuid.UUID, req SubmitQuestionnaireRequest) (*QuestionnaireResponse, error) {
	log.Printf("Processing questionnaire submission for user: %s", userID.String())

	studentProfile, err := s.repo.GetStudentProfileByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("student profile tidak ditemukan: %w", err)
	}

	questionnaire, err := s.repo.GetQuestionnaireByID(req.QuestionnaireID)
	if err != nil {
		return nil, fmt.Errorf("kuesioner tidak ditemukan: %w", err)
	}

	questionMap := make(map[uuid.UUID]*QuestionnaireQuestion)
	for i := range questionnaire.Questions {
		questionMap[questionnaire.Questions[i].ID] = &questionnaire.Questions[i]
	}

	for _, answer := range req.Answers {
		if _, exists := questionMap[answer.QuestionID]; !exists {
			return nil, fmt.Errorf("pertanyaan dengan ID %s tidak ditemukan", answer.QuestionID.String())
		}
	}

	existingResponse, err := s.repo.GetResponseByStudentAndQuestionnaire(studentProfile.ID, req.QuestionnaireID)
	if err == nil && existingResponse != nil {
		return nil, fmt.Errorf("kuesioner sudah pernah dijawab")
	}

	totalScore := 0
	for _, answer := range req.Answers {
		if answer.Score != nil {
			totalScore += *answer.Score
		}
	}

	response := &QuestionnaireResponse{
		StudentProfileID: studentProfile.ID,
		QuestionnaireID:  req.QuestionnaireID,
		Answers:          convertAnswersToJSON(req.Answers),
		TotalScore:       &totalScore,
		SubmittedAt:      time.Now(),
	}

	err = s.repo.CreateQuestionnaireResponse(response)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan respons: %w", err)
	}

	log.Printf("Response saved with ID: %s", response.ID.String())

	go s.processWithAI(response, questionnaire.Questions, req.Answers)

	return response, nil
}

func convertAnswersToJSON(answers []AnswerItem) string {
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		log.Printf("Error marshaling answers: %v", err)
		return "[]"
	}
	return string(answersJSON)
}

func (s *QuestionnaireService) processWithAI(response *QuestionnaireResponse, questions []QuestionnaireQuestion, answers []AnswerItem) error {
	prompt := s.buildAIPrompt(questions, answers)

	ctx := context.Background()
	aiResult, err := s.aiService.GenerateCareerRecommendations(ctx, prompt)
	if err != nil {
		log.Printf("Error generating AI recommendations: %v", err)
		return err
	}

	return s.processAIResponse(response, aiResult, questions, answers)
}

func (s *QuestionnaireService) processAIResponse(response *QuestionnaireResponse, aiResult *ai.CareerAnalysisResponse, questions []QuestionnaireQuestion, answers []AnswerItem) error {
	log.Printf("Processing AI response with %d recommendations", len(aiResult.Recommendations))

	now := time.Now()
	response.ProcessedAt = &now

	analysisJSON, _ := json.Marshal(aiResult.Analysis)
	analysisStr := string(analysisJSON)
	response.AIAnalysis = &analysisStr

	aiRecommendations := make([]AIRecommendation, 0)

	for i, rec := range aiResult.Recommendations {
		var roleID string
		if _, err := uuid.Parse(rec.RoleID); err != nil {
			roleID = generateRoleUUID(rec.RoleName).String()
			log.Printf("Generated UUID %s for role: %s (original: %s)", roleID, rec.RoleName, rec.RoleID)
		} else {
			roleID = rec.RoleID
		}

		aiRec := AIRecommendation{
			RoleID:        roleID,
			RoleName:      rec.RoleName,
			Score:         rec.Score,
			Justification: rec.Justification,
		}
		aiRecommendations = append(aiRecommendations, aiRec)

		log.Printf("Recommendation %d: %s (Score: %.1f)", i+1, rec.RoleName, rec.Score)
	}

	if len(aiRecommendations) > 0 {
		topRecommendation := aiRecommendations[0]

		if roleUUID, err := uuid.Parse(topRecommendation.RoleID); err == nil {
			response.RecommendedProfilingRoleID = &roleUUID
		} else {
			log.Printf("Warning: Failed to parse role ID as UUID: %s", topRecommendation.RoleID)
			response.RecommendedProfilingRoleID = nil
		}

		log.Printf("Top recommendation: %s", topRecommendation.RoleName)
	}

	recommendationsJSON, err := json.Marshal(aiRecommendations)
	if err != nil {
		log.Printf("Error marshaling AI recommendations: %v", err)
		return fmt.Errorf("gagal convert rekomendasi AI: %w", err)
	}
	recommendationsStr := string(recommendationsJSON)
	response.AIRecommendations = &recommendationsStr

	response.AIModelVersion = stringPtr("gemini-1.5-flash-v1")

	err = s.repo.UpdateResponse(response)
	if err != nil {
		log.Printf("Error updating response: %v", err)
		return fmt.Errorf("gagal update response: %w", err)
	}

	log.Printf("AI processing completed successfully for response %s", response.ID.String())
	return nil
}

func generateRoleUUID(roleName string) uuid.UUID {
	roleMap := map[string]string{
		"Backend Developer":      "a1b2c3d4-e5f6-7890-1234-567890abcdef",
		"Frontend Developer":     "b2c3d4e5-f6a7-8901-2345-678901bcdef0",
		"Full Stack Developer":   "c3d4e5f6-a7b8-9012-3456-789012cdef01",
		"Database Administrator": "d4e5f6a7-b8c9-0123-4567-890123def012",
		"API Developer":          "e5f6a7b8-c9d0-1234-5678-901234ef0123",
		"Mobile Developer":       "f6a7b8c9-d0e1-2345-6789-012345f01234",
		"DevOps Engineer":        "a7b8c9d0-e1f2-3456-7890-123456012345",
		"Data Scientist":         "b8c9d0e1-f2a3-4567-8901-234567123456",
		"UI/UX Designer":         "c9d0e1f2-a3b4-5678-9012-345678234567",
		"System Administrator":   "d0e1f2a3-b4c5-6789-0123-456789345678",
	}

	if existingUUID, exists := roleMap[roleName]; exists {
		parsed, _ := uuid.Parse(existingUUID)
		return parsed
	}

	return uuid.New()
}

func stringPtr(s string) *string {
	return &s
}

func (s *QuestionnaireService) buildAIPrompt(questions []QuestionnaireQuestion, answers []AnswerItem) string {
	prompt := `
Analisis jawaban kuesioner profiling karir berikut:

JAWABAN KUESIONER:
`

	for i, answer := range answers {
		var question *QuestionnaireQuestion
		for j := range questions {
			if questions[j].ID == answer.QuestionID {
				question = &questions[j]
				break
			}
		}

		if question != nil {
			prompt += fmt.Sprintf("\n%d. %s", i+1, question.QuestionText)
			if answer.SelectedOption != nil {
				prompt += fmt.Sprintf("\n   Jawaban: %s", *answer.SelectedOption)
			}
			if answer.Score != nil {
				prompt += fmt.Sprintf("\n   Skor: %d/%d", *answer.Score, question.MaxScore)
			}
			if answer.TextAnswer != nil {
				prompt += fmt.Sprintf("\n   Jawaban teks: %s", *answer.TextAnswer)
			}
		}
	}

	prompt += `

INSTRUKSI ANALISIS:
Berdasarkan jawaban di atas, berikan rekomendasi karir dalam format JSON berikut:

{
  "analysis": {
    "personality_traits": ["analitis", "detail-oriented", "problem-solver"],
    "interests": ["backend development", "database management"],
    "strengths": ["logical thinking", "technical skills"],
    "work_style": "Lebih suka bekerja dengan data dan logika"
  },
  "recommendations": [
    {
      "role_id": "backend-dev",
      "role_name": "Backend Developer",
      "score": 87.5,
      "justification": "Berdasarkan preferensi backend dan kemampuan problem solving"
    },
    {
      "role_id": "api-dev", 
      "role_name": "API Developer",
      "score": 82.0,
      "justification": "Pengalaman dengan API dan minat pada backend"
    }
  ]
}

PENTING:
- role_id harus berupa string pendek (bukan UUID)
- Berikan maksimal 3 rekomendasi
- Urutkan berdasarkan score tertinggi
- Score dalam range 0-100
- Justification harus spesifik berdasarkan jawaban

Berikan HANYA JSON, tanpa text tambahan.`

	return prompt
}

func (s *QuestionnaireService) GetQuestionnaireResult(responseID uuid.UUID) (*QuestionnaireResultResponse, error) {
	response, err := s.repo.GetResponseByID(responseID)
	if err != nil {
		return nil, err
	}

	result := &QuestionnaireResultResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		SubmittedAt:     response.SubmittedAt,
		TotalScore:      response.TotalScore,
	}

	if response.ProcessedAt != nil {
		result.ProcessedAt = response.ProcessedAt
	}

	if response.AIRecommendations != nil {
		var recommendations []AIRecommendation
		if err := json.Unmarshal([]byte(*response.AIRecommendations), &recommendations); err == nil {
			result.AIRecommendations = recommendations

			if len(recommendations) > 0 {
				result.RecommendedRole = &recommendations[0].RoleName
			}
		}
	}

	return result, nil
}

func (s *QuestionnaireService) GetLatestResultByStudentProfile(studentProfileID uuid.UUID) (*QuestionnaireResultResponse, error) {
	log.Printf("Getting latest response for student profile: %s", studentProfileID.String())

	response, err := s.repo.GetLatestResponseByStudentProfile(studentProfileID)
	if err != nil {
		log.Printf("Error in GetLatestResponseByStudentProfile: %v", err)
		return nil, err
	}

	log.Printf("Found response: %s", response.ID.String())
	return s.GetQuestionnaireResult(response.ID)
}

func (s *QuestionnaireService) GetStudentProfileByUserID(userID uuid.UUID) (*user.StudentProfile, error) {
	return s.repo.GetStudentProfileByUserID(userID)
}

func (s *QuestionnaireService) GenerateQuestionnaire(request GenerateQuestionnaireRequest) (*GenerationStatusResponse, error) {
	questionnaire := &ProfilingQuestionnaire{
		Name:        request.Name,
		GeneratedBy: "ai",
		Version:     1,
		Active:      false,
	}

	err := s.repo.CreateQuestionnaire(questionnaire)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat kuesioner: %w", err)
	}

	prompt := s.buildQuestionGenerationPrompt(request)

	go s.processQuestionGeneration(questionnaire.ID, prompt, request)

	return &GenerationStatusResponse{
		QuestionnaireID: questionnaire.ID,
		Status:          "memproses",
		Progress:        0,
		Message:         "Generasi kuesioner dimulai",
	}, nil
}

func (s *QuestionnaireService) buildQuestionGenerationPrompt(req GenerateQuestionnaireRequest) string {
	customInstructions := ""
	if req.CustomInstructions != nil {
		customInstructions = *req.CustomInstructions
	}

	prompt := fmt.Sprintf(`Anda adalah seorang ahli psikologi karir dan teknologi yang akan membuat kuesioner profiling karir untuk siswa SMK.

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
      "question_text": "Pertanyaan lengkap",
      "question_type": "mcq|likert|case|text",
      "options": [{"label": "Pilihan A", "value": "a"}],
      "category": "personality|skills|interests|preferences|experience", 
      "reasoning": "Penjelasan kenapa pertanyaan ini efektif untuk membedakan role"
    }
  ],
  "metadata": {
    "total_questions": %d,
    "distribution": {"mcq": 5, "likert": 4, "case": 2, "text": 1},
    "target_roles_coverage": ["Frontend Developer", "Backend Developer"]
  }
}`,
		req.QuestionCount,
		req.TargetRoles,
		req.DifficultyLevel,
		req.FocusAreas,
		customInstructions,
		req.QuestionCount,
	)

	return prompt
}

func (s *QuestionnaireService) processQuestionGeneration(questionnaireID uuid.UUID, promptUsed string, req GenerateQuestionnaireRequest) {
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

	questionnaire, err := s.repo.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		log.Printf("Error fetching questionnaire: %v", err)
		return
	}

	questions := make([]QuestionnaireQuestion, len(aiResponse.Questions))
	for i, aiQ := range aiResponse.Questions {
		var optionsJSON *string
		if len(aiQ.Options) > 0 {
			optionsBytes, _ := json.Marshal(aiQ.Options)
			optionsStr := string(optionsBytes)
			optionsJSON = &optionsStr
		}

		questionType := QuestionType(aiQ.QuestionType)
		maxScore := s.getMaxScoreForType(questionType)

		questions[i] = QuestionnaireQuestion{
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

	questionnaire.AIPromptUsed = &promptUsed
	err = s.repo.UpdateQuestionnaire(questionnaire)
	if err != nil {
		log.Printf("Error updating questionnaire: %v", err)
		return
	}

	log.Printf("Successfully generated %d questions for questionnaire %s", len(questions), questionnaireID)
}

func (s *QuestionnaireService) getMaxScoreForType(questionType QuestionType) int {
	switch questionType {
	case QuestionTypeLikert:
		return 5
	case QuestionTypeMCQ:
		return 1
	case QuestionTypeCase:
		return 5
	case QuestionTypeText:
		return 0
	default:
		return 1
	}
}

func (s *QuestionnaireService) GetGenerationStatus(questionnaireID uuid.UUID) (*GenerationStatusResponse, error) {
	questionnaire, err := s.repo.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		return nil, err
	}

	questions, _ := s.repo.GetQuestionsByQuestionnaireID(questionnaireID)

	status := "selesai"
	progress := 100
	message := "Kuesioner berhasil dibuat"

	if len(questions) == 0 {
		status = "memproses"
		progress = 50
		message = "Sedang menghasilkan pertanyaan"
	}

	return &GenerationStatusResponse{
		QuestionnaireID: questionnaire.ID,
		Status:          status,
		Progress:        progress,
		Message:         message,
	}, nil
}

func (s *QuestionnaireService) GetAllQuestionnaires(page, limit int) ([]QuestionnaireDetailResponse, int64, error) {
	questionnaires, total, err := s.repo.GetAllQuestionnaires(page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]QuestionnaireDetailResponse, len(questionnaires))
	for i, q := range questionnaires {
		questions := make([]QuestionnaireQuestionResponse, len(q.Questions))
		for j, question := range q.Questions {
			var options []QuestionOption
			if question.Options != nil {
				json.Unmarshal([]byte(*question.Options), &options)
			}

			questions[j] = QuestionnaireQuestionResponse{
				ID:            question.ID,
				QuestionText:  question.QuestionText,
				QuestionType:  string(question.QuestionType),
				MaxScore:      question.MaxScore,
				QuestionOrder: question.QuestionOrder,
				Category:      question.Category,
				Options:       options,
			}
		}

		responses[i] = QuestionnaireDetailResponse{
			ID:          q.ID,
			Name:        q.Name,
			Version:     q.Version,
			GeneratedBy: q.GeneratedBy,
			Active:      q.Active,
			Questions:   questions,
		}
	}

	return responses, total, nil
}

func (s *QuestionnaireService) GetQuestionnaireByID(id uuid.UUID) (*QuestionnaireDetailResponse, error) {
	questionnaire, err := s.repo.GetQuestionnaireByID(id)
	if err != nil {
		return nil, err
	}

	questions := make([]QuestionnaireQuestionResponse, len(questionnaire.Questions))
	for i, q := range questionnaire.Questions {
		var options []QuestionOption
		if q.Options != nil {
			json.Unmarshal([]byte(*q.Options), &options)
		}

		questions[i] = QuestionnaireQuestionResponse{
			ID:            q.ID,
			QuestionText:  q.QuestionText,
			QuestionType:  string(q.QuestionType),
			MaxScore:      q.MaxScore,
			QuestionOrder: q.QuestionOrder,
			Category:      q.Category,
			Options:       options,
		}
	}

	response := &QuestionnaireDetailResponse{
		ID:          questionnaire.ID,
		Name:        questionnaire.Name,
		Version:     questionnaire.Version,
		GeneratedBy: questionnaire.GeneratedBy,
		Active:      questionnaire.Active,
		Questions:   questions,
	}

	return response, nil
}

func (s *QuestionnaireService) ActivateQuestionnaire(id uuid.UUID) error {
	return s.repo.ActivateQuestionnaire(id)
}

func (s *QuestionnaireService) DeactivateAllQuestionnaires() error {
	return s.repo.DeactivateAllQuestionnaires()
}

func (s *QuestionnaireService) DeleteQuestionnaire(id uuid.UUID) error {
	return s.repo.DeleteQuestionnaire(id)
}

func (s *QuestionnaireService) GetQuestionnaireResponses(questionnaireID uuid.UUID, page, limit int) ([]QuestionnaireResultResponse, int64, error) {
	responses, total, err := s.repo.GetResponsesByQuestionnaireID(questionnaireID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	results := make([]QuestionnaireResultResponse, len(responses))
	for i, response := range responses {
		result := QuestionnaireResultResponse{
			ID:              response.ID,
			QuestionnaireID: response.QuestionnaireID,
			SubmittedAt:     response.SubmittedAt,
			TotalScore:      response.TotalScore,
		}

		if response.ProcessedAt != nil {
			result.ProcessedAt = response.ProcessedAt
		}

		if response.AIRecommendations != nil {
			var recommendations []AIRecommendation
			if err := json.Unmarshal([]byte(*response.AIRecommendations), &recommendations); err == nil {
				result.AIRecommendations = recommendations

				if len(recommendations) > 0 {
					result.RecommendedRole = &recommendations[0].RoleName
				}
			}
		}

		results[i] = result
	}

	return results, total, nil
}

func (s *QuestionnaireService) CreateRole(roleName, description, category string) (*RoleRecommendation, error) {
	role := &RoleRecommendation{
		RoleName:    roleName,
		Description: description,
		Category:    category,
		Active:      true,
	}

	err := s.repo.CreateRole(role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (s *QuestionnaireService) GetAllRoles() ([]RoleRecommendation, error) {
	return s.repo.GetAllRoles()
}

func (s *QuestionnaireService) DeleteRole(id uuid.UUID) error {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return err
	}

	return s.repo.DeleteRole(role.ID)
}

func (s *QuestionnaireService) GetRoleByName(roleName string) (*RoleRecommendation, error) {
	return s.repo.GetRoleByName(roleName)
}

func (s *QuestionnaireService) UpdateRole(id uuid.UUID, roleName, description, category string) (*RoleRecommendation, error) {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return nil, err
	}

	role.RoleName = roleName
	role.Description = description
	role.Category = category

	err = s.repo.UpdateRole(role)
	if err != nil {
		return nil, err
	}

	return role, nil
}
