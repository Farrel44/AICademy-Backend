package admin

import (
	"time"

	"github.com/google/uuid"
)

// Roadmap Management DTOs
type CreateRoadmapRequest struct {
	ProfilingRoleID uuid.UUID `json:"profiling_role_id" validate:"required"`
	RoadmapName     string    `json:"roadmap_name" validate:"required,min=3,max=100"`
	Description     *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
}

type UpdateRoadmapRequest struct {
	RoadmapName string  `json:"roadmap_name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Status      string  `json:"status,omitempty" validate:"omitempty,oneof=draft active archived"`
}

type RoadmapResponse struct {
	ID              uuid.UUID `json:"id"`
	ProfilingRoleID uuid.UUID `json:"profiling_role_id"`
	RoadmapName     string    `json:"roadmap_name"`
	Description     *string   `json:"description"`
	Status          string    `json:"status"`
	CreatedBy       uuid.UUID `json:"created_by"`
	TotalSteps      int       `json:"total_steps"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Optional populated fields
	ProfilingRole   *RoleInfo         `json:"profiling_role,omitempty"`
	CreatedByUser   *UserInfo         `json:"created_by_user,omitempty"`
	Steps           []StepResponse    `json:"steps,omitempty"`
	StudentProgress []ProgressSummary `json:"student_progress,omitempty"`
}

type RoleInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
}

type UserInfo struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// Step Management DTOs
type CreateStepRequest struct {
	Title                string  `json:"title" validate:"required,min=3,max=200"`
	Description          string  `json:"description" validate:"required,min=10"`
	LearningObjectives   string  `json:"learning_objectives" validate:"required,min=10"`
	SubmissionGuidelines string  `json:"submission_guidelines" validate:"required,min=10"`
	ResourceLinks        *string `json:"resource_links,omitempty"`
	EstimatedDuration    int     `json:"estimated_duration" validate:"required,min=1,max=500"`
	DifficultyLevel      string  `json:"difficulty_level" validate:"required,oneof=beginner intermediate advanced"`
}

type UpdateStepRequest struct {
	Title                string  `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Description          string  `json:"description,omitempty" validate:"omitempty,min=10"`
	LearningObjectives   string  `json:"learning_objectives,omitempty" validate:"omitempty,min=10"`
	SubmissionGuidelines string  `json:"submission_guidelines,omitempty" validate:"omitempty,min=10"`
	ResourceLinks        *string `json:"resource_links,omitempty"`
	EstimatedDuration    int     `json:"estimated_duration,omitempty" validate:"omitempty,min=1,max=500"`
	DifficultyLevel      string  `json:"difficulty_level,omitempty" validate:"omitempty,oneof=beginner intermediate advanced"`
	StepOrder            int     `json:"step_order,omitempty" validate:"omitempty,min=1"`
}

type StepResponse struct {
	ID                   uuid.UUID `json:"id"`
	RoadmapID            uuid.UUID `json:"roadmap_id"`
	StepOrder            int       `json:"step_order"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	LearningObjectives   string    `json:"learning_objectives"`
	SubmissionGuidelines string    `json:"submission_guidelines"`
	ResourceLinks        *string   `json:"resource_links"`
	EstimatedDuration    int       `json:"estimated_duration"`
	DifficultyLevel      string    `json:"difficulty_level"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Progress Monitoring DTOs
type ProgressSummary struct {
	StudentID       uuid.UUID  `json:"student_id"`
	StudentName     string     `json:"student_name"`
	StudentEmail    string     `json:"student_email"`
	TotalSteps      int        `json:"total_steps"`
	CompletedSteps  int        `json:"completed_steps"`
	ProgressPercent float64    `json:"progress_percent"`
	StartedAt       *time.Time `json:"started_at"`
	LastActivityAt  *time.Time `json:"last_activity_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

type StudentProgressDetail struct {
	StudentInfo     StudentInfo          `json:"student_info"`
	RoadmapInfo     RoadmapSummary       `json:"roadmap_info"`
	OverallProgress ProgressSummary      `json:"overall_progress"`
	StepProgress    []StepProgressDetail `json:"step_progress"`
}

type StudentInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	NIS   string    `json:"nis"`
	Class string    `json:"class"`
}

type RoadmapSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	TotalSteps  int       `json:"total_steps"`
}

type StepProgressDetail struct {
	StepInfo        StepSummary     `json:"step_info"`
	Status          string          `json:"status"`
	EvidenceLink    *string         `json:"evidence_link"`
	EvidenceType    *string         `json:"evidence_type"`
	SubmissionNotes *string         `json:"submission_notes"`
	ValidationInfo  *ValidationInfo `json:"validation_info,omitempty"`
	StartedAt       *time.Time      `json:"started_at"`
	SubmittedAt     *time.Time      `json:"submitted_at"`
	CompletedAt     *time.Time      `json:"completed_at"`
}

type StepSummary struct {
	ID                uuid.UUID `json:"id"`
	Order             int       `json:"order"`
	Title             string    `json:"title"`
	EstimatedDuration int       `json:"estimated_duration"`
	DifficultyLevel   string    `json:"difficulty_level"`
}

type ValidationInfo struct {
	ValidatedBy     *UserInfo  `json:"validated_by"`
	ValidationNotes *string    `json:"validation_notes"`
	ValidationScore *int       `json:"validation_score"`
	ValidatedAt     *time.Time `json:"validated_at"`
}

// Submission Review DTOs
type PendingSubmissionResponse struct {
	ID              uuid.UUID      `json:"id"`
	StudentInfo     StudentInfo    `json:"student_info"`
	RoadmapInfo     RoadmapSummary `json:"roadmap_info"`
	StepInfo        StepSummary    `json:"step_info"`
	EvidenceLink    *string        `json:"evidence_link"`
	EvidenceType    *string        `json:"evidence_type"`
	SubmissionNotes *string        `json:"submission_notes"`
	SubmittedAt     *time.Time     `json:"submitted_at"`
}

// List Response DTOs
type PaginatedRoadmapsResponse struct {
	Data       []RoadmapResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type PaginatedStepsResponse struct {
	Data       []StepResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

type PaginatedSubmissionsResponse struct {
	Data       []PendingSubmissionResponse `json:"data"`
	Total      int64                       `json:"total"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalPages int                         `json:"total_pages"`
}

type RoadmapStatistics struct {
	TotalRoadmaps    int64                   `json:"total_roadmaps"`
	ActiveRoadmaps   int64                   `json:"active_roadmaps"`
	DraftRoadmaps    int64                   `json:"draft_roadmaps"`
	TotalSteps       int64                   `json:"total_steps"`
	StudentsEnrolled int64                   `json:"students_enrolled"`
	CompletionStats  CompletionStatistics    `json:"completion_stats"`
	RoadmapsByRole   []RoadmapRoleStatistics `json:"roadmaps_by_role"`
}

type CompletionStatistics struct {
	AverageCompletion  float64 `json:"average_completion"`
	StudentsCompleted  int64   `json:"students_completed"`
	StudentsInProgress int64   `json:"students_in_progress"`
	StudentsNotStarted int64   `json:"students_not_started"`
}

type RoadmapRoleStatistics struct {
	RoleID       uuid.UUID `json:"role_id"`
	RoleName     string    `json:"role_name"`
	RoadmapCount int64     `json:"roadmap_count"`
	StudentCount int64     `json:"student_count"`
}
