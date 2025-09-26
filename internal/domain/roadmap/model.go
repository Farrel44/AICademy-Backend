package roadmap

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoadmapVisibility enum for roadmap visibility
type RoadmapVisibility string

const (
	RoadmapVisibilityPrivate RoadmapVisibility = "private"
	RoadmapVisibilitySchool  RoadmapVisibility = "school"
	RoadmapVisibilityPublic  RoadmapVisibility = "public"
)

// RoadmapStatus enum for roadmap status
type RoadmapStatus string

const (
	RoadmapStatusDraft    RoadmapStatus = "draft"
	RoadmapStatusActive   RoadmapStatus = "active"
	RoadmapStatusArchived RoadmapStatus = "archived"
)

// RoadmapProgressStatus enum for roadmap progress status
type RoadmapProgressStatus string

const (
	RoadmapProgressStatusLocked     RoadmapProgressStatus = "locked"
	RoadmapProgressStatusUnlocked   RoadmapProgressStatus = "unlocked"
	RoadmapProgressStatusInProgress RoadmapProgressStatus = "in_progress"
	RoadmapProgressStatusCompleted  RoadmapProgressStatus = "completed"
)

// FeatureRoadmap represents a personalized roadmap for a student
type FeatureRoadmap struct {
	ID               uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StudentProfileID *uuid.UUID        `json:"student_profile_id" gorm:"type:uuid"`
	ProfilingRoleID  *uuid.UUID        `json:"profiling_role_id" gorm:"type:uuid"`
	RoadmapName      string            `json:"roadmap_name" gorm:"not null"`
	Description      *string           `json:"description" gorm:"type:text"`
	Visibility       RoadmapVisibility `json:"visibility" gorm:"type:varchar(20);default:'private'"`
	Status           RoadmapStatus     `json:"status" gorm:"type:varchar(20);default:'draft'"`

	// AI Generation metadata
	AIGeneratedBy           *string    `json:"ai_generated_by" gorm:"type:varchar(100)"`   // AI model version used
	AIPromptUsed            *string    `json:"ai_prompt_used" gorm:"type:text"`            // Prompt used for generation
	QuestionnaireResponseID *uuid.UUID `json:"questionnaire_response_id" gorm:"type:uuid"` // Reference to questionnaire that triggered this roadmap

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	StudentProfile interface{}       `json:"student_profile,omitempty" gorm:"foreignKey:StudentProfileID"`
	ProfilingRole  interface{}       `json:"profiling_role,omitempty" gorm:"foreignKey:ProfilingRoleID"`
	Steps          []RoadmapStep     `json:"steps,omitempty" gorm:"foreignKey:RoadmapID"`
	Progress       []RoadmapProgress `json:"progress,omitempty" gorm:"foreignKey:RoadmapID"`
}

// RoadmapStep represents individual steps/modules in a roadmap
type RoadmapStep struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoadmapID        uuid.UUID `json:"roadmap_id" gorm:"type:uuid;not null"`
	StepOrder        int       `json:"step_order" gorm:"not null"`
	Title            string    `json:"title" gorm:"not null"`
	ShortDescription *string   `json:"short_description" gorm:"type:text"`
	UnlockCondition  *string   `json:"unlock_condition" gorm:"type:text"`

	// AI Generated content
	AIGeneratedContent   *string `json:"ai_generated_content" gorm:"type:text"`    // Full step description from AI
	ExpectedOutcome      *string `json:"expected_outcome" gorm:"type:text"`        // What student should achieve
	SubmissionGuidelines *string `json:"submission_guidelines" gorm:"type:text"`   // How to submit evidence
	ResourceLinks        *string `json:"resource_links" gorm:"type:text"`          // JSON array of learning resources
	EstimatedDuration    *int    `json:"estimated_duration"`                       // In hours
	DifficultyLevel      *string `json:"difficulty_level" gorm:"type:varchar(20)"` // beginner, intermediate, advanced

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Roadmap  FeatureRoadmap    `json:"roadmap,omitempty" gorm:"foreignKey:RoadmapID"`
	Progress []RoadmapProgress `json:"progress,omitempty" gorm:"foreignKey:RoadmapStepID"`
}

