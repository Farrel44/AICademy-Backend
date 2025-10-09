package user

import (
	"time"

	"github.com/google/uuid"
)

type UpdateStudentRequest struct {
	ProfilePicture *string `json:"profile_picture"`
	Bio            *string `json:"bio"`
	Headline       *string `json:"headline"`
	CvFile         *string `json:"cv_file"`
}

type RecommendedRoleInfo struct {
	RoleID          uuid.UUID `json:"role_id"`
	RoleName        string    `json:"role_name"`
	RoleDescription string    `json:"role_description"`
	RoleCategory    string    `json:"role_category"`
	Score           *float64  `json:"score,omitempty"`
	Justification   *string   `json:"justification,omitempty"`
}

type EnhancedStudentProfile struct {
	ID              uuid.UUID            `json:"id"`
	UserID          uuid.UUID            `json:"user_id"`
	Fullname        string               `json:"fullname"`
	NIS             string               `json:"nis"`
	Class           string               `json:"class"`
	ProfilePicture  string               `json:"profile_picture"`
	Headline        string               `json:"headline"`
	Bio             string               `json:"bio"`
	CVFile          *string              `json:"cv_file"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	RecommendedRole *RecommendedRoleInfo `json:"recommended_role,omitempty"`
}

type EnhancedUserResponse struct {
	ID             uuid.UUID               `json:"id"`
	Email          string                  `json:"email"`
	Role           UserRole                `json:"role"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	StudentProfile *EnhancedStudentProfile `json:"student_profile,omitempty"`
}
