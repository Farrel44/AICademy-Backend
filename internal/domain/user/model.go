package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleStudent UserRole = "student"
	RoleTeacher UserRole = "teacher"
	RoleAlumni  UserRole = "alumni"
	RoleAdmin   UserRole = "admin"
	RoleCompany UserRole = "company"
)

type User struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Email              string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash       string     `gorm:"column:password_hash;not null" json:"-"`
	Role               UserRole   `gorm:"not null" json:"role"`
	PasswordResetToken *string    `gorm:"column:password_reset_token" json:"-"`
	PasswordResetAt    *time.Time `gorm:"column:password_reset_at" json:"-"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type ResetPasswordToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Token     string     `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type StudentProfile struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Fullname       string    `gorm:"not null" json:"fullname"`
	NIS            string    `gorm:"not null" json:"nis"`
	Class          string    `gorm:"not null" json:"class"`
	ProfilePicture string    `json:"profile_picture"`
	Headline       string    `json:"headline"`
	Bio            string    `json:"bio"`
	CVFile         *string   `json:"cv_file"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *StudentProfile) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type AlumniProfile struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Fullname       string    `gorm:"not null" json:"fullname"`
	ProfilePicture string    `json:"profile_picture"`
	Headline       string    `json:"headline"`
	Bio            string    `json:"bio"`
	CVFile         *string   `json:"cv_file"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (a *AlumniProfile) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type TeacherProfile struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Fullname       string    `gorm:"not null" json:"fullname"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
}

func (t *TeacherProfile) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type CompanyProfile struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CompanyName     string    `gorm:"not null" json:"company_name"`
	CompanyLogo     *string   `json:"company_logo"`
	CompanyLocation *string   `json:"company_location"`
	Description     *string   `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

func (c *CompanyProfile) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Method untuk cek apakah token masih valid
func (r *ResetPasswordToken) IsValid() bool {
	return r.UsedAt == nil && time.Now().Before(r.ExpiresAt)
}

// Method untuk mark token sebagai used
func (r *ResetPasswordToken) MarkAsUsed() {
	now := time.Now()
	r.UsedAt = &now
}

// BlacklistedToken untuk menyimpan token JWT yang sudah di-logout
type BlacklistedToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"token_hash"` // Hash dari JWT token untuk security
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *BlacklistedToken) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// RefreshToken untuk menyimpan refresh token yang valid
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// Method untuk cek apakah refresh token masih valid
func (r *RefreshToken) IsValid() bool {
	return time.Now().Before(r.ExpiresAt)
}
