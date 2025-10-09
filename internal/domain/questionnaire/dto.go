package questionnaire

import (
	"time"

	"github.com/google/uuid"
)

// Types needed by repository and admin components

type QuestionnaireResponseSummary struct {
	ID              uuid.UUID  `json:"id"`
	StudentName     string     `json:"student_name"`
	StudentEmail    string     `json:"student_email"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	TotalScore      *int       `json:"total_score,omitempty"`
	RecommendedRole *string    `json:"recommended_role,omitempty"`
	Status          string     `json:"status"`
}

type AIRecommendation struct {
	RoleID        string  `json:"role_id"`
	RoleName      string  `json:"role_name"`
	Score         float64 `json:"score"`
	Justification string  `json:"justification"`
	Category      string  `json:"category"`
}

type QuestionOption struct {
	Text  string `json:"text"`
	Label string `json:"label"`
	Value string `json:"value"`
	Score int    `json:"score,omitempty"`
}
