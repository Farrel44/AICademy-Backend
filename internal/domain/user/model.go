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

	StudentProfile *StudentProfile `gorm:"foreignKey:UserID;references:ID" json:"student_profile,omitempty"`
	AlumniProfile  *AlumniProfile  `gorm:"foreignKey:UserID;references:ID" json:"alumni_profile,omitempty"`
	CompanyProfile *CompanyProfile `gorm:"foreignKey:UserID;references:ID" json:"company_profile,omitempty"`
	TeacherProfile *TeacherProfile `gorm:"foreignKey:UserID;references:ID" json:"teacher_profile,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Add these methods to implement JWTUser interface
func (u *User) GetID() uuid.UUID {
	return u.ID
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetRole() string {
	return string(u.Role)
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

	Photos []CompanyPhoto `gorm:"foreignKey:CompanyID;constraint:OnDelete:CASCADE" json:"photos"`
}

type CompanyPhoto struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	PhotoURL    string    `gorm:"not null" json:"photo_url"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (CompanyPhoto) TableName() string {
	return "company_photos"
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

// Add these models at the end of the file
type Project struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerStudentProfileID uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_student_profile_id"`
	ProjectName           string    `gorm:"not null" json:"project_name"`
	Description           string    `gorm:"type:text" json:"description"`
	LinkURL               *string   `gorm:"type:text" json:"link_url"`
	StartDate             time.Time `gorm:"type:date" json:"start_date"`
	EndDate               time.Time `gorm:"type:date" json:"end_date"`
	CreatedAt             time.Time `gorm:"type:timestamptz" json:"created_at"`

	Photos []struct {
		ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
		ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
		URL       string    `gorm:"type:text;not null" json:"url"`
		Caption   *string   `gorm:"type:text" json:"caption"`
		IsPrimary bool      `gorm:"default:false" json:"is_primary"`
		CreatedAt time.Time `gorm:"type:timestamptz" json:"created_at"`
	} `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"photos,omitempty"`
}

func (Project) TableName() string {
	return "projects"
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

	Photos []struct {
		ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
		CertificationID uuid.UUID `gorm:"type:uuid;not null;index" json:"certification_id"`
		URL             string    `gorm:"type:text;not null" json:"url"`
		Caption         *string   `gorm:"type:text" json:"caption"`
		IsPrimary       bool      `gorm:"default:false" json:"is_primary"`
		CreatedAt       time.Time `gorm:"type:timestamptz" json:"created_at"`
	} `gorm:"foreignKey:CertificationID;constraint:OnDelete:CASCADE" json:"photos,omitempty"`
}

func (Certification) TableName() string {
	return "certifications"
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
