package cv

import (
	"strings"
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
	PDFLink  string    `json:"pdf_link"`
	Content  CVContent `json:"content"`
	ATSScore *ATSScore `json:"ats_score,omitempty"`
}

type CVPreviewResponse struct {
	Content  CVContent `json:"content"`
	ATSScore ATSScore  `json:"ats_score"`
}

type PublicPersonalInfo struct {
	FullName  string `json:"full_name"`
	Location  string `json:"location"`
	LinkedIn  string `json:"linkedin,omitempty"`
	GitHub    string `json:"github,omitempty"`
	Portfolio string `json:"portfolio,omitempty"`
}

type PublicCVProject struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Role         string     `json:"role"`
	Technologies []string   `json:"technologies"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	URL          string     `json:"url,omitempty"`
	TechStack    []string   `json:"tech_stack,omitempty"`
	Highlights   []string   `json:"highlights"`
}

func NewPublicCVProject(cvProject CVProject) PublicCVProject {
	var techStack []string
	if cvProject.TechStack != "" {
		techStackStr := strings.Trim(cvProject.TechStack, "{}")
		if techStackStr != "" {
			parts := strings.Split(techStackStr, ",")
			for _, part := range parts {
				cleaned := strings.TrimSpace(part)
				if cleaned != "" {
					techStack = append(techStack, cleaned)
				}
			}
		}
	}

	return PublicCVProject{
		ID:           cvProject.ID,
		Name:         cvProject.Name,
		Description:  cvProject.Description,
		Role:         cvProject.Role,
		Technologies: cvProject.Technologies,
		StartDate:    cvProject.StartDate,
		EndDate:      cvProject.EndDate,
		URL:          cvProject.URL,
		TechStack:    techStack,
		Highlights:   cvProject.Highlights,
	}
}

type PublicCVResponse struct {
	ID             uuid.UUID          `json:"id"`
	Title          string             `json:"title"`
	Status         CVStatus           `json:"status"`
	IsPublic       bool               `json:"is_public"`
	HasPDF         bool               `json:"has_pdf"`
	GeneratedAt    time.Time          `json:"generated_at"`
	PublishedAt    *time.Time         `json:"published_at,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	PersonalInfo   PublicPersonalInfo `json:"personal_info"`
	Summary        string             `json:"summary"`
	Experiences    []CVExperience     `json:"experiences"`
	Projects       []PublicCVProject  `json:"projects"`
	Skills         []CVSkill          `json:"skills"`
	Certifications []CVCertification  `json:"certifications"`
	Education      CVEducation        `json:"education"`
	Languages      []CVLanguage       `json:"languages"`
	Keywords       []string           `json:"keywords"`
}
