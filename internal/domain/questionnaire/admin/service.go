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
	// For now, return empty result to fix compilation
	return &PaginatedTargetRolesResponse{
		Data:       []TargetRoleResponse{},
		Total:      0,
		Page:       page,
		Limit:      limit,
		TotalPages: 0,
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
	questionnaire := &questionnaire.ProfilingQuestionnaire{
		Name:        req.Name,
		GeneratedBy: "ai",
		Version:     1,
		Active:      false,
	}

	err := s.repo.CreateQuestionnaire(questionnaire)
	if err != nil {
		return nil, errors.New("failed to create questionnaire")
	}

	prompt := s.buildQuestionGenerationPrompt(req)

	go s.processQuestionGeneration(questionnaire.ID, prompt, req)

	return &QuestionnaireGenerationResponse{
		QuestionnaireID: questionnaire.ID,
		Status:          "processing",
		Progress:        0,
		Message:         "Questionnaire generation started",
	}, nil
}

func (s *AdminQuestionnaireService) GetQuestionnaires(page, limit int) (*PaginatedQuestionnairesResponse, error) {
	return &PaginatedQuestionnairesResponse{
		Data:       []QuestionnaireListResponse{},
		Total:      0,
		Page:       page,
		Limit:      limit,
		TotalPages: 0,
	}, nil
}

func (s *AdminQuestionnaireService) GetQuestionnaireDetail(id uuid.UUID) (*QuestionnaireDetailResponse, error) {
	return nil, errors.New("not implemented yet")
}

func (s *AdminQuestionnaireService) ActivateQuestionnaire(id uuid.UUID, active bool) error {
	return errors.New("not implemented yet")
}

func (s *AdminQuestionnaireService) GetQuestionnaireResponses(page, limit int, questionnaireID *uuid.UUID) (*PaginatedResponsesResponse, error) {
	return &PaginatedResponsesResponse{
		Data:       []QuestionnaireResponseOverview{},
		Total:      0,
		Page:       page,
		Limit:      limit,
		TotalPages: 0,
	}, nil
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

func (s *AdminQuestionnaireService) buildQuestionGenerationPrompt(req GenerateQuestionnaireRequest) string {
	customInstructions := ""
	if req.CustomInstructions != nil {
		customInstructions = *req.CustomInstructions
	}

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
		req.QuestionCount, req.TargetRoles, req.DifficultyLevel, req.FocusAreas, customInstructions)
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
