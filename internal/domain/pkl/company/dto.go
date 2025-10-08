package pkl

import (
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
)

type CreateInternshipRequest struct {
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description" validate:"required"`
	Type        pkl.InternshipType `json:"type" validate:"required"`
	Deadline    *time.Time         `json:"deadline"`
}

type UpdateInternshipRequest struct {
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description" validate:"required"`
	Type        pkl.InternshipType `json:"type" validate:"required"`
	Deadline    *time.Time         `json:"deadline"`
}

type UpdateApplicationStatusRequest struct {
	Status pkl.ApplicationStatus `json:"status" validate:"required"`
}
