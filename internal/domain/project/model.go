package project

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerStudentProfileID uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_student_profile_id"`
	ProjectName           string    `gorm:"not null" json:"project_name"`
	Description           string    `gorm:"type:text" json:"description"`
	LinkURL               *string   `gorm:"type:text" json:"link_url"`
	StartDate             time.Time `gorm:"type:date" json:"start_date"`
	EndDate               time.Time `gorm:"type:date" json:"end_date"`
	CreatedAt             time.Time `gorm:"type:timestamptz" json:"created_at"`

	Contributors []ProjectContributor `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"contributors,omitempty"`
	Photos       []ProjectPhoto       `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"photos,omitempty"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ProjectContributor struct {
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"project_id"`
	StudentProfileID uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"student_profile_id"`
	ProjectRole      *string    `gorm:"type:varchar" json:"project_role"`
	ProfilingRoleID  *uuid.UUID `gorm:"type:uuid;index" json:"profiling_role_id"`
	Description      *string    `gorm:"type:text" json:"description"`

	Project    *Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	TargetRole *TargetRole `gorm:"foreignKey:ProfilingRoleID;references:ID" json:"target_role,omitempty"`
}

func (ProjectContributor) TableName() string {
	return "project_contributors"
}

type TargetRole struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`
	Description string    `gorm:"not null" json:"description"`
	Category    string    `gorm:"not null" json:"category"` // e.g., "Technology", "Business", "Creative"
	Active      bool      `gorm:"default:true" json:"active"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Remove the circular reference to challenge.TeamMember
	ProjectContributors []ProjectContributor `gorm:"foreignKey:ProfilingRoleID;constraint:OnDelete:SET NULL" json:"project_contributors,omitempty"`
}

func (t *TargetRole) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type ProjectPhoto struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	URL       string    `gorm:"type:text;not null" json:"url"`
	Caption   *string   `gorm:"type:text" json:"caption"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time `gorm:"type:timestamptz" json:"created_at"`

	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (p *ProjectPhoto) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (ProjectPhoto) TableName() string {
	return "project_photos"
}

type Certification struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	StudentProfileID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"student_profile_id"`
	Name                string     `gorm:"not null" json:"name"`
	IssuingOrganization string     `gorm:"not null" json:"issuing_organization"`
	IssueDate           time.Time  `gorm:"type:date" json:"issue_date"`
	ExpirationDate      *time.Time `gorm:"type:date" json:"expiration_date"`
	CredentialID        *string    `json:"credential_id"`
	CredentialURL       *string    `gorm:"type:text" json:"credential_url"`
	CreatedAt           time.Time  `gorm:"type:timestamptz" json:"created_at"`

	Photos []CertificationPhoto `gorm:"foreignKey:CertificationID;constraint:OnDelete:CASCADE" json:"photos,omitempty"`
}

func (c *Certification) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c *Certification) IsExpired() bool {
	if c.ExpirationDate == nil {
		return false
	}
	return time.Now().After(*c.ExpirationDate)
}

func (c *Certification) IsExpiringSoon() bool {
	if c.ExpirationDate == nil {
		return false
	}
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
	return c.ExpirationDate.Before(thirtyDaysFromNow) && !c.IsExpired()
}

type CertificationPhoto struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CertificationID uuid.UUID `gorm:"type:uuid;not null;index" json:"certification_id"`
	URL             string    `gorm:"type:text;not null" json:"url"`
	Caption         *string   `gorm:"type:text" json:"caption"`
	IsPrimary       bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt       time.Time `gorm:"type:timestamptz" json:"created_at"`

	Certification *Certification `gorm:"foreignKey:CertificationID" json:"certification,omitempty"`
}

func (c *CertificationPhoto) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (CertificationPhoto) TableName() string {
	return "certification_photos"
}
