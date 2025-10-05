package student

import (
	"time"

	"github.com/google/uuid"
)

type ActiveQuestionnaireResponse struct {
	ID          uuid.UUID                  `json:"id"`
	Name        string                     `json:"name"`
	Version     int                        `json:"version"`
	Questions   []QuestionnaireQuestionDTO `json:"questions"`
	Instruction string                     `json:"instruction"`
	Submitted   bool                       `json:"submitted"`
}

type QuestionnaireQuestionDTO struct {
	ID           uuid.UUID   `json:"id"`
	QuestionText string      `json:"question_text"`
	QuestionType string      `json:"question_type"`
	Options      []OptionDTO `json:"options,omitempty"`
	Category     string      `json:"category"`
	Order        int         `json:"order"`
}

type OptionDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Score int    `json:"score,omitempty"`
}

type SubmitQuestionnaireRequest struct {
	QuestionnaireID uuid.UUID    `json:"questionnaire_id" validate:"required"`
	Answers         []AnswerItem `json:"answers" validate:"required,dive"`
}

type AnswerItem struct {
	QuestionID uuid.UUID `json:"question_id" validate:"required"`
	Answer     string    `json:"answer" validate:"required"`
	Score      int       `json:"score,omitempty"`
}

type QuestionnaireSubmissionResponse struct {
	ResponseID      uuid.UUID `json:"response_id"`
	QuestionnaireID uuid.UUID `json:"questionnaire_id"`
	StudentID       uuid.UUID `json:"student_id"`
	SubmittedAt     time.Time `json:"submitted_at"`
	Status          string    `json:"status"`
}

type StudentRoleResponse struct {
	HasCompletedQuestionnaire bool                 `json:"has_completed_questionnaire"`
	RecommendedRole           *RecommendedRoleInfo `json:"recommended_role,omitempty"`
}

type RecommendedRoleInfo struct {
	RoleID        uuid.UUID `json:"role_id"`
	RoleName      string    `json:"role_name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Score         float64   `json:"score"`
	Justification string    `json:"justification"`
}

type QuestionnaireResultResponse struct {
	ID              uuid.UUID            `json:"id"`
	QuestionnaireID uuid.UUID            `json:"questionnaire_id"`
	StudentID       uuid.UUID            `json:"student_id"`
	SubmittedAt     time.Time            `json:"submitted_at"`
	TotalScore      *int                 `json:"total_score,omitempty"`
	MaxScore        int                  `json:"max_score"`
	RecommendedRole *RecommendedRoleInfo `json:"recommended_role,omitempty"`
	Status          string               `json:"status"`
}
