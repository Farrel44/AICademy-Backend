package pkl

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CompanyPklHandler struct {
	service *CompanyPklService
}

func NewCompanyPklHandler(service *CompanyPklService) *CompanyPklHandler {
	return &CompanyPklHandler{
		service: service,
	}
}

func (h *CompanyPklHandler) GetMyInternships(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	internships, total, err := h.service.GetCompanyInternships(c, offset, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	response := map[string]interface{}{
		"internships": internships,
		"pagination": map[string]interface{}{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	}

	return utils.SendSuccess(c, "Company internships retrieved successfully", response)
}

func (h *CompanyPklHandler) CreateInternship(c *fiber.Ctx) error {
	var req CreateInternshipRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	internship, err := h.service.CreateInternshipPosition(c, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Internship created successfully", internship)
}

func (h *CompanyPklHandler) UpdateInternship(c *fiber.Ctx) error {
	internshipIDParam := c.Params("id")
	internshipID, err := uuid.Parse(internshipIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	var req UpdateInternshipRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	internship, err := h.service.UpdateInternshipPosition(c, internshipID, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Internship updated successfully", internship)
}

func (h *CompanyPklHandler) DeleteInternship(c *fiber.Ctx) error {
	internshipIDParam := c.Params("id")
	internshipID, err := uuid.Parse(internshipIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	err = h.service.DeleteInternshipPosition(c, internshipID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Internship deleted successfully", nil)
}

func (h *CompanyPklHandler) GetInternshipApplications(c *fiber.Ctx) error {
	internshipIDParam := c.Params("id")
	internshipID, err := uuid.Parse(internshipIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	applications, err := h.service.GetInternshipApplications(c, internshipID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Applications retrieved successfully", applications)
}

func (h *CompanyPklHandler) UpdateApplicationStatus(c *fiber.Ctx) error {
	applicationIDParam := c.Params("id")
	applicationID, err := uuid.Parse(applicationIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid application ID")
	}

	var req UpdateApplicationStatusRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err = h.service.UpdateApplicationStatus(c, applicationID, req.Status)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Application status updated successfully", nil)
}
