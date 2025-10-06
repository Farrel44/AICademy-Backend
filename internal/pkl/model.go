package pkl_model

import (
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InternshipType enum for internship types
type InternshipType string

const (
	InternshipTypePKL       InternshipType = "PKL"
	InternshipTypeJob       InternshipType = "Job"
	InternshipTypeFreelance InternshipType = "Freelance"
)

// ApplicationStatus enum for application status
type ApplicationStatus string

const (
	ApplicationStatusPending           ApplicationStatus = "pending"
	ApplicationStatusApprovedByTeacher ApplicationStatus = "approved_by_teacher"
	ApplicationStatusRejected          ApplicationStatus = "rejected"
)

// Internship represents an internship opportunity posted by companies
type Internship struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyProfileID uuid.UUID      `gorm:"type:uuid;not null" json:"company_profile_id"`
	Title            string         `gorm:"not null" json:"title"`
	Description      string         `gorm:"type:text;not null" json:"description"`
	Type             InternshipType `gorm:"type:varchar(20);not null" json:"type"`
	PostedAt         time.Time      `gorm:"autoCreateTime" json:"posted_at"`
	Deadline         *time.Time     `json:"deadline"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`

	// Relationships
	CompanyProfile         *user.CompanyProfile    `gorm:"foreignKey:CompanyProfileID" json:"company_profile,omitempty"`
	InternshipApplications []InternshipApplication `gorm:"foreignKey:InternshipID" json:"internship_applications,omitempty"`
	InternshipReviews      []InternshipReview      `gorm:"foreignKey:InternshipID" json:"internship_reviews,omitempty"`
}

// InternshipApplication represents student applications to internships
type InternshipApplication struct {
	ID               uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	InternshipID     uuid.UUID         `gorm:"type:uuid;not null" json:"internship_id"`
	StudentProfileID uuid.UUID         `gorm:"type:uuid;not null" json:"student_profile_id"`
	Status           ApplicationStatus `gorm:"type:varchar(30);default:'pending'" json:"status"`
	AppliedAt        time.Time         `gorm:"autoCreateTime" json:"applied_at"`
	ReviewedAt       *time.Time        `json:"reviewed_at"`
	ApprovedBy       *uuid.UUID        `gorm:"type:uuid" json:"approved_by"` // Teacher who approved
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`

	// Relationships
	Internship        *Internship          `gorm:"foreignKey:InternshipID" json:"internship,omitempty"`
	StudentProfile    *user.StudentProfile `gorm:"foreignKey:StudentProfileID" json:"student_profile,omitempty"`
	ApprovedByTeacher *user.TeacherProfile `gorm:"foreignKey:ApprovedBy" json:"approved_by_teacher,omitempty"`
}

// InternshipReview represents student reviews/testimonials for completed internships
type InternshipReview struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	InternshipID     uuid.UUID `gorm:"type:uuid;not null" json:"internship_id"`
	StudentProfileID uuid.UUID `gorm:"type:uuid;not null" json:"student_profile_id"`
	Rating           int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"` // 1-5 rating
	Testimonial      string    `gorm:"type:text;not null" json:"testimonial"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationships
	Internship     *Internship          `gorm:"foreignKey:InternshipID" json:"internship,omitempty"`
	StudentProfile *user.StudentProfile `gorm:"foreignKey:StudentProfileID" json:"student_profile,omitempty"`
}

// BeforeCreate hooks for UUID generation (mengikuti pattern user model)
func (i *Internship) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

func (ia *InternshipApplication) BeforeCreate(tx *gorm.DB) error {
	if ia.ID == uuid.Nil {
		ia.ID = uuid.New()
	}
	return nil
}

func (ir *InternshipReview) BeforeCreate(tx *gorm.DB) error {
	if ir.ID == uuid.Nil {
		ir.ID = uuid.New()
	}
	return nil
}

// Table names
func (Internship) TableName() string {
	return "internships"
}

func (InternshipApplication) TableName() string {
	return "internship_applications"
}

func (InternshipReview) TableName() string {
	return "internship_reviews"
}

// Helper methods for validation and business logic

// IsDeadlinePassed checks if the internship deadline has passed
func (i *Internship) IsDeadlinePassed() bool {
	if i.Deadline == nil {
		return false // No deadline means always open
	}
	return time.Now().After(*i.Deadline)
}

// CanApply checks if a student can apply to this internship
func (i *Internship) CanApply() bool {
	return !i.IsDeadlinePassed()
}

// IsApproved checks if the application is approved by teacher
func (ia *InternshipApplication) IsApproved() bool {
	return ia.Status == ApplicationStatusApprovedByTeacher
}

// IsPending checks if the application is still pending
func (ia *InternshipApplication) IsPending() bool {
	return ia.Status == ApplicationStatusPending
}

// IsRejected checks if the application is rejected
func (ia *InternshipApplication) IsRejected() bool {
	return ia.Status == ApplicationStatusRejected
}

// SetReviewed marks the application as reviewed
func (ia *InternshipApplication) SetReviewed(approvedBy *uuid.UUID, status ApplicationStatus) {
	now := time.Now()
	ia.ReviewedAt = &now
	ia.ApprovedBy = approvedBy
	ia.Status = status
}

// IsValidRating checks if the rating is within valid range (1-5)
func (ir *InternshipReview) IsValidRating() bool {
	return ir.Rating >= 1 && ir.Rating <= 5
}
