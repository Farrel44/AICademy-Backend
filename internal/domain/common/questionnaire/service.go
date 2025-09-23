package questionnaire

import (
	questionnaireRepo "aicademy-backend/internal/domain/questionnaire"
	"aicademy-backend/internal/services/ai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CommonQuestionnaireService struct {
	repo      *questionnaireRepo.QuestionnaireRepository
	aiService ai.AIService
}

func NewCommonQuestionnaireService(repo *questionnaireRepo.QuestionnaireRepository, aiService ai.AIService) *CommonQuestionnaireService {
	return &CommonQuestionnaireService{
		repo:      repo,
		aiService: aiService,
	}
}

func (s *CommonQuestionnaireService) GetActiveQuestionnaire() (*ActiveQuestionnaireResponse, error) {
	questionnaire, err := s.repo.GetActiveQuestionnaire()
	if err != nil {
		return nil, errors.New("no active questionnaire found")
	}

	questions, err := s.repo.GetQuestionsByQuestionnaireID(questionnaire.ID)
	if err != nil {
		return nil, errors.New("failed to get questionnaire questions")
	}

	var questionDTOs []QuestionnaireQuestionDTO
	for _, q := range questions {
		questionDTO := QuestionnaireQuestionDTO{
			ID:           q.ID,
			QuestionText: q.QuestionText,
			QuestionType: string(q.QuestionType),
			Category:     q.Category,
			Order:        q.QuestionOrder,
		}

		// Parse options if available
		if q.Options != nil && *q.Options != "" {
			var options []OptionDTO
			if err := json.Unmarshal([]byte(*q.Options), &options); err == nil {
				questionDTO.Options = options
			}
		}

		questionDTOs = append(questionDTOs, questionDTO)
	}

	return &ActiveQuestionnaireResponse{
		ID:          questionnaire.ID,
		Name:        questionnaire.Name,
		Version:     questionnaire.Version,
		Questions:   questionDTOs,
		Instruction: "Jawab semua pertanyaan dengan jujur untuk mendapatkan rekomendasi karir yang sesuai dengan kepribadian dan minat Anda.",
	}, nil
}

// SubmitQuestionnaire - Students only
func (s *CommonQuestionnaireService) SubmitQuestionnaire(userID uuid.UUID, req SubmitQuestionnaireRequest) (*QuestionnaireResponse, error) {
	// Verify questionnaire exists and is active
	questionnaire, err := s.repo.GetQuestionnaireByID(req.QuestionnaireID)
	if err != nil || !questionnaire.Active {
		return nil, errors.New("questionnaire not found or not active")
	}

	// Get student profile
	studentProfile, err := s.repo.GetStudentProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	// Check if student already submitted this questionnaire
	existingResponse, _ := s.repo.GetResponseByStudentAndQuestionnaire(studentProfile.ID, req.QuestionnaireID)
	if existingResponse != nil {
		return nil, errors.New("questionnaire already submitted")
	}

	// Calculate total score
	questions, err := s.repo.GetQuestionsByQuestionnaireID(req.QuestionnaireID)
	if err != nil {
		return nil, errors.New("failed to get questionnaire questions")
	}

	totalScore, _ := s.calculateScore(questions, req.Answers)

	// Create response using the questionnaire model
	responseID := uuid.New()
	answersJSON := convertAnswersToJSON(req.Answers)

	// Create response record using questionnaire model
	response := &questionnaireRepo.QuestionnaireResponse{
		ID:               responseID,
		QuestionnaireID:  req.QuestionnaireID,
		StudentProfileID: studentProfile.ID,
		Answers:          answersJSON,
		SubmittedAt:      time.Now(),
		TotalScore:       &totalScore,
	}

	err = s.repo.CreateQuestionnaireResponse(response)
	if err != nil {
		return nil, errors.New("failed to save questionnaire response")
	}

	// Process with AI for career recommendations
	responseDTO := &QuestionnaireResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		StudentID:       userID,
		SubmittedAt:     response.SubmittedAt,
		TotalScore:      totalScore,
		MaxScore:        100, // Default max score
		Status:          "processing",
	}

	// Process AI analysis in background
	go s.processWithAI(response, questions, req.Answers)

	return responseDTO, nil
}

func (s *CommonQuestionnaireService) calculateScore(questions []questionnaireRepo.QuestionnaireQuestion, answers []AnswerItem) (int, int) {
	totalScore := 0
	maxScore := 0

	answerMap := make(map[uuid.UUID]AnswerItem)
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer
	}

	for _, question := range questions {
		maxScore += question.MaxScore
		if answer, exists := answerMap[question.ID]; exists {
			totalScore += answer.Score
		}
	}

	return totalScore, maxScore
}

func (s *CommonQuestionnaireService) processWithAI(response *questionnaireRepo.QuestionnaireResponse, questions []questionnaireRepo.QuestionnaireQuestion, answers []AnswerItem) error {
	prompt := s.buildAIPrompt(questions, answers)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aiResult, err := s.aiService.GenerateCareerRecommendations(ctx, prompt)
	if err != nil {
		// Update response with error - use UpdateResponse
		response.ProcessedAt = &time.Time{}
		now := time.Now()
		response.ProcessedAt = &now
		s.repo.UpdateResponse(response)
		return err
	}

	return s.processAIResponse(response, aiResult, questions, answers)
}

