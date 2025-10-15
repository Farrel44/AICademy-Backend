package cv

import (
	"fmt"
	"strings"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler interface {
	GenerateCV(c *fiber.Ctx) error
	PreviewCV(c *fiber.Ctx) error
	GetStudentCVs(c *fiber.Ctx) error
	GetCVByID(c *fiber.Ctx) error
	UpdateCV(c *fiber.Ctx) error
	DeleteCV(c *fiber.Ctx) error

	PublishCV(c *fiber.Ctx) error
	UnpublishCV(c *fiber.Ctx) error
	GetPublicCVs(c *fiber.Ctx) error

	DownloadCV(c *fiber.Ctx) error
	AnalyzeATS(c *fiber.Ctx) error
}

type CVHandler struct {
	service *CVService
}

func NewCVHandler(service *CVService) *CVHandler {
	return &CVHandler{service: service}
}

func (h *CVHandler) GenerateCV(c *fiber.Ctx) error {
	studentID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, 401, "Unauthorized")
	}

	var req GenerateCVRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	cv, err := h.service.GenerateCV(studentID, req.Title)
	if err != nil {
		return utils.SendError(c, 500, "Failed to generate CV: "+err.Error())
	}

	atsScore, _ := h.service.AnalyzeATS(&cv.Content)

	response := CVDetailResponse{
		CVResponse: CVResponse{
			ID:          cv.ID,
			Title:       cv.Title,
			Status:      cv.Status,
			IsPublic:    cv.IsPublic,
			HasPDF:      cv.PDFPath != "",
			GeneratedAt: cv.GeneratedAt,
			PublishedAt: cv.PublishedAt,
			CreatedAt:   cv.CreatedAt,
			UpdatedAt:   cv.UpdatedAt,
		},
		Content:  cv.Content,
		ATSScore: atsScore,
	}

	return utils.SendSuccess(c, "CV generated successfully with PDF. You can now apply for internships", response)
}

func (h *CVHandler) PreviewCV(c *fiber.Ctx) error {
	studentID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, 401, "Unauthorized")
	}

	content, err := h.service.PreviewCV(studentID)
	if err != nil {
		return utils.SendError(c, 500, "Failed to generate preview: "+err.Error())
	}

	atsScore, _ := h.service.AnalyzeATS(content)
	if atsScore == nil {
		atsScore = &ATSScore{Overall: 75}
	}

	response := CVPreviewResponse{
		Content:  *content,
		ATSScore: *atsScore,
	}

	return utils.SendSuccess(c, "CV preview generated successfully", response)
}

func (h *CVHandler) GetStudentCVs(c *fiber.Ctx) error {
	studentID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, 401, "Unauthorized")
	}

	cvs, err := h.service.GetStudentCVs(studentID)
	if err != nil {
		return utils.SendError(c, 500, "Failed to get CVs: "+err.Error())
	}

	var responses []CVResponse
	for _, cv := range cvs {
		responses = append(responses, CVResponse{
			ID:          cv.ID,
			Title:       cv.Title,
			Status:      cv.Status,
			IsPublic:    cv.IsPublic,
			HasPDF:      cv.PDFPath != "",
			GeneratedAt: cv.GeneratedAt,
			PublishedAt: cv.PublishedAt,
			CreatedAt:   cv.CreatedAt,
			UpdatedAt:   cv.UpdatedAt,
		})
	}

	return utils.SendSuccess(c, "CVs retrieved successfully", responses)
}

func (h *CVHandler) GetCVByID(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	cv, err := h.service.GetCVByID(cvID)
	if err != nil {
		return utils.SendError(c, 404, "CV not found")
	}

	atsScore, _ := h.service.AnalyzeATS(&cv.Content)

	response := CVDetailResponse{
		CVResponse: CVResponse{
			ID:          cv.ID,
			Title:       cv.Title,
			Status:      cv.Status,
			IsPublic:    cv.IsPublic,
			HasPDF:      cv.PDFPath != "",
			GeneratedAt: cv.GeneratedAt,
			PublishedAt: cv.PublishedAt,
			CreatedAt:   cv.CreatedAt,
			UpdatedAt:   cv.UpdatedAt,
		},
		Content:  cv.Content,
		ATSScore: atsScore,
	}

	return utils.SendSuccess(c, "CV retrieved successfully", response)
}

