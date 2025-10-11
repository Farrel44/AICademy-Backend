package project

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// Project DTOs
type CreateProjectRequest struct {
	ProjectName  string                     `json:"project_name" validate:"required"`
	Description  string                     `json:"description" validate:"required"`
	LinkURL      *string                    `json:"link_url"`
	StartDate    time.Time                  `json:"start_date" validate:"required"`
	EndDate      time.Time                  `json:"end_date" validate:"required"`
	Photos       []*multipart.FileHeader    `form:"photos"`
	Contributors []CreateContributorRequest `json:"contributors"`
}

type UpdateProjectRequest struct {
	ProjectName string                  `json:"project_name"`
	Description string                  `json:"description"`
	LinkURL     *string                 `json:"link_url"`
	StartDate   *time.Time              `json:"start_date"`
	EndDate     *time.Time              `json:"end_date"`
	Photos      []*multipart.FileHeader `form:"photos"`
}

type CreateContributorRequest struct {
	StudentID string    `json:"student_id" validate:"required"` // NIS atau identifier lain
	RoleID    uuid.UUID `json:"role_id" validate:"required"`    // Target role ID
}

type AddProjectContributorRequest struct {
	StudentID string    `json:"student_id" validate:"required"` // NIS atau identifier lain
	RoleID    uuid.UUID `json:"role_id" validate:"required"`    // Target role ID
}

// Certification DTOs
type CreateCertificationRequest struct {
	Name                string                  `json:"name" validate:"required"`
	IssuingOrganization string                  `json:"issuing_organization" validate:"required"`
	IssueDate           time.Time               `json:"issue_date" validate:"required"`
	ExpirationDate      *time.Time              `json:"expiration_date"`
	CredentialID        *string                 `json:"credential_id"`
	CredentialURL       *string                 `json:"credential_url"`
	Media               *string                 `json:"media"`
	Photos              []*multipart.FileHeader `form:"photos"`
}

type UpdateCertificationRequest struct {
	Name                *string                 `json:"name"`
	IssuingOrganization *string                 `json:"issuing_organization"`
	IssueDate           *time.Time              `json:"issue_date"`
	ExpirationDate      *time.Time              `json:"expiration_date"`
	CredentialID        *string                 `json:"credential_id"`
	CredentialURL       *string                 `json:"credential_url"`
	Media               *string                 `json:"media"`
	Photos              []*multipart.FileHeader `form:"photos"`
}

// Response DTOs
type ProjectResponse struct {
	ID                    uuid.UUID                    `json:"id"`
	OwnerStudentProfileID uuid.UUID                    `json:"owner_student_profile_id"`
	ProjectName           string                       `json:"project_name"`
	Description           string                       `json:"description"`
	LinkURL               *string                      `json:"link_url"`
	StartDate             time.Time                    `json:"start_date"`
	EndDate               time.Time                    `json:"end_date"`
	CreatedAt             time.Time                    `json:"created_at"`
	Contributors          []ProjectContributorResponse `json:"contributors,omitempty"`
	Photos                []ProjectPhoto               `json:"photos,omitempty"`
}

type ProjectContributorResponse struct {
	StudentProfileID uuid.UUID           `json:"student_profile_id"`
	StudentName      string              `json:"student_name"`
	StudentNIS       string              `json:"student_nis"`
	StudentClass     string              `json:"student_class"`
	StudentEmail     string              `json:"student_email"`
	RoleID           uuid.UUID           `json:"role_id"`
	Role             *TargetRoleResponse `json:"role,omitempty"`
}

type TargetRoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
}

type CertificationResponse struct {
	ID                  uuid.UUID            `json:"id"`
	StudentProfileID    uuid.UUID            `json:"student_profile_id"`
	Name                string               `json:"name"`
	IssuingOrganization string               `json:"issuing_organization"`
	IssueDate           time.Time            `json:"issue_date"`
	ExpirationDate      *time.Time           `json:"expiration_date"`
	CredentialID        *string              `json:"credential_id"`
	CredentialURL       *string              `json:"credential_url"`
	Media               *string              `json:"media"`
	CreatedAt           time.Time            `json:"created_at"`
	IsExpired           bool                 `json:"is_expired"`
	IsExpiringSoon      bool                 `json:"is_expiring_soon"`
	Photos              []CertificationPhoto `json:"photos,omitempty"`
}
