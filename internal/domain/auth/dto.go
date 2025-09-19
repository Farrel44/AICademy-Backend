package auth

import "github.com/google/uuid"

type RegisterAlumniRequest struct {
	Fullname string `json:"fullname" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]"`
}

type CreateTeacherRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Fullname string `json:"fullname" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=8"`
}

type CreateStudentRequest struct {
	NIS      string `json:"nis" validate:"required,min=8,max=20"`
	Class    string `json:"class" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Fullname string `json:"fullname" validate:"required,min=2"`
}

type CreateCompanyRequest struct {
	CompanyName     string  `json:"company_name" validate:"required,min=2"`
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required,min=8"`
	CompanyLogo     *string `json:"company_logo"`
	CompanyLocation *string `json:"company_location"`
	Description     *string `json:"description"`
}

type StudentCSVRow struct {
	NIS      string `csv:"nis" validate:"required"`
	Class    string `csv:"class" validate:"required"`
	Email    string `csv:"email" validate:"required,email"`
	Fullname string `csv:"fullname" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Password        string `json:"password" validate:"required,min=8,regexp=^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required"`
}

type AuthResponse struct {
	Token                 string      `json:"token"`
	User                  UserProfile `json:"user"`
	RequirePasswordChange bool        `json:"require_password_change,omitempty"`
}

type UserProfile struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
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
