package student_challenge

import (
	"time"

	"github.com/google/uuid"
)

// Search student DTOs
type SearchStudentRequest struct {
	Query string `json:"query" validate:"required,min=1"`
	Limit int    `json:"limit"`
}

// Team creation DTOs
type CreateTeamRequest struct {
	TeamName           string      `json:"team_name" validate:"required,min=3,max=100"`
	About              *string     `json:"about"`
	TeamProfilePicture *string     `json:"team_profile_picture"`
	MemberIDs          []uuid.UUID `json:"member_ids" validate:"required,len=2"`
}

type TeamMemberInfo struct {
	StudentProfileID uuid.UUID `json:"student_profile_id"`
	MemberRole       *string   `json:"member_role"`
	FullName         string    `json:"full_name"`
	NIS              string    `json:"nis"`
	ProfilePicture   *string   `json:"profile_picture"`
}

type CreateTeamResponse struct {
	ID                 uuid.UUID        `json:"id"`
	TeamName           string           `json:"team_name"`
	About              *string          `json:"about"`
	TeamProfilePicture *string          `json:"team_profile_picture"`
	CreatedAt          time.Time        `json:"created_at"`
	Members            []TeamMemberInfo `json:"members"`
	MemberCount        int              `json:"member_count"`
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
	ThumbnailImage      string     `json:"thumbnail_image"`
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
}

type MyTeamResponse struct {
	ID                   uuid.UUID                 `json:"id"`
	TeamName             string                    `json:"team_name"`
	About                *string                   `json:"about"`
	TeamProfilePicture   *string                   `json:"team_profile_picture"`
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
