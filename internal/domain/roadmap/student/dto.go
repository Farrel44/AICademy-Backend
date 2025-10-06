package student

import (
	"time"

	"github.com/google/uuid"
)

type AvailableRoadmapResponse struct {
	ID                uuid.UUID `json:"id"`
	RoadmapName       string    `json:"roadmap_name"`
	Description       *string   `json:"description"`
	RoleName          string    `json:"role_name"`
	TotalSteps        int       `json:"total_steps"`
	EstimatedDuration int       `json:"estimated_duration"`
	DifficultyLevel   string    `json:"difficulty_level"`
	IsStarted         bool      `json:"is_started"`
	ProgressPercent   float64   `json:"progress_percent"`
}

type StartRoadmapRequest struct {
	RoadmapID uuid.UUID `json:"roadmap_id" validate:"required"`
}

type StartRoadmapResponse struct {
	Message         string                  `json:"message"`
	RoadmapProgress RoadmapProgressResponse `json:"roadmap_progress"`
}

// Roadmap Progress DTOs
type RoadmapProgressResponse struct {
	ID                 uuid.UUID             `json:"id"`
	RoadmapID          uuid.UUID             `json:"roadmap_id"`
	RoadmapName        string                `json:"roadmap_name"`
	RoadmapDescription *string               `json:"roadmap_description"`
	RoleName           string                `json:"role_name"`
	TotalSteps         int                   `json:"total_steps"`
	CompletedSteps     int                   `json:"completed_steps"`
	ProgressPercent    float64               `json:"progress_percent"`
	StartedAt          *time.Time            `json:"started_at"`
	LastActivityAt     *time.Time            `json:"last_activity_at"`
	CompletedAt        *time.Time            `json:"completed_at"`
	Steps              []StudentStepResponse `json:"steps"`
}

type StudentStepResponse struct {
	ID                   uuid.UUID `json:"id"`
	StepOrder            int       `json:"step_order"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	LearningObjectives   string    `json:"learning_objectives"`
	SubmissionGuidelines string    `json:"submission_guidelines"`
	ResourceLinks        *string   `json:"resource_links"`
	EstimatedDuration    int       `json:"estimated_duration"`
	DifficultyLevel      string    `json:"difficulty_level"`
	Status               string    `json:"status"`

	// Progress info
	EvidenceLink    *string    `json:"evidence_link"`
	EvidenceType    *string    `json:"evidence_type"`
	SubmissionNotes *string    `json:"submission_notes"`
	ValidationNotes *string    `json:"validation_notes"`
	ValidationScore *int       `json:"validation_score"`
	StartedAt       *time.Time `json:"started_at"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	CompletedAt     *time.Time `json:"completed_at"`

	// Helper flags
	CanStart  bool `json:"can_start"`
	CanSubmit bool `json:"can_submit"`
	IsLocked  bool `json:"is_locked"`
}

// Step Actions DTOs
type StartStepRequest struct {
	StepID uuid.UUID `json:"step_id" validate:"required"`
}

type StartStepResponse struct {
	Message    string    `json:"message"`
	StepStatus string    `json:"step_status"`
	StartedAt  time.Time `json:"started_at"`
}

type SubmitEvidenceRequest struct {
	StepID          uuid.UUID `json:"step_id" validate:"required"`
	EvidenceLink    string    `json:"evidence_link" validate:"required,url"`
	EvidenceType    string    `json:"evidence_type" validate:"required,oneof=url"` //nanti di tambahin lagi untuk type nya, misal file
	SubmissionNotes *string   `json:"submission_notes,omitempty" validate:"omitempty,max=1000"`
}

type SubmitEvidenceResponse struct {
	Message     string    `json:"message"`
	StepStatus  string    `json:"step_status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// My Progress DTOs
type MyProgressResponse struct {
	ActiveRoadmaps    []ActiveRoadmapSummary    `json:"active_roadmaps"`
	CompletedRoadmaps []CompletedRoadmapSummary `json:"completed_roadmaps"`
	OverallStats      StudentStatistics         `json:"overall_stats"`
}

type ActiveRoadmapSummary struct {
	ID                  uuid.UUID  `json:"id"`
	RoadmapName         string     `json:"roadmap_name"`
	RoleName            string     `json:"role_name"`
	TotalSteps          int        `json:"total_steps"`
	CompletedSteps      int        `json:"completed_steps"`
	ProgressPercent     float64    `json:"progress_percent"`
	CurrentStep         *StepInfo  `json:"current_step"`
	StartedAt           *time.Time `json:"started_at"`
	LastActivityAt      *time.Time `json:"last_activity_at"`
	EstimatedCompletion *time.Time `json:"estimated_completion"`
}

type CompletedRoadmapSummary struct {
	ID            uuid.UUID `json:"id"`
	RoadmapName   string    `json:"roadmap_name"`
	RoleName      string    `json:"role_name"`
	TotalSteps    int       `json:"total_steps"`
	CompletedAt   time.Time `json:"completed_at"`
	TotalDuration int       `json:"total_duration"`
	AverageScore  *float64  `json:"average_score"`
}

type StepInfo struct {
	ID              uuid.UUID `json:"id"`
	Order           int       `json:"order"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	DifficultyLevel string    `json:"difficulty_level"`
}

type StudentStatistics struct {
	TotalRoadmapsStarted   int     `json:"total_roadmaps_started"`
	TotalRoadmapsCompleted int     `json:"total_roadmaps_completed"`
	TotalStepsCompleted    int     `json:"total_steps_completed"`
	TotalHoursSpent        int     `json:"total_hours_spent"`
	AverageCompletionRate  float64 `json:"average_completion_rate"`
	StreakDays             int     `json:"streak_days"`
}

// Detailed Step View DTOs
type StepDetailResponse struct {
	StepInfo        StudentStepResponse `json:"step_info"`
	RoadmapContext  RoadmapContext      `json:"roadmap_context"`
	ProgressContext ProgressContext     `json:"progress_context"`
	Resources       []ResourceInfo      `json:"resources"`
}

type RoadmapContext struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RoleName    string    `json:"role_name"`
	TotalSteps  int       `json:"total_steps"`
	CurrentStep int       `json:"current_step"`
}

type ProgressContext struct {
	PreviousStep *StepInfo `json:"previous_step"`
	NextStep     *StepInfo `json:"next_step"`
	StepsLeft    int       `json:"steps_left"`
}

type ResourceInfo struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Type        string `json:"type"` // video, article, tool, etc.
	Description string `json:"description"`
}

// Response wrapper DTOs
type PaginatedRoadmapsResponse struct {
	Data       []AvailableRoadmapResponse `json:"data"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}

// Achievement/Badge DTOs (Future enhancement)
type AchievementInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	EarnedAt    time.Time `json:"earned_at"`
}

// Notification DTOs (Future enhancement)
type NotificationInfo struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"` // step_approved, step_rejected, roadmap_completed
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
