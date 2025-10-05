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

func (s *StudentQuestionnaireService) GetActiveQuestionnaire(studentProfileID uuid.UUID) (*ActiveQuestionnaireResponse, error) {
	questionnaire, err := s.repo.GetActiveQuestionnaire()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("no active questionnaire found")
		}
		return nil, errors.New("failed to get active questionnaire")
	}

	existingResponse, _ := s.repo.GetResponseByStudentAndQuestionnaire(studentProfileID, questionnaire.ID)
	submitted := existingResponse != nil

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
		Submitted:   submitted,
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

	if latestResponse.RecommendedProfilingRoleID != nil {
		roleDetails, err := s.repo.GetTargetRoleByID(*latestResponse.RecommendedProfilingRoleID)
		if err == nil {
			var topRecommendation *struct {
				RoleID        string  `json:"role_id"`
				RoleName      string  `json:"role_name"`
				Score         float64 `json:"score"`
				Justification string  `json:"justification"`
				Category      string  `json:"category"`
			}

			if latestResponse.AIRecommendations != nil && *latestResponse.AIRecommendations != "" {
				recommendations, err := s.parseAIRecommendations(*latestResponse.AIRecommendations)
				if err == nil && len(recommendations) > 0 {
					topRecommendation = &recommendations[0]
				}
			}

			response.RecommendedRole = &RecommendedRoleInfo{
				RoleID:        roleDetails.ID,
				RoleName:      roleDetails.Name,
				Description:   roleDetails.Description,
				Category:      roleDetails.Category,
				Score:         90,
				Justification: "Career recommendation based on questionnaire analysis",
			}

			if topRecommendation != nil {
				response.RecommendedRole.Score = topRecommendation.Score
				response.RecommendedRole.Justification = topRecommendation.Justification
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

func (s *StudentQuestionnaireService) GetLatestQuestionnaireResult(studentProfileID uuid.UUID) (*QuestionnaireResultResponse, error) {
	latestResponse, err := s.repo.GetLatestResponseByStudentProfile(studentProfileID)
	if err != nil {
		if err.Error() == "belum ada respons" {
			return nil, errors.New("belum ada kuesioner yang dikerjakan")
		}
		return nil, errors.New("gagal mengambil hasil kuesioner")
	}

	questions, err := s.repo.GetQuestionsByQuestionnaireID(latestResponse.QuestionnaireID)
	if err != nil {
		return nil, errors.New("gagal mengambil pertanyaan kuesioner")
	}

	maxScore := s.calculateMaxScore(questions)
	status := s.determineResponseStatus(latestResponse)

	result := &QuestionnaireResultResponse{
		ID:              latestResponse.ID,
		QuestionnaireID: latestResponse.QuestionnaireID,
		StudentID:       latestResponse.StudentProfileID,
		SubmittedAt:     latestResponse.SubmittedAt,
		TotalScore:      latestResponse.TotalScore,
		MaxScore:        maxScore,
		Status:          status,
	}

	if latestResponse.AIRecommendations != nil && *latestResponse.AIRecommendations != "" {
		recommendations, err := s.parseAIRecommendations(*latestResponse.AIRecommendations)
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
	prompt := s.buildAIPrompt(answers, questions, maxScore)

	ctx := context.Background()
	aiResponse, err := s.aiService.GenerateCareerRecommendations(ctx, prompt)
	if err != nil {
		return
	}

	response, err := s.repo.GetResponseByID(responseID)
	if err != nil {
		return
	}

	recommendationsJSON, err := json.Marshal(aiResponse.Recommendations)
	if err != nil {
		return
	}

	analysisJSON, err := json.Marshal(aiResponse.Analysis)
	if err != nil {
		return
	}

	var recommendedRoleID *uuid.UUID
	if len(aiResponse.Recommendations) > 0 {
		topRecommendation := aiResponse.Recommendations[0]

		targetRole, err := s.repo.GetTargetRoleByName(topRecommendation.RoleName)
		if err != nil {
			allTargetRoles, _, err := s.repo.GetTargetRoles(0, 1)
			if err == nil && len(allTargetRoles) > 0 {
				recommendedRoleID = &allTargetRoles[0].ID
			}
		} else {
			recommendedRoleID = &targetRole.ID
		}
	}

	now := time.Now()
	response.ProcessedAt = &now
	recommendationsStr := string(recommendationsJSON)
	analysisStr := string(analysisJSON)
	response.AIRecommendations = &recommendationsStr
	response.AIAnalysis = &analysisStr
	response.RecommendedProfilingRoleID = recommendedRoleID

	s.repo.UpdateResponse(response)
}

func (s *StudentQuestionnaireService) buildAIPrompt(answers []AnswerItem, questions []questionnaireRepo.QuestionnaireQuestion, maxScore int) string {
	questionMap := make(map[uuid.UUID]questionnaireRepo.QuestionnaireQuestion)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	availableRoles, _, _ := s.repo.GetTargetRoles(0, 100)
	var rolesList string
	if len(availableRoles) > 0 {
		rolesList = "\nAvailable Target Roles (you MUST only recommend from these):\n"
		for _, role := range availableRoles {
			rolesList += fmt.Sprintf("- %s (Category: %s): %s\n", role.Name, role.Category, role.Description)
		}
		rolesList += "\n"
	}

	prompt := "Analyze the following questionnaire responses and provide career recommendations:\n\n"

	for _, answer := range answers {
		if question, exists := questionMap[answer.QuestionID]; exists {
			prompt += fmt.Sprintf("Q: %s\nA: %s (Score: %d)\n\n",
				question.QuestionText, answer.Answer, answer.Score)
		}
	}

	prompt += fmt.Sprintf("Total Score: %d/%d\n\n", s.calculateTotalScore(answers, questions), maxScore)
	prompt += rolesList
	prompt += "IMPORTANT: You MUST only recommend roles from the available target roles list above. Use the exact role names.\n\n"
	prompt += "Please provide career recommendations in JSON format with role_id (use arbitrary ID), role_name (EXACT match from available roles), score, justification, and category."

	return prompt
}
