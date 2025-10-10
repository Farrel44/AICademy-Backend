package admin_challenge

import (
	"time"

	"github.com/google/uuid"
)

type CreateChallengeRequest struct {
	ThumbnailImage  string    `json:"thumbnail_image" validate:"required"`
	Title           string    `json:"title" validate:"required"`
	Description     string    `json:"description" validate:"required"`
	Deadline        time.Time `json:"deadline" validate:"required"`
	Prize           *string   `json:"prize"`
	MaxParticipants int       `json:"max_participants" validate:"required,min=1"`
}

type UpdateChallengeRequest struct {
	ThumbnailImage  *string    `json:"thumbnail_image"`
	Title           *string    `json:"title"`
	Description     *string    `json:"description"`
	Deadline        *time.Time `json:"deadline"`
	Prize           *string    `json:"prize"`
	MaxParticipants *int       `json:"max_participants"`
}

type ScoreSubmissionRequest struct {
	SubmissionID uuid.UUID `json:"submission_id" validate:"required"`
	Points       int       `json:"points" validate:"required,min=1,max=100"`
}

type LeaderboardRequest struct {
	ChallengeID *uuid.UUID `json:"challenge_id,omitempty"`
}

type PaginatedSubmissionsResponse struct {
	Data       []SubmissionResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

type SubmissionResponse struct {
	ID            uuid.UUID `json:"id"`
	ChallengeID   uuid.UUID `json:"challenge_id"`
	ChallengeName string    `json:"challenge_name"`
	TeamName      string    `json:"team_name"`
	GitHubURL     string    `json:"github_url"`
	LiveURL       *string   `json:"live_url"`
	Description   string    `json:"description"`
	Points        *int      `json:"points"`
	SubmittedAt   time.Time `json:"submitted_at"`
}
