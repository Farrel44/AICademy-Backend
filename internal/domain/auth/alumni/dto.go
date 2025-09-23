package alumni

type RegisterAlumniRequest struct {
	Fullname string `json:"fullname" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateAlumniProfileRequest struct {
	Fullname       string  `json:"fullname,omitempty"`
	ProfilePicture *string `json:"profile_picture,omitempty"`
	Headline       *string `json:"headline,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	CVFile         *string `json:"cv_file,omitempty"`
	CompanyName    *string `json:"company_name,omitempty"`
	JobTitle       *string `json:"job_title,omitempty"`
	LinkedinURL    *string `json:"linkedin_url,omitempty"`
}
