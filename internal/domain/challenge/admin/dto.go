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
