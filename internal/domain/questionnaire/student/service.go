package student

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	questionnaireRepo "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/services/ai"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentQuestionnaireService struct {
	repo      *questionnaireRepo.QuestionnaireRepository
	aiService ai.AIService
}

func NewStudentQuestionnaireService(repo *questionnaireRepo.QuestionnaireRepository, aiService ai.AIService) *StudentQuestionnaireService {
	return &StudentQuestionnaireService{
		repo:      repo,
		aiService: aiService,
	}
}

func (s *StudentQuestionnaireService) GetActiveQuestionnaire() (*ActiveQuestionnaireResponse, error) {
	questionnaire, err := s.repo.GetActiveQuestionnaire()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("no active questionnaire found")
		}
		return nil, errors.New("failed to get active questionnaire")
	}

	questions, err := s.repo.GetQuestionsByQuestionnaireID(questionnaire.ID)
	if err != nil {
		return nil, errors.New("failed to get questionnaire questions")
	}

	response := &ActiveQuestionnaireResponse{
		ID:          questionnaire.ID,
		Name:        questionnaire.Name,
		Version:     questionnaire.Version,
		Instruction: "Please answer all questions honestly to get the best career recommendation.",
		Questions:   make([]QuestionnaireQuestionDTO, len(questions)),
	}

	for i, question := range questions {
		questionDTO := QuestionnaireQuestionDTO{
			ID:           question.ID,
			QuestionText: question.QuestionText,
			QuestionType: string(question.QuestionType),
			Category:     question.Category,
			Order:        question.QuestionOrder,
		}

		if question.Options != nil {
			var options []map[string]interface{}
			if err := json.Unmarshal([]byte(*question.Options), &options); err == nil {
				questionDTO.Options = make([]OptionDTO, len(options))
				for j, option := range options {
					questionDTO.Options[j] = OptionDTO{
						Value: fmt.Sprintf("%v", option["value"]),
						Label: fmt.Sprintf("%v", option["text"]),
						Score: int(option["score"].(float64)),
					}
				}
			}
		}

		response.Questions[i] = questionDTO
	}

	return response, nil
}

func (s *StudentQuestionnaireService) SubmitQuestionnaire(studentProfileID uuid.UUID, req SubmitQuestionnaireRequest) (*QuestionnaireSubmissionResponse, error) {
	questionnaire, err := s.repo.GetQuestionnaireByID(req.QuestionnaireID)
	if err != nil {
		return nil, errors.New("questionnaire not found")
	}

	if !questionnaire.Active {
		return nil, errors.New("questionnaire is not active")
	}

	existingResponse, _ := s.repo.GetResponseByStudentAndQuestionnaire(studentProfileID, req.QuestionnaireID)
	if existingResponse != nil {
		return nil, errors.New("questionnaire already submitted")
	}

	questions, err := s.repo.GetQuestionsByQuestionnaireID(req.QuestionnaireID)
	if err != nil {
		return nil, errors.New("failed to get questionnaire questions")
	}

	if len(req.Answers) != len(questions) {
		return nil, errors.New("answer count does not match question count")
	}

	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, errors.New("failed to serialize answers")
	}

	totalScore := s.calculateTotalScore(req.Answers, questions)
	maxScore := s.calculateMaxScore(questions)

	response := &questionnaireRepo.QuestionnaireResponse{
		QuestionnaireID:  req.QuestionnaireID,
		StudentProfileID: studentProfileID,
		SubmittedAt:      time.Now(),
		Answers:          string(answersJSON),
		TotalScore:       &totalScore,
	}

	err = s.repo.CreateQuestionnaireResponse(response)
	if err != nil {
		return nil, errors.New("failed to save questionnaire response")
	}

	go s.processAIRecommendation(response.ID, req.Answers, questions, maxScore)

	return &QuestionnaireSubmissionResponse{
		ResponseID:      response.ID,
		QuestionnaireID: response.QuestionnaireID,
		StudentID:       response.StudentProfileID,
		SubmittedAt:     response.SubmittedAt,
		Status:          "processing",
	}, nil
}

