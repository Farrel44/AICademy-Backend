package student

import (
	"time"

	"github.com/google/uuid"
)

// Dashboard roadmap response - same structure as my-roadmap
type DashboardRoadmapResponse struct {
	ID                uuid.UUID                 `json:"id"`
	RoadmapName       string                    `json:"roadmap_name"`
	Description       *string                   `json:"description"`
	RoleName          string                    `json:"role_name"`
	TotalSteps        int                       `json:"total_steps"`
	EstimatedDuration int                       `json:"estimated_duration"`
	DifficultyLevel   string                    `json:"difficulty_level"`
	Progress          *DashboardRoadmapProgress `json:"progress,omitempty"`
	Steps             []DashboardStepResponse   `json:"steps"`
}

type DashboardRoadmapProgress struct {
	ID              uuid.UUID  `json:"id"`
	TotalSteps      int        `json:"total_steps"`
	CompletedSteps  int        `json:"completed_steps"`
	ProgressPercent float64    `json:"progress_percent"`
	IsFinished      bool       `json:"is_finished"`
	StartedAt       *time.Time `json:"started_at"`
	LastActivityAt  *time.Time `json:"last_activity_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

type DashboardStepResponse struct {
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

type StudentChallengeItem struct {
	ChallengeID    uuid.UUID  `json:"challenge_id"`
	ChallengeTitle string     `json:"challenge_title"`
	TeamID         uuid.UUID  `json:"team_id"`
	TeamName       string     `json:"team_name"`
	SubmissionID   *uuid.UUID `json:"submission_id"`
	Points         int        `json:"points"`
	Deadline       time.Time  `json:"deadline"`
	IsActive       bool       `json:"is_active"`
	RegisteredAt   time.Time  `json:"registered_at"`
}

type StudentSummary struct {
	TotalRoadmaps       int `json:"total_roadmaps"`
	CompletedSteps      int `json:"completed_steps"`
	TotalSteps          int `json:"total_steps"`
	ActiveChallenges    int `json:"active_challenges"`
	CompletedChallenges int `json:"completed_challenges"`
	TotalPoints         int `json:"total_points"`
}

type StudentDashboardData struct {
	Roadmap    *DashboardRoadmapResponse `json:"roadmap,omitempty"`
	Challenges []StudentChallengeItem    `json:"challenges"`
	Summary    StudentSummary            `json:"summary"`
}
