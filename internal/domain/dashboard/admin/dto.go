package admin

import (
	"time"

	"github.com/google/uuid"
)

type AdminTotals struct {
	TotalUsers     int64 `json:"total_users"`
	TotalStudents  int64 `json:"total_students"`
	TotalTeachers  int64 `json:"total_teachers"`
	TotalCompanies int64 `json:"total_companies"`
}

type StudentStatistics struct {
	TotalTKJ  int64 `json:"total_tkj"`
	TotalTJA  int64 `json:"total_tja"`
	TotalPPLG int64 `json:"total_pplg"`
	TotalRPL  int64 `json:"total_rpl"`
}

type AdminChallenge struct {
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

type AdminDashboardData struct {
	Totals             AdminTotals             `json:"totals"`
	StudentStats       StudentStatistics       `json:"student_stats"`
	Challenges         []AdminChallenge        `json:"challenges"`
	RoadmapSubmissions []RoadmapStepSubmission `json:"roadmap_submissions"`
}

type RoadmapStepSubmission struct {
	ID               uuid.UUID `json:"id"`
	RoadmapStepID    uuid.UUID `json:"roadmap_step_id"`
	StepTitle        string    `json:"step_title"`
	StudentProfileID uuid.UUID `json:"student_profile_id"`
	StudentName      string    `json:"student_name"`
	EvidenceLink     string    `json:"evidence_link"`
	EvidenceType     string    `json:"evidence_type"`
	SubmissionNotes  string    `json:"submission_notes"`
	ValidationNotes  *string   `json:"validation_notes"`
	ValidationScore  *int      `json:"validation_score"`
	Status           string    `json:"status"`
	SubmittedAt      time.Time `json:"submitted_at"`
	IsValidated      bool      `json:"is_validated"`
}
