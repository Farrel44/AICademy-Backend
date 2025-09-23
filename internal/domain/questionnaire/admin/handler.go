package admin

import (
	"aicademy-backend/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminQuestionnaireHandler struct {
	service *AdminQuestionnaireService
}

func NewAdminQuestionnaireHandler(service *AdminQuestionnaireService) *AdminQuestionnaireHandler {
	return &AdminQuestionnaireHandler{service: service}
}

// Target Role Management
func (h *AdminQuestionnaireHandler) CreateTargetRole(c *fiber.Ctx) error {
	var req CreateTargetRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.CreateTargetRole(req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Target role created successfully", result)
}

func (h *AdminQuestionnaireHandler) GetTargetRoles(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.service.GetTargetRoles(page, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Target roles retrieved successfully", result)
}

func (h *AdminQuestionnaireHandler) GetTargetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid target role ID")
	}

	result, err := h.service.GetTargetRoleByID(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Target role retrieved successfully", result)
}

func (h *AdminQuestionnaireHandler) UpdateTargetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid target role ID")
	}

	var req UpdateTargetRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.UpdateTargetRole(id, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Target role updated successfully", result)
}

func (h *AdminQuestionnaireHandler) DeleteTargetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid target role ID")
	}

	err = h.service.DeleteTargetRole(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Target role deleted successfully", nil)
}

// Questionnaire Generation
func (h *AdminQuestionnaireHandler) GenerateQuestionnaire(c *fiber.Ctx) error {
	var req GenerateQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.GenerateQuestionnaire(req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Questionnaire generated successfully", result)
}

// Questionnaire Management
func (h *AdminQuestionnaireHandler) GetQuestionnaires(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.service.GetQuestionnaires(page, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Questionnaires retrieved successfully", result)
}

func (h *AdminQuestionnaireHandler) GetQuestionnaireDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid questionnaire ID")
	}

	result, err := h.service.GetQuestionnaireDetail(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Questionnaire detail retrieved successfully", result)
}

func (h *AdminQuestionnaireHandler) ActivateQuestionnaire(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid questionnaire ID")
	}

	var req ActivateQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	err = h.service.ActivateQuestionnaire(id, req.Active)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	status := "deactivated"
	if req.Active {
		status = "activated"
	}

	return utils.SendSuccess(c, "Questionnaire "+status+" successfully", nil)
}

// Response Management
func (h *AdminQuestionnaireHandler) GetQuestionnaireResponses(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Optional filter by questionnaire ID
	var questionnaireID *uuid.UUID
	if qIDStr := c.Query("questionnaire_id"); qIDStr != "" {
		if id, err := uuid.Parse(qIDStr); err == nil {
			questionnaireID = &id
		}
	}

	result, err := h.service.GetQuestionnaireResponses(page, limit, questionnaireID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Questionnaire responses retrieved successfully", result)
}

func (h *AdminQuestionnaireHandler) GetResponseDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid response ID")
	}

	result, err := h.service.GetResponseDetail(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Response detail retrieved successfully", result)
}

// SetupAdminQuestionnaireRoutes - Setup admin questionnaire routes
func SetupAdminQuestionnaireRoutes(router fiber.Router, handler *AdminQuestionnaireHandler) {
	admin := router.Group("/admin/questionnaire")

	// Target Role Management
	targetRoles := admin.Group("/target-roles")
	targetRoles.Post("/", handler.CreateTargetRole)
	targetRoles.Get("/", handler.GetTargetRoles)
	targetRoles.Get("/:id", handler.GetTargetRole)
	targetRoles.Put("/:id", handler.UpdateTargetRole)
	targetRoles.Delete("/:id", handler.DeleteTargetRole)

	// Questionnaire Management
	questionnaires := admin.Group("/questionnaires")
	questionnaires.Post("/generate", handler.GenerateQuestionnaire)
	questionnaires.Get("/", handler.GetQuestionnaires)
	questionnaires.Get("/:id", handler.GetQuestionnaireDetail)
	questionnaires.Put("/:id/activate", handler.ActivateQuestionnaire)

	// Response Management
	responses := admin.Group("/responses")
	responses.Get("/", handler.GetQuestionnaireResponses)
	responses.Get("/:id", handler.GetResponseDetail)
}