func (s *CommonQuestionnaireService) buildAIPrompt(questions []questionnaireRepo.QuestionnaireQuestion, answers []AnswerItem) string {
	prompt := "Berdasarkan jawaban kuesioner profiling karir berikut, berikan rekomendasi karir yang sesuai:\n\n"

	answerMap := make(map[uuid.UUID]AnswerItem)
	for _, answer := range answers {
		answerMap[answer.QuestionID] = answer
	}

	for _, question := range questions {
		if answer, exists := answerMap[question.ID]; exists {
			prompt += fmt.Sprintf("Q: %s\nA: %s\n\n", question.QuestionText, answer.Answer)
		}
	}

	prompt += `
Berikan rekomendasi dalam format JSON berikut:
{
  "analysis": {
    "personality_traits": ["trait1", "trait2"],
    "interests": ["interest1", "interest2"],
    "strengths": ["strength1", "strength2"],
    "work_style": "description"
  },
  "recommendations": [
    {
      "role_id": "generate-uuid",
      "role_name": "nama_role",
      "score": 0.9,
      "justification": "alasan rekomendasi"
    }
  ]
}`

	return prompt
}

func (s *CommonQuestionnaireService) processAIResponse(response *questionnaireRepo.QuestionnaireResponse, aiResult *ai.CareerAnalysisResponse, questions []questionnaireRepo.QuestionnaireQuestion, answers []AnswerItem) error {
	// Save AI analysis result
	analysisJSON, _ := json.Marshal(aiResult.Analysis)
	analysisStr := string(analysisJSON)
	response.AIAnalysis = &analysisStr

	// Update response with AI results
	now := time.Now()
	response.ProcessedAt = &now

	err := s.repo.UpdateResponse(response)
	if err != nil {
		return err
	}

	// Note: For recommendations, we'd need to create a separate table or store in AIRecommendations field
	// For now, let's store in AIRecommendations field as JSON
	recJSON, _ := json.Marshal(aiResult.Recommendations)
	recStr := string(recJSON)
	response.AIRecommendations = &recStr
	s.repo.UpdateResponse(response)

	return nil
}

// GetQuestionnaireResult - Students can get their own results
func (s *CommonQuestionnaireService) GetQuestionnaireResult(userID uuid.UUID, responseID uuid.UUID) (*QuestionnaireResponse, error) {
	response, err := s.repo.GetResponseByID(responseID)
	if err != nil {
		return nil, errors.New("questionnaire response not found")
	}

	// Verify ownership
	studentProfile, err := s.repo.GetStudentProfileByUserID(userID)
	if err != nil || response.StudentProfileID != studentProfile.ID {
		return nil, errors.New("unauthorized access to questionnaire result")
	}

	// Parse recommendations from AIRecommendations field
	var recDTOs []CareerRecommendationDTO
	if response.AIRecommendations != nil {
		var aiRecs []struct {
			RoleName      string  `json:"role_name"`
			Score         float64 `json:"score"`
			Justification string  `json:"justification"`
		}
		json.Unmarshal([]byte(*response.AIRecommendations), &aiRecs)

		for _, rec := range aiRecs {
			recDTOs = append(recDTOs, CareerRecommendationDTO{
				RoleID:        uuid.New(), // Generate temp ID
				RoleName:      rec.RoleName,
				Score:         rec.Score,
				Justification: rec.Justification,
				Category:      "AI Generated",
			})
		}
	}

	totalScore := 0
	if response.TotalScore != nil {
		totalScore = *response.TotalScore
	}

	return &QuestionnaireResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		StudentID:       userID,
		SubmittedAt:     response.SubmittedAt,
		TotalScore:      totalScore,
		MaxScore:        100, // Default, should be calculated from questions
		Recommendations: recDTOs,
		Status:          "completed",
	}, nil
}

// GetLatestResultByStudent - Students can get their latest result
func (s *CommonQuestionnaireService) GetLatestResultByStudent(userID uuid.UUID) (*QuestionnaireResponse, error) {
	studentProfile, err := s.repo.GetStudentProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("student profile not found")
	}

	response, err := s.repo.GetLatestResponseByStudentProfile(studentProfile.ID)
	if err != nil {
		return nil, errors.New("no questionnaire results found")
	}

	// Parse recommendations from AIRecommendations field
	var recDTOs []CareerRecommendationDTO
	if response.AIRecommendations != nil {
		var aiRecs []struct {
			RoleName      string  `json:"role_name"`
			Score         float64 `json:"score"`
			Justification string  `json:"justification"`
		}
		json.Unmarshal([]byte(*response.AIRecommendations), &aiRecs)

		for _, rec := range aiRecs {
			recDTOs = append(recDTOs, CareerRecommendationDTO{
				RoleID:        uuid.New(), // Generate temp ID
				RoleName:      rec.RoleName,
				Score:         rec.Score,
				Justification: rec.Justification,
				Category:      "AI Generated",
			})
		}
	}

	totalScore := 0
	if response.TotalScore != nil {
		totalScore = *response.TotalScore
	}

	return &QuestionnaireResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		StudentID:       userID,
		SubmittedAt:     response.SubmittedAt,
		TotalScore:      totalScore,
		MaxScore:        100, // Default
		Recommendations: recDTOs,
		Status:          "completed",
	}, nil
}

// Helper functions
func convertAnswersToJSON(answers []AnswerItem) string {
	data, _ := json.Marshal(answers)
	return string(data)
}
