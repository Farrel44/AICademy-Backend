package student_challenge

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// Search student DTOs
type SearchStudentRequest struct {
	Query string `json:"query" validate:"required,min=1"`
	Limit int    `json:"limit"`
}

type CreateTeamRequest struct {
	TeamName  string      `json:"team_name" validate:"required,min=3,max=100"`
	About     *string     `json:"about"`
	MemberIDs []uuid.UUID `json:"member_ids" validate:"required,min=2"`
}

type TeamMemberInfo struct {
	StudentProfileID uuid.UUID `json:"student_profile_id"`
	MemberRole       *string   `json:"member_role"`
	FullName         string    `json:"full_name"`
	NIS              string    `json:"nis"`
}

type CreateTeamResponse struct {
	ID          uuid.UUID        `json:"id"`
	TeamName    string           `json:"team_name"`
	About       *string          `json:"about"`
	CreatedAt   time.Time        `json:"created_at"`
	Members     []TeamMemberInfo `json:"members"`
	MemberCount int              `json:"member_count"`
}

// Challenge registration DTOs
type RegisterChallengeRequest struct {
	ChallengeID uuid.UUID `json:"challenge_id" validate:"required"`
	TeamID      uuid.UUID `json:"team_id" validate:"required"`
}

type RegisterChallengeResponse struct {
	Message             string    `json:"message"`
	ChallengeID         uuid.UUID `json:"challenge_id"`
	TeamID              uuid.UUID `json:"team_id"`
	RegistrationDate    time.Time `json:"registration_date"`
	CurrentParticipants int       `json:"current_participants"`
}

// Get available challenges
type ChallengeListResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	Deadline            time.Time  `json:"deadline"`
	Prize               *string    `json:"prize"`
	MaxParticipants     int        `json:"max_participants"`
	CurrentParticipants int        `json:"current_participants"`
	CanRegister         bool       `json:"can_register"`
	IsActive            bool       `json:"is_active"`
	OrganizerName       string     `json:"organizer_name"`
	IsRegistered        bool       `json:"is_registered"`
	MyTeamID            *uuid.UUID `json:"my_team_id,omitempty"`
	// New fields for winner info
	WinnerTeamID      *uuid.UUID `json:"winner_team_id,omitempty"`
	WinnerTeamName    *string    `json:"winner_team_name,omitempty"`
	IsWinnerAnnounced bool       `json:"is_winner_announced"`
}

type MyTeamResponse struct {
	ID                   uuid.UUID                 `json:"id"`
	TeamName             string                    `json:"team_name"`
	About                *string                   `json:"about"`
	CreatedAt            time.Time                 `json:"created_at"`
	Members              []TeamMemberInfo          `json:"members"`
	MemberCount          int                       `json:"member_count"`
	RegisteredChallenges []RegisteredChallengeInfo `json:"registered_challenges"`
}

type RegisteredChallengeInfo struct {
	ChallengeID    uuid.UUID `json:"challenge_id"`
	ChallengeTitle string    `json:"challenge_title"`
	Deadline       time.Time `json:"deadline"`
	IsActive       bool      `json:"is_active"`
}

// Submit challenge DTOs
type SubmitChallengeRequest struct {
	ChallengeID uuid.UUID             `form:"challenge_id" validate:"required"`
	Title       string                `form:"title" validate:"required,min=3,max=200"`
	RepoURL     *string               `form:"repo_url"`
	DocsFile    *multipart.FileHeader `form:"docs_file"`
	ImageFile   *multipart.FileHeader `form:"image_file"`
}

type SubmitChallengeResponse struct {
	ID          uuid.UUID `json:"id"`
	ChallengeID uuid.UUID `json:"challenge_id"`
	TeamID      uuid.UUID `json:"team_id"`
	Title       string    `json:"title"`
	RepoURL     *string   `json:"repo_url"`
	DocsURL     *string   `json:"docs_url"`
	ImageURL    *string   `json:"image_url"`
	SubmittedAt time.Time `json:"submitted_at"`
	Message     string    `json:"message"`
}

// Auto register team to challenge (no team_id needed)
type AutoRegisterChallengeRequest struct {
	ChallengeID uuid.UUID `json:"challenge_id" validate:"required"`
}

type AutoRegisterChallengeResponse struct {
	Message             string    `json:"message"`
	ChallengeID         uuid.UUID `json:"challenge_id"`
	TeamID              uuid.UUID `json:"team_id"`
	TeamName            string    `json:"team_name"`
	RegistrationDate    time.Time `json:"registration_date"`
	CurrentParticipants int       `json:"current_participants"`
}
