package roadmap

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Farrel44/AICademy-Backend/internal/domain/project"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
)

type RoadmapStatus string

const (
	RoadmapStatusDraft    RoadmapStatus = "draft"
	RoadmapStatusActive   RoadmapStatus = "active"
	RoadmapStatusArchived RoadmapStatus = "archived"
)

type RoadmapProgressStatus string

const (
	RoadmapProgressStatusLocked     RoadmapProgressStatus = "locked"
	RoadmapProgressStatusUnlocked   RoadmapProgressStatus = "unlocked"
	RoadmapProgressStatusInProgress RoadmapProgressStatus = "in_progress"
	RoadmapProgressStatusSubmitted  RoadmapProgressStatus = "submitted"
	RoadmapProgressStatusApproved   RoadmapProgressStatus = "approved"
	RoadmapProgressStatusRejected   RoadmapProgressStatus = "rejected"
	RoadmapProgressStatusCompleted  RoadmapProgressStatus = "completed"
)

// FeatureRoadmap - Template roadmap yang dibuat admin untuk role tertentu
type FeatureRoadmap struct {
	ID              uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProfilingRoleID uuid.UUID     `json:"profiling_role_id" gorm:"type:uuid;not null"`
	RoadmapName     string        `json:"roadmap_name" gorm:"not null"`
	Description     *string       `json:"description" gorm:"type:text"`
	Status          RoadmapStatus `json:"status" gorm:"type:varchar(20);default:'draft'"`
	CreatedBy       uuid.UUID     `json:"created_by" gorm:"type:uuid;not null"` // Admin yang membuat

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	ProfilingRole   *project.TargetRole      `json:"profiling_role,omitempty" gorm:"foreignKey:ProfilingRoleID"`
	CreatedByUser   *user.User               `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	Steps           []RoadmapStep            `json:"steps,omitempty" gorm:"foreignKey:RoadmapID;constraint:OnDelete:CASCADE"`
	StudentProgress []StudentRoadmapProgress `json:"student_progress,omitempty" gorm:"foreignKey:RoadmapID"`
}

// RoadmapStep - Step dalam roadmap template
type RoadmapStep struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoadmapID   uuid.UUID `json:"roadmap_id" gorm:"type:uuid;not null"`
	StepOrder   int       `json:"step_order" gorm:"not null"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text;not null"`

	// Learning objectives dan requirements
	LearningObjectives   string  `json:"learning_objectives" gorm:"type:text;not null"`
	SubmissionGuidelines string  `json:"submission_guidelines" gorm:"type:text;not null"`
	ResourceLinks        *string `json:"resource_links" gorm:"type:text"`                   // JSON array of resources
	EstimatedDuration    int     `json:"estimated_duration" gorm:"not null"`                // In hours
	DifficultyLevel      string  `json:"difficulty_level" gorm:"type:varchar(20);not null"` // beginner, intermediate, advanced

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Roadmap         FeatureRoadmap        `json:"roadmap,omitempty" gorm:"foreignKey:RoadmapID"`
	StudentProgress []StudentStepProgress `json:"student_progress,omitempty" gorm:"foreignKey:RoadmapStepID"`
}

// StudentRoadmapProgress - Progress siswa pada roadmap tertentu
type StudentRoadmapProgress struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoadmapID        uuid.UUID `json:"roadmap_id" gorm:"type:uuid;not null"`
	StudentProfileID uuid.UUID `json:"student_profile_id" gorm:"type:uuid;not null"`

	// Progress tracking
	TotalSteps      int     `json:"total_steps" gorm:"not null"`
	CompletedSteps  int     `json:"completed_steps" gorm:"default:0"`
	ProgressPercent float64 `json:"progress_percent" gorm:"default:0"`

	StartedAt      *time.Time `json:"started_at"`
	LastActivityAt *time.Time `json:"last_activity_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Roadmap        FeatureRoadmap        `json:"roadmap,omitempty" gorm:"foreignKey:RoadmapID"`
	StudentProfile *user.StudentProfile  `json:"student_profile,omitempty" gorm:"foreignKey:StudentProfileID"`
	StepProgress   []StudentStepProgress `json:"step_progress,omitempty" gorm:"foreignKey:StudentRoadmapProgressID"`
}

// StudentStepProgress - Progress siswa pada step tertentu
type StudentStepProgress struct {
	ID                       uuid.UUID             `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StudentRoadmapProgressID uuid.UUID             `json:"student_roadmap_progress_id" gorm:"type:uuid;not null"`
	RoadmapStepID            uuid.UUID             `json:"roadmap_step_id" gorm:"type:uuid;not null"`
	Status                   RoadmapProgressStatus `json:"status" gorm:"type:varchar(20);default:'locked'"`

	// Submission data
	EvidenceLink    *string `json:"evidence_link" gorm:"type:text"`
	EvidenceType    *string `json:"evidence_type" gorm:"type:varchar(50)"` // url, file, text
	SubmissionNotes *string `json:"submission_notes" gorm:"type:text"`

	// Teacher validation
	ValidatedByTeacherID *uuid.UUID `json:"validated_by_teacher_id" gorm:"type:uuid"`
	ValidationNotes      *string    `json:"validation_notes" gorm:"type:text"`
	ValidationScore      *int       `json:"validation_score"` // 0-100

	// Timestamps
	StartedAt   *time.Time `json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
	CompletedAt *time.Time `json:"completed_at"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	StudentRoadmapProgress StudentRoadmapProgress `json:"student_roadmap_progress,omitempty" gorm:"foreignKey:StudentRoadmapProgressID"`
	RoadmapStep            RoadmapStep            `json:"roadmap_step,omitempty" gorm:"foreignKey:RoadmapStepID"`
	ValidatedByTeacher     *user.TeacherProfile   `json:"validated_by_teacher,omitempty" gorm:"foreignKey:ValidatedByTeacherID"`
}

// Unique constraint untuk student progress
func (StudentRoadmapProgress) TableName() string {
	return "student_roadmap_progress"
}

func (StudentStepProgress) TableName() string {
	return "student_step_progress"
}

// BeforeCreate hooks
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

func (srp *StudentRoadmapProgress) BeforeCreate(tx *gorm.DB) error {
	if srp.ID == uuid.Nil {
		srp.ID = uuid.New()
	}
	return nil
}

func (ssp *StudentStepProgress) BeforeCreate(tx *gorm.DB) error {
	if ssp.ID == uuid.Nil {
		ssp.ID = uuid.New()
	}
	return nil
}
