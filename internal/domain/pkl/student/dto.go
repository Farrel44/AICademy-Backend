package pkl

import "github.com/google/uuid"

type ApplyInternshipRequest struct {
	InternshipID uuid.UUID `json:"internship_id" validate:"required"`
}
