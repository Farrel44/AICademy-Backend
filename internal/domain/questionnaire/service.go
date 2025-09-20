package questionnaire

import (
	"aicademy-backend/internal/services/ai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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

	questions := make([]QuestionnaireQuestionResponse, len(questionnaire.Questions))
	for i, q := range questionnaire.Questions {
		questions[i] = QuestionnaireQuestionResponse{
			ID:            q.ID,
			QuestionText:  q.QuestionText,
			QuestionType:  q.QuestionType,
			MaxScore:      q.MaxScore,
			QuestionOrder: q.QuestionOrder,
			Category:      q.Category,
		}

		if q.QuestionType == QuestionTypeMCQ && q.Options != nil {
			var options []QuestionOption
			if err := json.Unmarshal([]byte(*q.Options), &options); err == nil {
				questions[i].Options = options
			}
		}
	}

	return &ActiveQuestionnaireResponse{
		ID:        questionnaire.ID,
		Name:      questionnaire.Name,
		Questions: questions,
	}, nil
}

func (s *QuestionnaireService) SubmitQuestionnaire(studentID uuid.UUID, request SubmitQuestionnaireRequest) (*QuestionnaireResultResponse, error) {
	questionnaire, err := s.repo.GetQuestionnaireByID(request.QuestionnaireID)
	if err != nil {
		return nil, err
	}

	if err := s.validateAnswers(questionnaire.Questions, request.Answers); err != nil {
		return nil, err
	}

	hasSubmitted, err := s.repo.HasStudentSubmitted(studentID, request.QuestionnaireID)
	if err != nil {
		return nil, err
	}
	if hasSubmitted {
		return nil, errors.New("siswa sudah mengumpulkan kuesioner ini")
	}

	answersJSON, _ := json.Marshal(request.Answers)

	response := &QuestionnaireResponse{
		StudentProfileID: studentID,
		QuestionnaireID:  request.QuestionnaireID,
		Answers:          string(answersJSON),
		SubmittedAt:      time.Now(),
	}

	err = s.repo.CreateResponse(response)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := s.processWithAI(response, questionnaire.Questions, request.Answers); err != nil {
			fmt.Printf("Error dalam pemrosesan AI: %v\n", err)
		}
	}()

	return &QuestionnaireResultResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		SubmittedAt:     response.SubmittedAt.Format(time.RFC3339),
	}, nil
}

func (s *QuestionnaireService) validateAnswers(questions []QuestionnaireQuestion, answers []AnswerItem) error {
	answerMap := make(map[uuid.UUID]AnswerItem)
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer
	}

	for _, question := range questions {
		answer, exists := answerMap[question.ID]
		if !exists {
			return fmt.Errorf("jawaban tidak ditemukan untuk pertanyaan: %s", question.QuestionText)
		}

		switch question.QuestionType {
		case QuestionTypeMCQ:
			if answer.SelectedOption == nil || *answer.SelectedOption == "" {
				return fmt.Errorf("pilihan harus dipilih untuk pertanyaan pilihan ganda: %s", question.QuestionText)
			}
		case QuestionTypeLikert:
			if answer.Score == nil || *answer.Score < 1 || *answer.Score > 5 {
				return fmt.Errorf("skor likert harus antara 1-5 untuk pertanyaan: %s", question.QuestionText)
			}
		case QuestionTypeText:
			if answer.TextAnswer == nil || strings.TrimSpace(*answer.TextAnswer) == "" {
				return fmt.Errorf("jawaban teks diperlukan untuk pertanyaan: %s", question.QuestionText)
			}
		case QuestionTypeCase:
			if answer.TextAnswer == nil || strings.TrimSpace(*answer.TextAnswer) == "" {
				return fmt.Errorf("jawaban kasus diperlukan untuk pertanyaan: %s", question.QuestionText)
			}
		}
	}

	return nil
}

