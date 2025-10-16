package experience

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Experience struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StudentProfileID uuid.UUID      `gorm:"type:uuid;not null;index" json:"student_profile_id"`
	CompanyName      string         `gorm:"size:200;not null" json:"company_name"`
	Position         string         `gorm:"size:200;not null" json:"position"`
	Department       string         `gorm:"size:200" json:"department"`
	EmploymentType   string         `gorm:"size:50;not null" json:"employment_type"`
	Location         string         `gorm:"size:200" json:"location"`
	LocationType     string         `gorm:"size:50" json:"location_type"`
	Description      string         `gorm:"type:text" json:"description"`
	Responsibilities string         `gorm:"type:text" json:"responsibilities"`
	Achievements     string         `gorm:"type:text" json:"achievements"`
	Skills           string         `gorm:"type:text" json:"skills"`
	StartDate        time.Time      `gorm:"not null" json:"start_date"`
	EndDate          *time.Time     `json:"end_date"`
	IsCurrent        bool           `gorm:"default:false" json:"is_current"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e *Experience) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (Experience) TableName() string {
	return "experiences"
}
