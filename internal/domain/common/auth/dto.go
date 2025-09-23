package auth

import "github.com/google/uuid"

// Common login untuk semua role
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken           string      `json:"access_token"`
	RefreshToken          string      `json:"refresh_token"`
	TokenType             string      `json:"token_type"`
	ExpiresIn             int64       `json:"expires_in"` // dalam detik
	User                  UserProfile `json:"user"`
	RequirePasswordChange bool        `json:"require_password_change,omitempty"`
}

// Legacy response untuk backward compatibility
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

// Common forgot/reset password
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required"`
}

// Common change password
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
	ExpiresIn   int64  `json:"expires_in"` // dalam detik
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
