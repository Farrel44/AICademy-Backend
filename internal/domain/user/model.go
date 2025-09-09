package user

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleStudent UserRole = "Student"
	RoleTeacher UserRole = "Teacher"
	RoleAlumni  UserRole = "Alumni"
	RoleAdmin   UserRole = "Admin"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Email          string         `gorm:"uniqueIndex; not null" json:"email"`
	Password       string         `gorm:"not null" json:"-"`
	Role           UserRole       `gorm:"not null" json:"role"`
	Department     string         `json:"department,omitempty"` //Jurusan
	Cohort         int            `json:"cohort,omitempty"`     //Angkatan
	ProfileVisible bool           `gorm:"default:true" json:"profile_visible"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
