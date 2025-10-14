package experience

import (
	"time"

	"github.com/google/uuid"
)

type CreateExperienceRequest struct {
	CompanyName      string     `json:"company_name" validate:"required,min=2,max=200"`
	Position         string     `json:"position" validate:"required,min=2,max=200"`
	Department       string     `json:"department" validate:"omitempty,max=200"`
	EmploymentType   string     `json:"employment_type" validate:"required,oneof=Full-time Part-time Internship Freelance Contract"`
	Location         string     `json:"location" validate:"omitempty,max=200"`
	LocationType     string     `json:"location_type" validate:"omitempty,oneof=On-site Remote Hybrid"`
	Description      string     `json:"description" validate:"omitempty,max=2000"`
	Responsibilities string     `json:"responsibilities" validate:"omitempty,max=2000"`
	Achievements     string     `json:"achievements" validate:"omitempty,max=2000"`
	Skills           string     `json:"skills" validate:"omitempty,max=500"`
	StartDate        time.Time  `json:"start_date" validate:"required"`
	EndDate          *time.Time `json:"end_date"`
	IsCurrent        bool       `json:"is_current"`
}

type UpdateExperienceRequest struct {
	CompanyName      *string    `json:"company_name" validate:"omitempty,min=2,max=200"`
	Position         *string    `json:"position" validate:"omitempty,min=2,max=200"`
	Department       *string    `json:"department" validate:"omitempty,max=200"`
	EmploymentType   *string    `json:"employment_type" validate:"omitempty,oneof=Full-time Part-time Internship Freelance Contract"`
	Location         *string    `json:"location" validate:"omitempty,max=200"`
	LocationType     *string    `json:"location_type" validate:"omitempty,oneof=On-site Remote Hybrid"`
	Description      *string    `json:"description" validate:"omitempty,max=2000"`
	Responsibilities *string    `json:"responsibilities" validate:"omitempty,max=2000"`
	Achievements     *string    `json:"achievements" validate:"omitempty,max=2000"`
	Skills           *string    `json:"skills" validate:"omitempty,max=500"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	IsCurrent        *bool      `json:"is_current"`
}

type ExperienceResponse struct {
	ID               uuid.UUID  `json:"id"`
	StudentProfileID uuid.UUID  `json:"student_profile_id"`
	CompanyName      string     `json:"company_name"`
	Position         string     `json:"position"`
	Department       string     `json:"department"`
	EmploymentType   string     `json:"employment_type"`
	Location         string     `json:"location"`
	LocationType     string     `json:"location_type"`
	Description      string     `json:"description"`
	Responsibilities string     `json:"responsibilities"`
	Achievements     string     `json:"achievements"`
	Skills           string     `json:"skills"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	IsCurrent        bool       `json:"is_current"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ExperienceListResponse struct {
	Data       []ExperienceResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}
