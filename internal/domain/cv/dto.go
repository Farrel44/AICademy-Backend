package cv

import (
	"time"

	"github.com/google/uuid"
)

type GenerateCVRequest struct {
	Title string `json:"title" validate:"required,min=3,max=100"`
}

type UpdateCVRequest struct {
	Title   string    `json:"title,omitempty"`
	Content CVContent `json:"content,omitempty"`
}

type CVResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Status      CVStatus   `json:"status"`
	IsPublic    bool       `json:"is_public"`
	HasPDF      bool       `json:"has_pdf"`
	GeneratedAt time.Time  `json:"generated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CVDetailResponse struct {
	CVResponse
	Content  CVContent `json:"content"`
	ATSScore *ATSScore `json:"ats_score,omitempty"`
}

type CVPreviewResponse struct {
	Content  CVContent `json:"content"`
	ATSScore ATSScore  `json:"ats_score"`
}