func (s *StudentQuestionnaireService) GetStudentRole(studentProfileID uuid.UUID) (*StudentRoleResponse, error) {
	latestResponse, err := s.repo.GetLatestResponseByStudentProfile(studentProfileID)
	if err != nil {
		if err.Error() == "belum ada respons" {
			return &StudentRoleResponse{
				HasCompletedQuestionnaire: false,
			}, nil
		}
		return nil, errors.New("failed to get questionnaire response")
	}

	response := &StudentRoleResponse{
		HasCompletedQuestionnaire: true,
	}

	if latestResponse.AIRecommendations != nil && *latestResponse.AIRecommendations != "" {
		recommendations, err := s.parseAIRecommendations(*latestResponse.AIRecommendations)
		if err == nil && len(recommendations) > 0 {
			topRecommendation := recommendations[0]

			roleDetails, err := s.repo.GetTargetRoleByName(topRecommendation.RoleName)
			if err == nil {
				response.RecommendedRole = &RecommendedRoleInfo{
					RoleID:        roleDetails.ID,
					RoleName:      roleDetails.Name,
					Description:   roleDetails.Description,
					Category:      roleDetails.Category,
					Score:         topRecommendation.Score,
					Justification: topRecommendation.Justification,
				}
			} else {
				response.RecommendedRole = &RecommendedRoleInfo{
					RoleID:        uuid.New(),
					RoleName:      topRecommendation.RoleName,
					Description:   "Role generated by AI analysis",
					Category:      topRecommendation.Category,
					Score:         topRecommendation.Score,
					Justification: topRecommendation.Justification,
				}
			}
		}
	}

	return response, nil
}

func (s *StudentQuestionnaireService) GetQuestionnaireResult(studentProfileID uuid.UUID, responseID uuid.UUID) (*QuestionnaireResultResponse, error) {
	response, err := s.repo.GetResponseByIDAndStudent(responseID, studentProfileID)
	if err != nil {
		return nil, errors.New("questionnaire result not found")
	}

	questions, err := s.repo.GetQuestionsByQuestionnaireID(response.QuestionnaireID)
	if err != nil {
		return nil, errors.New("failed to get questionnaire questions")
	}

	maxScore := s.calculateMaxScore(questions)
	status := s.determineResponseStatus(response)

	result := &QuestionnaireResultResponse{
		ID:              response.ID,
		QuestionnaireID: response.QuestionnaireID,
		StudentID:       response.StudentProfileID,
		SubmittedAt:     response.SubmittedAt,
		TotalScore:      response.TotalScore,
		MaxScore:        maxScore,
		Status:          status,
	}

	if response.AIRecommendations != nil && *response.AIRecommendations != "" {
		recommendations, err := s.parseAIRecommendations(*response.AIRecommendations)
		if err == nil && len(recommendations) > 0 {
			topRecommendation := recommendations[0]

			roleDetails, err := s.repo.GetTargetRoleByName(topRecommendation.RoleName)
			if err == nil {
				result.RecommendedRole = &RecommendedRoleInfo{
					RoleID:        roleDetails.ID,
					RoleName:      roleDetails.Name,
					Description:   roleDetails.Description,
					Category:      roleDetails.Category,
					Score:         topRecommendation.Score,
					Justification: topRecommendation.Justification,
				}
			} else {
				result.RecommendedRole = &RecommendedRoleInfo{
					RoleID:        uuid.New(),
					RoleName:      topRecommendation.RoleName,
					Description:   "Role generated by AI analysis",
					Category:      topRecommendation.Category,
					Score:         topRecommendation.Score,
					Justification: topRecommendation.Justification,
				}
			}
		}
	}

	return result, nil
}

