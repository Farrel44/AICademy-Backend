package pkl

import (
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/google/uuid"
)

type CreateInternshipRequest struct {
	CompanyID   uuid.UUID          `json:"company_id" validate:"required"`
	Title       string             `json:"title" validate:"required,min=3,max=255"`
	Description string             `json:"description" validate:"required,min=10"`
	Type        pkl.InternshipType `json:"type" validate:"required,oneof=PKL Job Freelance"`
	Deadline    *time.Time         `json:"deadline,omitempty"`
}

type UpdateInternshipRequest struct {
	CompanyID   *uuid.UUID          `json:"company_id,omitempty"`
	Title       *string             `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Description *string             `json:"description,omitempty" validate:"omitempty,min=10"`
	Type        *pkl.InternshipType `json:"type,omitempty" validate:"omitempty,oneof=PKL Job Freelance"`
	Deadline    *time.Time          `json:"deadline,omitempty"`
}

type PaginatedInternshipResponse struct {
	Data       []pkl.Internship `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type InternshipResponse struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	PostedAt    time.Time      `json:"posted_at"`
	Deadline    *time.Time     `json:"deadline,omitempty"`
	Company     CompanySummary `json:"company"`
}

type CompanySummary struct {
	Name        string  `json:"company_name"`
	Logo        *string `json:"company_logo,omitempty"`
	Location    *string `json:"company_location,omitempty"`
	Description *string `json:"description,omitempty"`
	Email       string  `json:"email"`
}

type SubmissionResponse struct {
	ID               string             `json:"id"`
	InternshipID     string             `json:"internship_id"`
	StudentProfileID string             `json:"student_profile_id"`
	Status           string             `json:"status"`
	AppliedAt        time.Time          `json:"applied_at"`
	ReviewedAt       *time.Time         `json:"reviewed_at"`
	ApprovedByUserID *string            `json:"approved_by_user_id,omitempty"`
	ApprovedByRole   *string            `json:"approved_by_role,omitempty"`
	ApproverEmail    *string            `json:"approver_email,omitempty"`
	Student          StudentSummary     `json:"student"`
	Internship       *InternshipSummary `json:"internship,omitempty"`
}

type StudentSummary struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Fullname       string  `json:"fullname"`
	NIS            string  `json:"nis"`
	Class          string  `json:"class"`
	ProfilePicture *string `json:"profile_picture,omitempty"`
	Headline       *string `json:"headline,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	CvFile         *string `json:"cv_file,omitempty"`
}

type InternshipSummary struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Type     string         `json:"type"`
	PostedAt time.Time      `json:"posted_at"`
	Deadline *time.Time     `json:"deadline,omitempty"`
	Company  CompanySummary `json:"company"`
}

type InternshipWithSubmissionsResponse struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	Description     string               `json:"description"`
	Type            string               `json:"type"`
	PostedAt        time.Time            `json:"posted_at"`
	Deadline        *time.Time           `json:"deadline,omitempty"`
	Company         CompanySummary       `json:"company"`
	Submissions     []SubmissionResponse `json:"submissions"`
	SubmissionCount int                  `json:"submission_count"`
}

type UpdateSubmissionStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}

type PaginatedSubmissionResponse struct {
	Data       []SubmissionResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

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
	Photos          *string           `json:"photos,omitempty"`
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