func (h *CVHandler) UpdateCV(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	var req UpdateCVRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := h.service.UpdateCV(cvID, &req.Content); err != nil {
		return utils.SendError(c, 500, "Failed to update CV: "+err.Error())
	}

	return utils.SendSuccess(c, "CV updated successfully", nil)
}

func (h *CVHandler) DeleteCV(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	if err := h.service.DeleteCV(cvID); err != nil {
		return utils.SendError(c, 500, "Failed to delete CV: "+err.Error())
	}

	return utils.SendSuccess(c, "CV deleted successfully", nil)
}

func (h *CVHandler) PublishCV(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	if err := h.service.PublishCV(cvID); err != nil {
		return utils.SendError(c, 500, "Failed to publish CV: "+err.Error())
	}

	return utils.SendSuccess(c, "CV published successfully", nil)
}

func (h *CVHandler) UnpublishCV(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	if err := h.service.UnpublishCV(cvID); err != nil {
		return utils.SendError(c, 500, "Failed to unpublish CV: "+err.Error())
	}

	return utils.SendSuccess(c, "CV unpublished successfully", nil)
}

func (h *CVHandler) DownloadCV(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	cv, err := h.service.GetCVByID(cvID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	if cv.PDFPath == "" {
		return utils.SendError(c, fiber.StatusNotFound, "PDF file not found")
	}

	// Set proper headers for PDF download
	filename := fmt.Sprintf("%s.pdf", strings.ReplaceAll(cv.Title, " ", "_"))
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Redirect to R2 URL for download
	return c.Redirect(cv.PDFPath, fiber.StatusTemporaryRedirect)
}

func (h *CVHandler) AnalyzeATS(c *fiber.Ctx) error {
	cvID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, 400, "Invalid CV ID")
	}

	cv, err := h.service.GetCVByID(cvID)
	if err != nil {
		return utils.SendError(c, 404, "CV not found")
	}

	atsScore, err := h.service.AnalyzeATS(&cv.Content)
	if err != nil {
		return utils.SendError(c, 500, "Failed to analyze ATS: "+err.Error())
	}

	return utils.SendSuccess(c, "ATS analysis completed", atsScore)
}

func (h *CVHandler) GetPublicCVs(c *fiber.Ctx) error {
	nis := c.Params("studentId")
	if nis == "" {
		return utils.SendError(c, 400, "NIS is required")
	}

	cvs, err := h.service.GetPublicCVs(nis)
	if err != nil {
		return utils.SendError(c, 500, "Failed to get public CVs: "+err.Error())
	}

	var responses []PublicCVResponse
	for _, cv := range cvs {
		var publicProjects []PublicCVProject
		for _, project := range cv.Content.Projects {
			publicProjects = append(publicProjects, NewPublicCVProject(project))
		}

		responses = append(responses, PublicCVResponse{
			ID:          cv.ID,
			Title:       cv.Title,
			Status:      cv.Status,
			IsPublic:    cv.IsPublic,
			HasPDF:      cv.PDFPath != "",
			GeneratedAt: cv.GeneratedAt,
			PublishedAt: cv.PublishedAt,
			CreatedAt:   cv.CreatedAt,
			UpdatedAt:   cv.UpdatedAt,
			PersonalInfo: PublicPersonalInfo{
				FullName:  cv.Content.PersonalInfo.FullName,
				Location:  cv.Content.PersonalInfo.Location,
				LinkedIn:  cv.Content.PersonalInfo.LinkedIn,
				GitHub:    cv.Content.PersonalInfo.GitHub,
				Portfolio: cv.Content.PersonalInfo.Portfolio,
			},
			Summary:        cv.Content.Summary,
			Experiences:    cv.Content.Experiences,
			Projects:       publicProjects,
			Skills:         cv.Content.Skills,
			Certifications: cv.Content.Certifications,
			Education:      cv.Content.Education,
			Languages:      cv.Content.Languages,
			Keywords:       cv.Content.Keywords,
		})
	}

	return utils.SendSuccess(c, "Public CVs retrieved successfully", responses)
}
