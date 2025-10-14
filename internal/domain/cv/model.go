package cv

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CVStatus string

const (
	CVStatusDraft     CVStatus = "draft"
	CVStatusPublished CVStatus = "published"
)

type CV struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StudentProfileID uuid.UUID      `gorm:"type:uuid;not null" json:"student_profile_id"`
	Title            string         `gorm:"size:255;not null" json:"title"`
	Status           CVStatus       `gorm:"size:50;not null;default:'draft'" json:"status"`
	Content          CVContent      `gorm:"type:jsonb" json:"content"`
	IsPublic         bool           `gorm:"default:false" json:"is_public"`
	PDFPath          string         `gorm:"size:500" json:"pdf_path,omitempty"`
	GeneratedAt      time.Time      `json:"generated_at"`
	PublishedAt      *time.Time     `json:"published_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type CVContent struct {
	PersonalInfo   PersonalInfo      `json:"personal_info"`
	Summary        string            `json:"summary"`
	Experiences    []CVExperience    `json:"experiences"`
	Projects       []CVProject       `json:"projects"`
	Skills         []CVSkill         `json:"skills"`
	Certifications []CVCertification `json:"certifications"`
	Education      CVEducation       `json:"education"`
	Languages      []CVLanguage      `json:"languages"`
	Keywords       []string          `json:"keywords"`
}

type PersonalInfo struct {
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	PersonalEmail string `json:"personal_email"`
	Phone         string `json:"phone"`
	Location      string `json:"location"`
	LinkedIn      string `json:"linkedin,omitempty"`
	GitHub        string `json:"github,omitempty"`
	Portfolio     string `json:"portfolio,omitempty"`
}

type CVExperience struct {
	ID               uuid.UUID  `json:"id"`
	CompanyName      string     `json:"company_name"`
	Position         string     `json:"position"`
	Department       string     `json:"department,omitempty"`
	EmploymentType   string     `json:"employment_type"`
	Location         string     `json:"location"`
	LocationType     string     `json:"location_type"`
	Description      string     `json:"description"`
	Responsibilities []string   `json:"responsibilities"`
	Achievements     []string   `json:"achievements"`
	Skills           []string   `json:"skills"`
	StartDate        time.Time  `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	IsCurrent        bool       `json:"is_current"`
}

type CVLanguage struct {
	Name      string `json:"name"`
	Level     string `json:"level"`
	Certified bool   `json:"certified"`
}

type CVSkill struct {
	Name     string `json:"name"`
	Level    string `json:"level"`
	Category string `json:"category"`
}

type CVProject struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Role         string     `json:"role"`
	Technologies []string   `json:"technologies"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	URL          string     `json:"url,omitempty"`
	TechStack    string     `json:"tech_stack,omitempty"`
	Highlights   []string   `json:"highlights"`
}

type CVCertification struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	IssuingOrganization string     `json:"issuing_organization"`
	IssueDate           time.Time  `json:"issue_date"`
	ExpirationDate      *time.Time `json:"expiration_date,omitempty"`
	CredentialID        string     `json:"credential_id,omitempty"`
	CredentialURL       string     `json:"credential_url,omitempty"`
}

type CVEducation struct {
	School       string   `json:"school"`
	Degree       string   `json:"degree"`
	Major        string   `json:"major"`
	StartYear    int      `json:"start_year"`
	EndYear      *int     `json:"end_year,omitempty"`
	GPA          string   `json:"gpa,omitempty"`
	Achievements []string `json:"achievements,omitempty"`
}

type ATSScore struct {
	Overall     int      `json:"overall"`
	Keywords    int      `json:"keywords"`
	Format      int      `json:"format"`
	Structure   int      `json:"structure"`
	Suggestions []string `json:"suggestions"`
}

func (c *CV) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (c *CVContent) Scan(value interface{}) error {
	if value == nil {
		*c = CVContent{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-byte value into CVContent")
	}

	return json.Unmarshal(bytes, c)
}

func (c CVContent) Value() (driver.Value, error) {
	return json.Marshal(c)
}