// RoadmapProgress tracks student progress on each roadmap step
type RoadmapProgress struct {
	ID               uuid.UUID             `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoadmapID        uuid.UUID             `json:"roadmap_id" gorm:"type:uuid;not null"`
	RoadmapStepID    uuid.UUID             `json:"roadmap_step_id" gorm:"type:uuid;not null"`
	StudentProfileID uuid.UUID             `json:"student_profile_id" gorm:"type:uuid;not null"`
	Status           RoadmapProgressStatus `json:"status" gorm:"type:varchar(20);default:'locked'"`

	// Submission & Validation
	EvidenceLink                *string    `json:"evidence_link" gorm:"type:text"`                   // Student submission
	EvidenceType                *string    `json:"evidence_type" gorm:"type:varchar(50)"`            // file, url, text
	SubmissionNotes             *string    `json:"submission_notes" gorm:"type:text"`                // Student notes
	ValidatedByTeacherProfileID *uuid.UUID `json:"validated_by_teacher_profile_id" gorm:"type:uuid"` // Teacher who validated
	ValidationNotes             *string    `json:"validation_notes" gorm:"type:text"`                // Teacher feedback
	ValidationScore             *int       `json:"validation_score"`                                 // 1-100 score from teacher

	// Timestamps
	StartedAt   *time.Time `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
	CompletedAt *time.Time `json:"completed_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Roadmap            FeatureRoadmap `json:"roadmap,omitempty" gorm:"foreignKey:RoadmapID"`
	RoadmapStep        RoadmapStep    `json:"roadmap_step,omitempty" gorm:"foreignKey:RoadmapStepID"`
	StudentProfile     interface{}    `json:"student_profile,omitempty" gorm:"foreignKey:StudentProfileID"`
	ValidatedByTeacher interface{}    `json:"validated_by_teacher,omitempty" gorm:"foreignKey:ValidatedByTeacherProfileID"`
}

// RoadmapTemplate represents AI-generated roadmap templates for different roles
type RoadmapTemplate struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProfilingRoleID uuid.UUID `json:"profiling_role_id" gorm:"type:uuid;not null"`
	TemplateName    string    `json:"template_name" gorm:"not null"`
	Description     *string   `json:"description" gorm:"type:text"`

	// AI Generation metadata
	AIGeneratedBy string `json:"ai_generated_by" gorm:"not null"`          // AI model version
	AIPromptUsed  string `json:"ai_prompt_used" gorm:"type:text;not null"` // Prompt used
	VersionNumber int    `json:"version_number" gorm:"default:1"`          // Template version
	IsActive      bool   `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	ProfilingRole interface{}           `json:"profiling_role,omitempty" gorm:"foreignKey:ProfilingRoleID"`
	TemplateSteps []RoadmapTemplateStep `json:"template_steps,omitempty" gorm:"foreignKey:TemplateID"`
}

// RoadmapTemplateStep represents template steps that can be used to generate personal roadmaps
type RoadmapTemplateStep struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TemplateID           uuid.UUID `json:"template_id" gorm:"type:uuid;not null"`
	StepOrder            int       `json:"step_order" gorm:"not null"`
	Title                string    `json:"title" gorm:"not null"`
	ShortDescription     *string   `json:"short_description" gorm:"type:text"`
	AIGeneratedContent   *string   `json:"ai_generated_content" gorm:"type:text"`
	ExpectedOutcome      *string   `json:"expected_outcome" gorm:"type:text"`
	SubmissionGuidelines *string   `json:"submission_guidelines" gorm:"type:text"`
	ResourceLinks        *string   `json:"resource_links" gorm:"type:text"`
	EstimatedDuration    *int      `json:"estimated_duration"`
	DifficultyLevel      *string   `json:"difficulty_level" gorm:"type:varchar(20)"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Template RoadmapTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
}

// BeforeCreate hooks for UUID generation
func (r *FeatureRoadmap) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (rs *RoadmapStep) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}

func (rp *RoadmapProgress) BeforeCreate(tx *gorm.DB) error {
	if rp.ID == uuid.Nil {
		rp.ID = uuid.New()
	}
	return nil
}

func (rt *RoadmapTemplate) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (rts *RoadmapTemplateStep) BeforeCreate(tx *gorm.DB) error {
	if rts.ID == uuid.Nil {
		rts.ID = uuid.New()
	}
	return nil
}
