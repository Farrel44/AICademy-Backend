package project

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// Project DTOs
type CreateProjectRequest struct {
	ProjectName string                  `json:"project_name" validate:"required"`
	Description string                  `json:"description" validate:"required"`
	LinkURL     *string                 `json:"link_url"`
	StartDate   time.Time               `json:"start_date" validate:"required"`
	EndDate     time.Time               `json:"end_date" validate:"required"`
	Photos      []*multipart.FileHeader `form:"photos"`
}

type UpdateProjectRequest struct {
	ProjectName string                  `json:"project_name"`
	Description string                  `json:"description"`
	LinkURL     *string                 `json:"link_url"`
	StartDate   *time.Time              `json:"start_date"`
	EndDate     *time.Time              `json:"end_date"`
	Photos      []*multipart.FileHeader `form:"photos"`
}

type AddProjectContributorRequest struct {
	StudentProfileID uuid.UUID  `json:"student_profile_id" validate:"required"`
	ProjectRole      *string    `json:"project_role"`
	ProfilingRoleID  *uuid.UUID `json:"profiling_role_id"`
	Description      *string    `json:"description"`
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
	ID                    uuid.UUID            `json:"id"`
	OwnerStudentProfileID uuid.UUID            `json:"owner_student_profile_id"`
	ProjectName           string               `json:"project_name"`
	Description           string               `json:"description"`
	LinkURL               *string              `json:"link_url"`
	StartDate             time.Time            `json:"start_date"`
	EndDate               time.Time            `json:"end_date"`
	CreatedAt             time.Time            `json:"created_at"`
	Contributors          []ProjectContributor `json:"contributors,omitempty"`
	Photos                []ProjectPhoto       `json:"photos,omitempty"`
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
