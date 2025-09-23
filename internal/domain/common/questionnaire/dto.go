package questionnaire

import (
	"time"

	"github.com/google/uuid"
)

// Common DTOs untuk semua role

type ActiveQuestionnaireResponse struct {
	ID          uuid.UUID                   `json:"id"`
	Name        string                      `json:"name"`
	Version     int                         `json:"version"`
	Questions   []QuestionnaireQuestionDTO  `json:"questions"`
	Instruction string                      `json:"instruction"`
}

type QuestionnaireQuestionDTO struct {
	ID           uuid.UUID     `json:"id"`
	QuestionText string        `json:"question_text"`
	QuestionType string        `json:"question_type"`
	Options      []OptionDTO   `json:"options,omitempty"`
	Category     string        `json:"category"`
	Order        int           `json:"order"`
}

type OptionDTO struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Score int    `json:"score,omitempty"`
}

// Submit questionnaire request
type SubmitQuestionnaireRequest struct {
	QuestionnaireID uuid.UUID    `json:"questionnaire_id" validate:"required"`
	Answers         []AnswerItem `json:"answers" validate:"required,dive"`
}

type AnswerItem struct {
	QuestionID uuid.UUID `json:"question_id" validate:"required"`
	Answer     string    `json:"answer" validate:"required"`
	Score      int       `json:"score,omitempty"`
}

// Questionnaire response after submission
type QuestionnaireResponse struct {
	ID              uuid.UUID                  `json:"id"`
	QuestionnaireID uuid.UUID                  `json:"questionnaire_id"`
	StudentID       uuid.UUID                  `json:"student_id"`
	SubmittedAt     time.Time                  `json:"submitted_at"`
	TotalScore      int                        `json:"total_score"`
	MaxScore        int                        `json:"max_score"`
	Recommendations []CareerRecommendationDTO  `json:"recommendations"`
	Status          string                     `json:"status"`
}

type CareerRecommendationDTO struct {
	RoleID        uuid.UUID `json:"role_id"`
	RoleName      string    `json:"role_name"`
	Score         float64   `json:"score"`
	Justification string    `json:"justification"`
	Category      string    `json:"category"`
}

// Target Role DTOs
type TargetRoleDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}