func (s *QuestionnaireService) processWithAI(response *QuestionnaireResponse, questions []QuestionnaireQuestion, answers []AnswerItem) error {
	prompt := s.buildAIPrompt(questions, answers)

	ctx := context.Background()
	aiResult, err := s.aiService.GenerateCareerRecommendations(ctx, prompt)
	if err != nil {
		log.Printf("AI service failed for career recommendations: %v", err)
		return fmt.Errorf("gagal memproses rekomendasi karir: %w", err)
	}

	return s.processAIResponse(response, aiResult, questions, answers)
}

func (s *QuestionnaireService) processAIResponse(response *QuestionnaireResponse, aiResult *ai.CareerAnalysisResponse, questions []QuestionnaireQuestion, answers []AnswerItem) error {
	recommendations := make([]AIRecommendation, len(aiResult.Recommendations))
	for i, rec := range aiResult.Recommendations {
		recommendations[i] = AIRecommendation{
			RoleID:        rec.RoleID,
			RoleName:      rec.RoleName,
			Score:         rec.Score,
			Justification: rec.Justification,
		}
	}

	totalScore := s.calculateTotalScore(questions, answers)

	aiAnalysisJSON, _ := json.Marshal(aiResult.Analysis)
	aiRecommendationsJSON, _ := json.Marshal(recommendations)
	now := time.Now()

	response.AIAnalysis = stringPtr(string(aiAnalysisJSON))
	response.AIRecommendations = stringPtr(string(aiRecommendationsJSON))
	response.AIModelVersion = stringPtr("gemini-pro-v1.0")
	response.ProcessedAt = &now
	response.TotalScore = &totalScore

	if len(recommendations) > 0 {
		roleID := uuid.MustParse(recommendations[0].RoleID)
		response.RecommendedProfilingRoleID = &roleID
	}

	err := s.repo.UpdateResponse(response)
	if err != nil {
		return err
	}

	var roleRecommendations []RoleRecommendation
	for i, rec := range recommendations {
		roleRecommendations = append(roleRecommendations, RoleRecommendation{
			ResponseID:      response.ID,
			ProfilingRoleID: uuid.MustParse(rec.RoleID),
			Rank:            i + 1,
			Score:           rec.Score,
			Justification:   &rec.Justification,
			CreatedAt:       time.Now(),
		})
	}

	return s.repo.CreateRoleRecommendations(roleRecommendations)
}

func (s *QuestionnaireService) calculateTotalScore(questions []QuestionnaireQuestion, answers []AnswerItem) int {
	total := 0
	for _, answer := range answers {
		if answer.Score != nil {
			total += *answer.Score
		} else if answer.SelectedOption != nil {
			total += 3
		} else {
			total += 1
		}
	}
	return total
}