func (s *StudentQuestionnaireService) calculateTotalScore(answers []AnswerItem, questions []questionnaireRepo.QuestionnaireQuestion) int {
	totalScore := 0
	for _, answer := range answers {
		totalScore += answer.Score
	}
	return totalScore
}

func (s *StudentQuestionnaireService) calculateMaxScore(questions []questionnaireRepo.QuestionnaireQuestion) int {
	maxScore := 0
	for _, question := range questions {
		maxScore += question.MaxScore
	}
	return maxScore
}

func (s *StudentQuestionnaireService) determineResponseStatus(response *questionnaireRepo.QuestionnaireResponse) string {
	if response.ProcessedAt == nil {
		return "processing"
	}
	if response.AIRecommendations == nil || *response.AIRecommendations == "" {
		return "failed"
	}
	return "completed"
}

func (s *StudentQuestionnaireService) parseAIRecommendations(recommendationsJSON string) ([]struct {
	RoleID        string  `json:"role_id"`
	RoleName      string  `json:"role_name"`
	Score         float64 `json:"score"`
	Justification string  `json:"justification"`
	Category      string  `json:"category"`
}, error) {
	var recommendations []struct {
		RoleID        string  `json:"role_id"`
		RoleName      string  `json:"role_name"`
		Score         float64 `json:"score"`
		Justification string  `json:"justification"`
		Category      string  `json:"category"`
	}

	err := json.Unmarshal([]byte(recommendationsJSON), &recommendations)
	return recommendations, err
}

func (s *StudentQuestionnaireService) processAIRecommendation(responseID uuid.UUID, answers []AnswerItem, questions []questionnaireRepo.QuestionnaireQuestion, maxScore int) {
	ctx := context.Background()

	aiInput := s.buildAIPrompt(answers, questions, maxScore)

	aiResponse, err := s.aiService.GenerateCareerRecommendations(ctx, aiInput)
	if err != nil {
		s.repo.UpdateResponseProcessingStatus(responseID, "failed", nil, nil)
		return
	}

	recommendations, err := json.Marshal(aiResponse.Recommendations)
	if err != nil {
		s.repo.UpdateResponseProcessingStatus(responseID, "failed", nil, nil)
		return
	}

	analysis, err := json.Marshal(aiResponse.Analysis)
	if err != nil {
		s.repo.UpdateResponseProcessingStatus(responseID, "failed", nil, nil)
		return
	}

	recommendationsStr := string(recommendations)
	analysisStr := string(analysis)
	processedAt := time.Now()

	err = s.repo.UpdateResponseProcessingStatus(responseID, "completed", &recommendationsStr, &analysisStr)
	if err != nil {
		return
	}

	s.repo.UpdateResponseProcessedAt(responseID, &processedAt)

	if len(aiResponse.Recommendations) > 0 {
		topRecommendation := aiResponse.Recommendations[0]
		roleID, _ := uuid.Parse(topRecommendation.RoleID)
		s.repo.UpdateResponseRecommendedRole(responseID, &roleID)
	}
}

func (s *StudentQuestionnaireService) buildAIPrompt(answers []AnswerItem, questions []questionnaireRepo.QuestionnaireQuestion, maxScore int) string {
	questionMap := make(map[uuid.UUID]questionnaireRepo.QuestionnaireQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	prompt := "Analyze the following questionnaire responses and provide career recommendations:\n\n"

	for _, answer := range answers {
		if question, exists := questionMap[answer.QuestionID]; exists {
			prompt += fmt.Sprintf("Q: %s\nA: %s (Score: %d)\n\n",
				question.QuestionText, answer.Answer, answer.Score)
		}
	}

	prompt += fmt.Sprintf("Total Score: %d/%d\n\n", s.calculateTotalScore(answers, questions), maxScore)
	prompt += "Please provide career recommendations in JSON format with role_id, role_name, score, justification, and category."

	return prompt
}
