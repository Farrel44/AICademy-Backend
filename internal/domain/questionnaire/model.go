package questionnaire

import (
	"time"

	"github.com/google/uuid"
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

type QuestionnaireQuestion struct {
	ID              uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	QuestionnaireID uuid.UUID    `gorm:"not null"`
	QuestionText    string       `gorm:"not null"`
	QuestionType    QuestionType `gorm:"not null"`
	Options         *string
	MaxScore        int       `gorm:"default:1"`
	QuestionOrder   int       `gorm:"not null"`
	Category        string    `gorm:"not null"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`

	Questionnaire ProfilingQuestionnaire `gorm:"foreignKey:QuestionnaireID"`
}

type QuestionnaireResponse struct {
	ID                         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	StudentProfileID           uuid.UUID `gorm:"not null"`
	QuestionnaireID            uuid.UUID `gorm:"not null"`
	Answers                    string    `gorm:"type:jsonb;not null"`
	SubmittedAt                time.Time `gorm:"autoCreateTime"`
	ProcessedAt                *time.Time
	TotalScore                 *int
	AIAnalysis                 *string `gorm:"type:jsonb"`
	AIRecommendations          *string `gorm:"type:jsonb"`
	AIModelVersion             *string
	RecommendedProfilingRoleID *uuid.UUID
	CreatedAt                  time.Time `gorm:"autoCreateTime"`
	UpdatedAt                  time.Time `gorm:"autoUpdateTime"`

	Questionnaire ProfilingQuestionnaire `gorm:"foreignKey:QuestionnaireID"`
}

type RoleRecommendation struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ResponseID      uuid.UUID `gorm:"not null"`
	ProfilingRoleID uuid.UUID `gorm:"not null"`
	Rank            int       `gorm:"not null"`
	Score           float64   `gorm:"not null"`
	Justification   *string
	CreatedAt       time.Time `gorm:"autoCreateTime"`

	Response QuestionnaireResponse `gorm:"foreignKey:ResponseID"`
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