func (s *QuestionnaireService) buildAIPrompt(questions []QuestionnaireQuestion, answers []AnswerItem) string {
	var prompt strings.Builder
	prompt.WriteString("Analisis jawaban kuesioner profiling karir siswa SMK berikut:\n\n")

	answerMap := make(map[uuid.UUID]AnswerItem)
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer
	}

	for _, question := range questions {
		if answer, exists := answerMap[question.ID]; exists {
			prompt.WriteString(fmt.Sprintf("Pertanyaan: %s\n", question.QuestionText))

			switch question.QuestionType {
			case QuestionTypeMCQ:
				if answer.SelectedOption != nil {
					prompt.WriteString(fmt.Sprintf("Jawaban: %s\n", *answer.SelectedOption))
				}
			case QuestionTypeLikert:
				if answer.Score != nil {
					prompt.WriteString(fmt.Sprintf("Skor: %d/5\n", *answer.Score))
				}
			case QuestionTypeText, QuestionTypeCase:
				if answer.TextAnswer != nil {
					prompt.WriteString(fmt.Sprintf("Jawaban: %s\n", *answer.TextAnswer))
				}
			}
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString(`
Berdasarkan jawaban di atas, berikan analisis dalam format JSON:
{
  "analysis": {
    "personality_traits": ["trait1", "trait2"],
    "interests": ["interest1", "interest2"],
    "strengths": ["strength1", "strength2"],
    "work_style": "deskripsi gaya kerja"
  },
  "recommendations": [
    {
      "role_id": "uuid",
      "role_name": "nama role",
      "score": float,
      "justification": "alasan rekomendasi"
    }
  ]
}`)

	return prompt.String()
}

func (s *QuestionnaireService) GetQuestionnaireResult(responseID uuid.UUID) (*QuestionnaireResultResponse, error) {
	response, err := s.repo.GetResponseByID(responseID)
	if err != nil {
		return nil, err
	}

	result := &QuestionnaireResultResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		SubmittedAt:     response.SubmittedAt.Format(time.RFC3339),
		TotalScore:      response.TotalScore,
	}

	if response.ProcessedAt != nil {
		processedTime := response.ProcessedAt.Format(time.RFC3339)
		result.ProcessedAt = &processedTime
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

func (s *QuestionnaireService) GetLatestResultByStudent(studentID uuid.UUID) (*QuestionnaireResultResponse, error) {
	response, err := s.repo.GetLatestResponseByStudent(studentID)
	if err != nil {
		return nil, err
	}

	return s.GetQuestionnaireResult(response.ID)
}

func (s *QuestionnaireService) GenerateQuestionnaire(request GenerateQuestionnaireRequest) (*GenerationStatusResponse, error) {
	if err := s.validateGenerationRequest(request); err != nil {
		return nil, err
	}

	questionnaire := &ProfilingQuestionnaire{
		Name:        request.Name,
		GeneratedBy: "ai",
		Version:     1,
		Active:      false,
	}

	if err := s.repo.CreateQuestionnaire(questionnaire); err != nil {
		return nil, err
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

func (s *QuestionnaireService) validateGenerationRequest(request GenerateQuestionnaireRequest) error {
	if request.QuestionCount < 5 || request.QuestionCount > 50 {
		return errors.New("jumlah pertanyaan harus antara 5-50")
	}
	if len(request.TargetRoles) == 0 {
		return errors.New("minimal satu target role harus dipilih")
	}
	return nil
}

func (s *QuestionnaireService) buildQuestionGenerationPrompt(req GenerateQuestionnaireRequest) string {
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
		req.CustomInstructions,
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
		log.Println("Generating questions using AI service...")
		ctx := context.Background()
		aiResponse, err = s.aiService.GenerateQuestions(ctx, promptUsed)
		if err != nil {
			log.Printf("AI service failed: %v", err)
			s.repo.DeleteQuestionnaire(questionnaireID)
			return
		}
	} else {
		log.Println("No AI service available")
		s.repo.DeleteQuestionnaire(questionnaireID)
		return
	}

	questionnaire, err := s.repo.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		log.Printf("Error getting questionnaire: %v", err)
		return
	}

	questions := make([]QuestionnaireQuestion, len(aiResponse.Questions))
	for i, aiQ := range aiResponse.Questions {
		question := QuestionnaireQuestion{
			QuestionnaireID: questionnaireID,
			QuestionText:    aiQ.QuestionText,
			QuestionType:    QuestionType(aiQ.QuestionType),
			Category:        aiQ.Category,
			QuestionOrder:   i + 1,
			MaxScore:        s.getMaxScoreForType(QuestionType(aiQ.QuestionType)),
		}

		if aiQ.QuestionType == "mcq" && len(aiQ.Options) > 0 {
			options := make([]QuestionOption, len(aiQ.Options))
			for j, opt := range aiQ.Options {
				options[j] = QuestionOption{
					Label: opt.Label,
					Value: opt.Value,
				}
			}
			optionsJSON, _ := json.Marshal(options)
			question.Options = stringPtr(string(optionsJSON))
		}

		questions[i] = question
	}

	err = s.repo.AddQuestionsToQuestionnaire(questionnaireID, questions)
	if err != nil {
		log.Printf("Error saving questions: %v", err)
	}

	questionnaire.AIPromptUsed = &promptUsed
	err = s.repo.UpdateQuestionnaire(questionnaire)
	if err != nil {
		log.Printf("Error updating questionnaire with prompt: %v", err)
	}

	log.Printf("Successfully generated %d questions for questionnaire %s", len(questions), questionnaireID)
}

func (s *QuestionnaireService) getMaxScoreForType(questionType QuestionType) int {
	switch questionType {
	case QuestionTypeMCQ:
		return 4
	case QuestionTypeLikert:
		return 5
	case QuestionTypeText:
		return 3
	case QuestionTypeCase:
		return 3
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
		message = "Sedang membuat pertanyaan..."
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
			questions[j] = QuestionnaireQuestionResponse{
				ID:            question.ID,
				QuestionText:  question.QuestionText,
				QuestionType:  question.QuestionType,
				MaxScore:      question.MaxScore,
				QuestionOrder: question.QuestionOrder,
				Category:      question.Category,
			}

			if question.QuestionType == QuestionTypeMCQ && question.Options != nil {
				var options []QuestionOption
				if err := json.Unmarshal([]byte(*question.Options), &options); err == nil {
					questions[j].Options = options
				}
			}
		}

		responses[i] = QuestionnaireDetailResponse{
			ID:          q.ID,
			Name:        q.Name,
			Version:     q.Version,
			GeneratedBy: q.GeneratedBy,
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
		questions[i] = QuestionnaireQuestionResponse{
			ID:            q.ID,
			QuestionText:  q.QuestionText,
			QuestionType:  q.QuestionType,
			MaxScore:      q.MaxScore,
			QuestionOrder: q.QuestionOrder,
			Category:      q.Category,
		}

		if q.QuestionType == QuestionTypeMCQ && q.Options != nil {
			var options []QuestionOption
			if err := json.Unmarshal([]byte(*q.Options), &options); err == nil {
				questions[i].Options = options
			}
		}
	}

	response := &QuestionnaireDetailResponse{
		ID:          questionnaire.ID,
		Name:        questionnaire.Name,
		Version:     questionnaire.Version,
		GeneratedBy: questionnaire.GeneratedBy,
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

func (s *QuestionnaireService) CloneQuestionnaire(originalID uuid.UUID, newName string) (*QuestionnaireDetailResponse, error) {
	clone, err := s.repo.CloneQuestionnaire(originalID, newName)
	if err != nil {
		return nil, err
	}

	return s.GetQuestionnaireByID(clone.ID)
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
			SubmittedAt:     response.SubmittedAt.Format(time.RFC3339),
			TotalScore:      response.TotalScore,
		}

		if response.ProcessedAt != nil {
			processedTime := response.ProcessedAt.Format(time.RFC3339)
			result.ProcessedAt = &processedTime
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

func (s *QuestionnaireService) GetQuestionnaireStats() (map[string]interface{}, error) {
	return s.repo.GetQuestionnaireStats()
}

func (s *QuestionnaireService) GetResponseAnalytics(questionnaireID *uuid.UUID) (map[string]interface{}, error) {
	return s.repo.GetResponseAnalytics(questionnaireID)
}

func (s *QuestionnaireService) SearchQuestionnaires(keyword string, page, limit int) ([]QuestionnaireDetailResponse, int64, error) {
	questionnaires, total, err := s.repo.SearchQuestionnaires(keyword, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]QuestionnaireDetailResponse, len(questionnaires))
	for i, q := range questionnaires {
		questions := make([]QuestionnaireQuestionResponse, len(q.Questions))
		for j, question := range q.Questions {
			questions[j] = QuestionnaireQuestionResponse{
				ID:            question.ID,
				QuestionText:  question.QuestionText,
				QuestionType:  question.QuestionType,
				MaxScore:      question.MaxScore,
				QuestionOrder: question.QuestionOrder,
				Category:      question.Category,
			}

			if question.QuestionType == QuestionTypeMCQ && question.Options != nil {
				var options []QuestionOption
				if err := json.Unmarshal([]byte(*question.Options), &options); err == nil {
					questions[j].Options = options
				}
			}
		}

		responses[i] = QuestionnaireDetailResponse{
			ID:          q.ID,
			Name:        q.Name,
			Version:     q.Version,
			GeneratedBy: q.GeneratedBy,
			Questions:   questions,
		}
	}

	return responses, total, nil
}

func (s *QuestionnaireService) GetQuestionnairesByGeneratedBy(generatedBy string, page, limit int) ([]QuestionnaireDetailResponse, int64, error) {
	questionnaires, total, err := s.repo.GetQuestionnairesByGeneratedBy(generatedBy, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]QuestionnaireDetailResponse, len(questionnaires))
	for i, q := range questionnaires {
		questions := make([]QuestionnaireQuestionResponse, len(q.Questions))
		for j, question := range q.Questions {
			questions[j] = QuestionnaireQuestionResponse{
				ID:            question.ID,
				QuestionText:  question.QuestionText,
				QuestionType:  question.QuestionType,
				MaxScore:      question.MaxScore,
				QuestionOrder: question.QuestionOrder,
				Category:      question.Category,
			}

			if question.QuestionType == QuestionTypeMCQ && question.Options != nil {
				var options []QuestionOption
				if err := json.Unmarshal([]byte(*question.Options), &options); err == nil {
					questions[j].Options = options
				}
			}
		}

		responses[i] = QuestionnaireDetailResponse{
			ID:          q.ID,
			Name:        q.Name,
			Version:     q.Version,
			GeneratedBy: q.GeneratedBy,
			Questions:   questions,
		}
	}

	return responses, total, nil
}

func (s *QuestionnaireService) AddQuestionToQuestionnaire(questionnaireID uuid.UUID, questionText string, questionType QuestionType, options []QuestionOption, category string, questionOrder int) error {
	question := QuestionnaireQuestion{
		QuestionnaireID: questionnaireID,
		QuestionText:    questionText,
		QuestionType:    questionType,
		Category:        category,
		QuestionOrder:   questionOrder,
		MaxScore:        s.getMaxScoreForType(questionType),
	}

	if questionType == QuestionTypeMCQ && len(options) > 0 {
		optionsJSON, err := json.Marshal(options)
		if err != nil {
			return err
		}
		question.Options = stringPtr(string(optionsJSON))
	}

	return s.repo.AddQuestionsToQuestionnaire(questionnaireID, []QuestionnaireQuestion{question})
}

func (s *QuestionnaireService) UpdateQuestion(questionID uuid.UUID, questionText string, questionType QuestionType, options []QuestionOption, category string, questionOrder int) error {
	question := QuestionnaireQuestion{
		ID:            questionID,
		QuestionText:  questionText,
		QuestionType:  questionType,
		Category:      category,
		QuestionOrder: questionOrder,
		MaxScore:      s.getMaxScoreForType(questionType),
	}

	if questionType == QuestionTypeMCQ && len(options) > 0 {
		optionsJSON, err := json.Marshal(options)
		if err != nil {
			return err
		}
		question.Options = stringPtr(string(optionsJSON))
	}

	return s.repo.UpdateQuestion(&question)
}

func (s *QuestionnaireService) DeleteQuestion(questionID uuid.UUID) error {
	return s.repo.DeleteQuestion(questionID)
}

func (s *QuestionnaireService) UpdateQuestionOrder(questionnaireID uuid.UUID, questions []struct {
	ID    uuid.UUID `json:"id" validate:"required"`
	Order int       `json:"order" validate:"required,min=1"`
}) error {
	updates := make([]struct {
		ID    uuid.UUID
		Order int
	}, len(questions))

	for i, q := range questions {
		updates[i] = struct {
			ID    uuid.UUID
			Order int
		}{
			ID:    q.ID,
			Order: q.Order,
		}
	}

	return s.repo.BulkUpdateQuestionOrder(updates)
}

func (s *QuestionnaireService) CreateQuestionTemplate(name, description, prompt string) (*QuestionGenerationTemplate, error) {
	template := &QuestionGenerationTemplate{
		Name:   name,
		Prompt: prompt,
		Active: true,
	}

	if description != "" {
		template.Description = &description
	}

	err := s.repo.CreateQuestionGenerationTemplate(template)
	if err != nil {
		return nil, err
	}

	return template, nil
}

func (s *QuestionnaireService) GetQuestionTemplates() ([]QuestionGenerationTemplate, error) {
	return s.repo.GetQuestionGenerationTemplates()
}

func stringPtr(s string) *string {
	return &s
}
