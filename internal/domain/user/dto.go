package user

type UpdateStudentRequest struct {
	ProfilePicture *string `json:"profile_picture"`
	Bio            *string `json:"bio"`
	Headline       *string `json:"headline"`
	CvFile         *string `json:"cv_file"`
}
