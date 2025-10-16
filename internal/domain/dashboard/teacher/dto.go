package teacher

import (
	"time"

	"github.com/google/uuid"
)

type TeacherChallengeStats struct {
	TotalChallenges     int `json:"total_challenges"`
	ActiveChallenges    int `json:"active_challenges"`
	CompletedChallenges int `json:"completed_challenges"`
}

type TeacherSubmissionStats struct {
	TotalSubmissions   int `json:"total_submissions"`
	ScoredSubmissions  int `json:"scored_submissions"`
	PendingSubmissions int `json:"pending_submissions"`
}

type TeacherRoadmapStats struct {
	TotalStepSubmissions int `json:"total_step_submissions"`
	ValidatedSubmissions int `json:"validated_submissions"`
	PendingValidations   int `json:"pending_validations"`
}

type TeacherChallenge struct {
	ID                  uuid.UUID `json:"id"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Deadline            time.Time `json:"deadline"`
	Prize               *string   `json:"prize"`
	MaxParticipants     int       `json:"max_participants"`
	CurrentParticipants int       `json:"current_participants"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
}

type ChallengeSubmissionItem struct {
	ID               uuid.UUID  `json:"id"`
	ChallengeID      uuid.UUID  `json:"challenge_id"`
	ChallengeTitle   string     `json:"challenge_title"`
	Title            string     `json:"title"`
	TeamID           *uuid.UUID `json:"team_id"`
	TeamName         *string    `json:"team_name"`
	StudentProfileID *uuid.UUID `json:"student_profile_id"`
	StudentName      *string    `json:"student_name"`
	ImageURL         *string    `json:"image_url"`
	RepoURL          *string    `json:"repo_url"`
	DocsURL          *string    `json:"docs_url"`
	SubmittedAt      time.Time  `json:"submitted_at"`
	Points           *int       `json:"points"`
	IsScored         bool       `json:"is_scored"`
}

type RoadmapStepSubmission struct {
	ID               uuid.UUID  `json:"id"`
	RoadmapStepID    uuid.UUID  `json:"roadmap_step_id"`
	StepTitle        string     `json:"step_title"`
	StudentProfileID uuid.UUID  `json:"student_profile_id"`
	StudentName      string     `json:"student_name"`
	EvidenceLink     *string    `json:"evidence_link"`
	EvidenceType     *string    `json:"evidence_type"`
	SubmissionNotes  *string    `json:"submission_notes"`
	ValidationNotes  *string    `json:"validation_notes"`
	ValidationScore  *int       `json:"validation_score"`
	Status           string     `json:"status"`
	SubmittedAt      *time.Time `json:"submitted_at"`
	IsValidated      bool       `json:"is_validated"`
}

type TeacherDashboardData struct {
	ChallengeStats       TeacherChallengeStats     `json:"challenge_stats"`
	SubmissionStats      TeacherSubmissionStats    `json:"submission_stats"`
	RoadmapStats         TeacherRoadmapStats       `json:"roadmap_stats"`
	Challenges           []TeacherChallenge        `json:"challenges"`
	ChallengeSubmissions []ChallengeSubmissionItem `json:"challenge_submissions"`
	RoadmapSubmissions   []RoadmapStepSubmission   `json:"roadmap_submissions"`
}
