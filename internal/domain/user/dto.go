package user

import (
	"time"

	"github.com/google/uuid"
)

type UpdateStudentRequest struct {
	ProfilePicture *string `json:"profile_picture"`
	Bio            *string `json:"bio"`
	Headline       *string `json:"headline"`
	CvFile         *string `json:"cv_file"`
}

type RecommendedRoleInfo struct {
	RoleID          uuid.UUID `json:"role_id"`
	RoleName        string    `json:"role_name"`
	RoleDescription string    `json:"role_description"`
	RoleCategory    string    `json:"role_category"`
	Score           *float64  `json:"score,omitempty"`
	Justification   *string   `json:"justification,omitempty"`
}

// Project data for profile
type UserProjectInfo struct {
	ID          uuid.UUID `json:"id"`
	ProjectName string    `json:"project_name"`
	Description string    `json:"description"`
	LinkURL     *string   `json:"link_url"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	PhotoCount  int       `json:"photo_count"`
	IsCompleted bool      `json:"is_completed"`
}

// Certification data for profile
type UserCertificationInfo struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	IssuingOrganization string     `json:"issuing_organization"`
	IssueDate           time.Time  `json:"issue_date"`
	ExpirationDate      *time.Time `json:"expiration_date"`
	CredentialID        *string    `json:"credential_id"`
	CredentialURL       *string    `json:"credential_url"`
	IsExpired           bool       `json:"is_expired"`
	IsExpiringSoon      bool       `json:"is_expiring_soon"`
	CreatedAt           time.Time  `json:"created_at"`
}

type EnhancedStudentProfile struct {
	ID              uuid.UUID               `json:"id"`
	UserID          uuid.UUID               `json:"user_id"`
	Fullname        string                  `json:"fullname"`
	NIS             string                  `json:"nis"`
	Class           string                  `json:"class"`
	ProfilePicture  string                  `json:"profile_picture"`
	Headline        string                  `json:"headline"`
	Bio             string                  `json:"bio"`
	CVFile          *string                 `json:"cv_file"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	RecommendedRole *RecommendedRoleInfo    `json:"recommended_role,omitempty"`
	Projects        []UserProjectInfo       `json:"projects"`
	Certifications  []UserCertificationInfo `json:"certifications"`
	ProfileURL      string                  `json:"profile_url"` // aiademy.smktelkom-pwt.sch.id/NIS
}

type EnhancedAlumniProfile struct {
	ID             uuid.UUID               `json:"id"`
	UserID         uuid.UUID               `json:"user_id"`
	Fullname       string                  `json:"fullname"`
	ProfilePicture string                  `json:"profile_picture"`
	Headline       string                  `json:"headline"`
	Bio            string                  `json:"bio"`
	CVFile         *string                 `json:"cv_file"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	Projects       []UserProjectInfo       `json:"projects"`
	Certifications []UserCertificationInfo `json:"certifications"`
	ProfileURL     string                  `json:"profile_url"`
}

type EnhancedUserResponse struct {
	ID             uuid.UUID               `json:"id"`
	Email          string                  `json:"email"`
	Role           UserRole                `json:"role"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	StudentProfile *EnhancedStudentProfile `json:"student_profile,omitempty"`
	AlumniProfile  *EnhancedAlumniProfile  `json:"alumni_profile,omitempty"`
	Username       string                  `json:"username"` // for aiademy.smktelkom-pwt.sch.id/username
}

// Public profile view by NIS (LinkedIn-like)
type PublicStudentProfileResponse struct {
	NIS             string                  `json:"nis"`
	Fullname        string                  `json:"fullname"`
	Class           string                  `json:"class"`
	ProfilePicture  string                  `json:"profile_picture"`
	Headline        string                  `json:"headline"`
	Bio             string                  `json:"bio"`
	Projects        []UserProjectInfo       `json:"projects"`
	Certifications  []UserCertificationInfo `json:"certifications"`
	ProfileURL      string                  `json:"profile_url"`
	JoinedAt        time.Time               `json:"joined_at"`
	LastActive      time.Time               `json:"last_active"`
	RecommendedRole *RecommendedRoleInfo    `json:"recommended_role,omitempty"`
	// Privacy settings
	ShowEmail       bool `json:"show_email"`
	ShowCV          bool `json:"show_cv"`
	IsPublicProfile bool `json:"is_public_profile"`
}
