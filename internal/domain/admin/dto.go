package admin

import (
	"time"

	"github.com/google/uuid"
)

type CreateStudentRequest struct {
	Fullname string `json:"fullname" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	NIS      string `json:"nis" validate:"required,min=8,max=20"`
	Class    string `json:"class" validate:"required"`
}

type UpdateStudentRequest struct {
	Fullname *string `json:"fullname,omitempty" form:"fullname" validate:"omitempty,min=2"`
	NIS      *string `json:"nis,omitempty" form:"nis" validate:"omitempty,min=8,max=20"`
	Class    *string `json:"class,omitempty" form:"class"`
	Headline *string `json:"headline,omitempty" form:"headline"`
	Bio      *string `json:"bio,omitempty" form:"bio"`
}

type StudentResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Fullname       string    `json:"fullname"`
	NIS            string    `json:"nis"`
	Class          string    `json:"class"`
	ProfilePicture *string   `json:"profile_picture"`
	Headline       *string   `json:"headline"`
	Bio            *string   `json:"bio"`
	CVFile         *string   `json:"cv_file"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PaginatedStudentsResponse struct {
	Data       []StudentResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type StudentStatisticsResponse struct {
	TotalStudents int `json:"total_students"`
	TotalRPL      int `json:"total_rpl"`
	TotalTKJ      int `json:"total_tkj"`
}

type CreateTeacherRequest struct {
	Fullname string `json:"fullname" form:"fullname" validate:"required,min=2"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=8"`
}

type UpdateTeacherRequest struct {
	Fullname       *string `json:"fullname,omitempty" form:"fullname" validate:"omitempty,min=2"`
	ProfilePicture *string `json:"profile_picture,omitempty" form:"profile_picture"`
}

type TeacherResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Fullname       string    `json:"fullname"`
	ProfilePicture *string   `json:"profile_picture"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaginatedTeachersResponse struct {
	Data       []TeacherResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type CreateCompanyRequest struct {
	CompanyName     string  `json:"company_name" form:"company_name" validate:"required,min=2"`
	Email           string  `json:"email" form:"email" validate:"required,email"`
	Password        string  `json:"password" form:"password" validate:"required,min=8"`
	CompanyLogo     *string `json:"company_logo,omitempty" form:"company_logo"`
	CompanyLocation *string `json:"company_location,omitempty" form:"company_location"`
	Description     *string `json:"description,omitempty" form:"description"`
}

type UpdateCompanyRequest struct {
	CompanyName     *string `json:"company_name,omitempty" form:"company_name" validate:"omitempty,min=2"`
	CompanyLogo     *string `json:"company_logo,omitempty" form:"company_logo"`
	CompanyLocation *string `json:"company_location,omitempty" form:"company_location"`
	Description     *string `json:"description,omitempty" form:"description"`
}

type CompanyResponse struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	CompanyName     string    `json:"company_name"`
	CompanyLogo     *string   `json:"company_logo"`
	CompanyLocation *string   `json:"company_location"`
	Description     *string   `json:"description"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"created_at"`
}

type PaginatedCompaniesResponse struct {
	Data       []CompanyResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type UpdateAlumniRequest struct {
	Fullname       *string `json:"fullname,omitempty" form:"fullname" validate:"omitempty,min=2"`
	ProfilePicture *string `json:"profile_picture,omitempty" form:"profile_picture"`
	Headline       *string `json:"headline,omitempty" form:"headline"`
	Bio            *string `json:"bio,omitempty" form:"bio"`
}

type AlumniResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Fullname       string    `json:"fullname"`
	ProfilePicture *string   `json:"profile_picture"`
	Headline       *string   `json:"headline"`
	Bio            *string   `json:"bio"`
	CVFile         *string   `json:"cv_file"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PaginatedAlumniResponse struct {
	Data       []AlumniResponse `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}
