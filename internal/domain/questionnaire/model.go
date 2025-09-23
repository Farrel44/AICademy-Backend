package questionnaire

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestionType string

const (
	QuestionTypeMCQ    QuestionType = "mcq"
	QuestionTypeLikert QuestionType = "likert"
	QuestionTypeCase   QuestionType = "case"
	QuestionTypeText   QuestionType = "text"
)

type ProfilingQuestionnaire struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name            string    `gorm:"not null"`
	ProfilingRoleID *uuid.UUID
	Version         int    `gorm:"default:1"`
	Active          bool   `gorm:"default:false"`
	GeneratedBy     string `gorm:"default:'manual'"`
	AIPromptUsed    *string
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`

	Questions []QuestionnaireQuestion `gorm:"foreignKey:QuestionnaireID"`
}

func (p *ProfilingQuestionnaire) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type QuestionnaireQuestion struct {
	ID              uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	QuestionnaireID uuid.UUID    `gorm:"not null"`
	QuestionText    string       `gorm:"not null"`
	QuestionType    QuestionType `gorm:"not null"`
	Options         *string      `gorm:"type:text"`
	MaxScore        int          `gorm:"default:1"`
	QuestionOrder   int          `gorm:"not null"`
	Category        string       `gorm:"not null"`
	CreatedAt       time.Time    `gorm:"autoCreateTime"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime"`

	Questionnaire ProfilingQuestionnaire `gorm:"foreignKey:QuestionnaireID"`
}

func (q *QuestionnaireQuestion) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

type QuestionnaireResponse struct {
	ID                         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	StudentProfileID           uuid.UUID `gorm:"not null"`
	QuestionnaireID            uuid.UUID `gorm:"not null"`
	Answers                    string    `gorm:"type:text;not null"`
	SubmittedAt                time.Time `gorm:"not null"`
	ProcessedAt                *time.Time
	TotalScore                 *int
	AIAnalysis                 *string `gorm:"type:text"`
	AIRecommendations          *string `gorm:"type:text"`
	AIModelVersion             *string
	RecommendedProfilingRoleID *uuid.UUID
	CreatedAt                  time.Time `gorm:"autoCreateTime"`
	UpdatedAt                  time.Time `gorm:"autoUpdateTime"`

	Questionnaire ProfilingQuestionnaire `gorm:"foreignKey:QuestionnaireID"`
}

func (q *QuestionnaireResponse) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

type QuestionGenerationTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"not null"`
	Description *string
	Prompt      string    `gorm:"not null"`
	Active      bool      `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (q *QuestionGenerationTemplate) BeforeCreate(tx *gorm.DB) error {
	if q.ID == uuid.Nil {
		q.ID = uuid.New()
	}
	return nil
}

type RoleRecommendation struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoleName    string    `gorm:"not null;unique"`
	Description string    `gorm:"not null"`
	Category    string    `gorm:"not null"`
	Active      bool      `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (r *RoleRecommendation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// TargetRole represents available target roles for questionnaire generation
type TargetRole struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"not null;uniqueIndex"`
	Description string    `gorm:"not null"`
	Category    string    `gorm:"not null"` // e.g., "Technology", "Business", "Creative"
	Active      bool      `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (t *TargetRole) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
