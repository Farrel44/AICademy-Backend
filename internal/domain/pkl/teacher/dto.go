package pkl

import (
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
)

type UpdateApplicationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}

type UpdateSubmissionStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}

type PaginatedInternshipResponse struct {
	Data       []pkl.Internship `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type PaginatedApplicationResponse struct {
	Data       []SubmissionResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
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

type CompanySummary struct {
	Name        string  `json:"company_name"`
	Logo        *string `json:"company_logo,omitempty"`
	Location    *string `json:"company_location,omitempty"`
	Description *string `json:"description,omitempty"`
	Email       string  `json:"email"`
}
