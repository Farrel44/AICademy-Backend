package student

type CreateStudentRequest struct {
	NIS      string `json:"nis" validate:"required,min=8,max=20"`
	Class    string `json:"class" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Fullname string `json:"fullname" validate:"required,min=2"`
}

type StudentCSVRow struct {
	NIS      string `csv:"nis" validate:"required"`
	Class    string `csv:"class" validate:"required"`
	Email    string `csv:"email" validate:"required,email"`
	Fullname string `csv:"fullname" validate:"required"`
}

type BulkCreateStudentsRequest struct {
	Students []CreateStudentRequest `json:"students" validate:"required,dive"`
}
type ChangeDefaultPasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

type BulkCreateResult struct {
	SuccessCount int      `json:"success_count"`
	FailedCount  int      `json:"failed_count"`
	Errors       []string `json:"errors,omitempty"`
}
