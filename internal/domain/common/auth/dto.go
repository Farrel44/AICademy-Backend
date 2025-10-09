package auth

import (
	"time"

	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken           string      `json:"access_token"`
	RefreshToken          string      `json:"refresh_token"`
	TokenType             string      `json:"token_type"`
	ExpiresIn             int64       `json:"expires_in"`
	User                  UserProfile `json:"user"`
	RequirePasswordChange bool        `json:"require_password_change,omitempty"`
}

type LegacyAuthResponse struct {
	Token                 string      `json:"token"`
	User                  UserProfile `json:"user"`
	RequirePasswordChange bool        `json:"require_password_change,omitempty"`
}

type UserProfile struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
	Name  string    `json:"name,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error"`
	Details []ValidationError `json:"details,omitempty"`
}
type StudentProfileResponse struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	Fullname       string     `json:"fullname"`
	NIS            string     `json:"nis"`
	Class          string     `json:"class"`
	ProfilePicture string     `json:"profile_picture"`
	Headline       string     `json:"headline"`
	Bio            string     `json:"bio"`
	CVFile         *string    `json:"cv_file"`
	Email          string     `json:"email"`
	Role           string     `json:"user_role"`
	RoleID         *uuid.UUID `json:"role_id,omitempty"`
	RoleName       *string    `json:"role_name,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
type AlumniProfileResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Fullname       string    `json:"fullname"`
	ProfilePicture string    `json:"profile_picture"`
	Headline       string    `json:"headline"`
	Bio            string    `json:"bio"`
	CVFile         *string   `json:"cv_file"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TeacherProfileResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Fullname       string    `json:"fullname"`
	ProfilePicture string    `json:"profile_picture"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}

type CompanyProfileResponse struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	CompanyName     string    `json:"company_name"`
	CompanyLogo     *string   `json:"company_logo"`
	CompanyLocation *string   `json:"company_location"`
	Description     *string   `json:"description"`
	Email           string    `json:"email"`
	Role            string    `json:"role"`
	CreatedAt       time.Time `json:"created_at"`
}

type AdminProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
