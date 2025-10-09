package pkl

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AlumniPklHandler struct {
	service *AlumniPklService
}

func NewAlumniPklHandler(service *AlumniPklService) *AlumniPklHandler {
	return &AlumniPklHandler{
		service: service,
	}
}

func (h *AlumniPklHandler) ApplyPklPosition(c *fiber.Ctx) error {
	var req ApplyInternshipRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.ApplyAlumniInternshipPosition(c, req.InternshipID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Application submitted successfully", result)
}

func (h *AlumniPklHandler) GetAvailablePositions(c *fiber.Ctx) error {
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

	internships, total, err := h.service.GetAvailablePositions(c, offset, limit, search)
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

	return utils.SendSuccess(c, "Available positions retrieved successfully", response)
}

func (h *AlumniPklHandler) GetMyApplications(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	applications, total, err := h.service.GetAlumniApplications(c, offset, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	response := map[string]interface{}{
		"applications": applications,
		"pagination": map[string]interface{}{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	}

	return utils.SendSuccess(c, "Applications retrieved successfully", response)
}

func (h *AlumniPklHandler) GetApplicationByID(c *fiber.Ctx) error {
	applicationIDParam := c.Params("id")
	applicationID, err := uuid.Parse(applicationIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid application ID")
	}

	application, err := h.service.GetApplicationByID(c, applicationID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Application retrieved successfully", application)
}
