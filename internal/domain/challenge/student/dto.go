package student_challenge

import (
	"time"

	"github.com/google/uuid"
)

// Search student DTOs
type SearchStudentRequest struct {
	Query string `json:"query" validate:"required,min=2"` // NIS, name, or email
	Limit int    `json:"limit" validate:"min=1,max=20"`   // Default 10
}

type StudentSearchResult struct {
	ID           uuid.UUID `json:"id"`
	NIS          string    `json:"nis"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	ProfilePhoto *string   `json:"profile_photo"`
	Class        *string   `json:"class"`
}

// Team creation DTOs
type CreateTeamRequest struct {
	TeamName           string      `json:"team_name" validate:"required,min=3,max=50"`
	About              *string     `json:"about" validate:"max=500"`
	TeamProfilePicture *string     `json:"team_profile_picture"`
	MemberIDs          []uuid.UUID `json:"member_ids" validate:"required,len=2"` // 2 other members (creator is automatic)
}

type TeamMemberInfo struct {
	StudentProfileID uuid.UUID  `json:"student_profile_id"`
	MemberRole       *string    `json:"member_role"`
	ProfilingRoleID  *uuid.UUID `json:"profiling_role_id"`
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
	ID                  uuid.UUID `json:"id"`
	ThumbnailImage      string    `json:"thumbnail_image"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	Deadline            time.Time `json:"deadline"`
	Prize               *string   `json:"prize"`
	MaxParticipants     int       `json:"max_participants"`
	CurrentParticipants int       `json:"current_participants"`
	CanRegister         bool      `json:"can_register"`
	IsActive            bool      `json:"is_active"`
	OrganizerName       string    `json:"organizer_name"`
}
