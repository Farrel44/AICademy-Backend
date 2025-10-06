package teacher

import (
	"time"

	"github.com/google/uuid"
)

type ReviewSubmissionRequest struct {
	Action          string  `json:"action" validate:"required,oneof=approve reject"`
	ValidationScore *int    `json:"validation_score,omitempty" validate:"omitempty,min=0,max=100"`
	ValidationNotes *string `json:"validation_notes,omitempty" validate:"omitempty,max=1000"`
}

type PendingSubmissionResponse struct {
	ID                   uuid.UUID  `json:"id"`
	StudentName          string     `json:"student_name"`
	StudentEmail         string     `json:"student_email"`
	RoadmapName          string     `json:"roadmap_name"`
	StepTitle            string     `json:"step_title"`
	StepOrder            int        `json:"step_order"`
	EvidenceLink         *string    `json:"evidence_link"`
	EvidenceType         *string    `json:"evidence_type"`
	SubmissionNotes      *string    `json:"submission_notes"`
	SubmittedAt          *time.Time `json:"submitted_at"`
	LearningObjectives   string     `json:"learning_objectives"`
	SubmissionGuidelines string     `json:"submission_guidelines"`
}

type ReviewSubmissionResponse struct {
	Message    string `json:"message"`
	StepStatus string `json:"step_status"`
	ReviewedAt string `json:"reviewed_at"`
}

type PaginatedSubmissionsResponse struct {
	Data       []PendingSubmissionResponse `json:"data"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalPages int                         `json:"total_pages"`
}
