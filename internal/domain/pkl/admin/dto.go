package pkl

import (
	"time"

	pkl_model "github.com/Farrel44/AICademy-Backend/internal/pkl"
	"github.com/google/uuid"
)

type CreateInternshipRequest struct {
	CompanyID   uuid.UUID                `json:"company_id" validate:"required"`
	Title       string                   `json:"title" validate:"required,min=3,max=255"`
	Description string                   `json:"description" validate:"required,min=10"`
	Type        pkl_model.InternshipType `json:"type" validate:"required,oneof=PKL Job Freelance"`
	Deadline    *time.Time               `json:"deadline,omitempty"`
}

type UpdateInternshipRequest struct {
	CompanyID   *uuid.UUID                `json:"company_id,omitempty"`
	Title       *string                   `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Description *string                   `json:"description,omitempty" validate:"omitempty,min=10"`
	Type        *pkl_model.InternshipType `json:"type,omitempty" validate:"omitempty,oneof=PKL Job Freelance"`
	Deadline    *time.Time                `json:"deadline,omitempty"`
}

type PaginatedInternshipResponse struct {
	Data       []pkl_model.Internship `json:"data"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}
