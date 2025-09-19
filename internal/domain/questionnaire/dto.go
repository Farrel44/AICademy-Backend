package questionnaire

import (
	"github.com/google/uuid"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type QuestionOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type AnswerItem struct {
	QuestionID     uuid.UUID `json:"question_id" validate:"required"`
	SelectedOption *string   `json:"selected_option,omitempty"`
	Score          *int      `json:"score,omitempty"`
	TextAnswer     *string   `json:"text_answer,omitempty"`
}

type SubmitQuestionnaireRequest struct {
	QuestionnaireID uuid.UUID    `json:"questionnaire_id" validate:"required"`
	Answers         []AnswerItem `json:"answers" validate:"required,dive"`
}

type GenerateQuestionnaireRequest struct {
	Name               string   `json:"name" validate:"required,min=2"`
	QuestionCount      int      `json:"question_count" validate:"required,min=5,max=50"`
	TargetRoles        []string `json:"target_roles" validate:"required,min=1"`
	DifficultyLevel    string   `json:"difficulty_level" validate:"required,oneof=basic intermediate advanced"`
	FocusAreas         []string `json:"focus_areas" validate:"required,min=1"`
	CustomInstructions string   `json:"custom_instructions,omitempty"`
}

type QuestionnaireQuestionResponse struct {
	ID            uuid.UUID        `json:"id"`
	QuestionText  string           `json:"question_text"`
	QuestionType  QuestionType     `json:"question_type"`
	MaxScore      int              `json:"max_score"`
	QuestionOrder int              `json:"question_order"`
	Category      string           `json:"category"`
	Options       []QuestionOption `json:"options,omitempty"`
}

type ActiveQuestionnaireResponse struct {
	ID        uuid.UUID                       `json:"id"`
	Name      string                          `json:"name"`
	Questions []QuestionnaireQuestionResponse `json:"questions"`
}

type QuestionnaireDetailResponse struct {
	ID          uuid.UUID                       `json:"id"`
	Name        string                          `json:"name"`
	Version     int                             `json:"version"`
	GeneratedBy string                          `json:"generated_by"`
	Questions   []QuestionnaireQuestionResponse `json:"questions"`
}

type AIRecommendation struct {
	RoleID        string  `json:"role_id"`
	RoleName      string  `json:"role_name"`
	Score         float64 `json:"score"`
	Justification string  `json:"justification"`
}

type QuestionnaireResultResponse struct {
	ID                uuid.UUID          `json:"id"`
	QuestionnaireID   uuid.UUID          `json:"questionnaire_id"`
	SubmittedAt       string             `json:"submitted_at"`
	ProcessedAt       *string            `json:"processed_at,omitempty"`
	TotalScore        *int               `json:"total_score,omitempty"`
	RecommendedRole   *string            `json:"recommended_role,omitempty"`
	AIRecommendations []AIRecommendation `json:"ai_recommendations,omitempty"`
}

type GenerationStatusResponse struct {
	QuestionnaireID uuid.UUID `json:"questionnaire_id"`
	Status          string    `json:"status"`
	Progress        int       `json:"progress"`
	Message         string    `json:"message"`
}
