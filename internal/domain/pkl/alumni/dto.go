package pkl

import (
	"time"

	"github.com/google/uuid"
)

type ApplyInternshipRequest struct {
	InternshipID uuid.UUID `json:"internship_id" validate:"required"`
}

// Clean DTOs for Alumni responses
type CleanInternshipResponse struct {
	ID               string                      `json:"id"`
	CompanyProfileID string                      `json:"company_profile_id"`
	Title            string                      `json:"title"`
	Description      string                      `json:"description"`
	Type             string                      `json:"type"`
	PostedAt         time.Time                   `json:"posted_at"`
	Deadline         *time.Time                  `json:"deadline,omitempty"`
	CompanyProfile   CleanCompanyProfileResponse `json:"company_profile"`
}

type CleanCompanyProfileResponse struct {
	ID              string            `json:"id"`
	CompanyName     string            `json:"company_name"`
	CompanyLogo     *string           `json:"company_logo,omitempty"`
	CompanyLocation *string           `json:"company_location,omitempty"`
	Description     *string           `json:"description,omitempty"`
	User            CleanUserResponse `json:"user"`
}

type CleanUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CleanPaginatedInternshipResponse struct {
	Data       []CleanInternshipResponse `json:"data"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	Limit      int                       `json:"limit"`
	TotalPages int                       `json:"total_pages"`
}

type CleanApplicationResponse struct {
	ID           string                  `json:"id"`
	InternshipID string                  `json:"internship_id"`
	Status       string                  `json:"status"`
	AppliedAt    time.Time               `json:"applied_at"`
	ReviewedAt   *time.Time              `json:"reviewed_at,omitempty"`
	Internship   CleanInternshipResponse `json:"internship"`
}

type CleanPaginatedApplicationResponse struct {
	Data       []CleanApplicationResponse `json:"data"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
	TotalPages int                        `json:"total_pages"`
}
