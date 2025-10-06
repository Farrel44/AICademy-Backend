package profile

import "github.com/google/uuid"

type UserProfileResponse struct {
	ID             uuid.UUID `json:"id"`
	Fullname       string    `json:"fullname"`
	Nis            string    `json:"nis"`
	Class          string    `json:"class"`
	ProfilePicture string    `json:"profile_picture"`
	Headline       string    `json:"headline"`
	Bio            string    `json:"bio"`
}
