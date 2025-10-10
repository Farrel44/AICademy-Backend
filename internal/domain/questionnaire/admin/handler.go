package admin

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/utils"

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
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.CreateTargetRole(req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Target role berhasil dibuat", result)
}

func (h *AdminQuestionnaireHandler) GetTargetRoles(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	// Validate page and limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Validate search parameter length
	if len(search) > 100 {
		return utils.SendError(c, fiber.StatusBadRequest, "Search parameter too long")
	}

	result, err := h.service.GetTargetRoles(page, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data target role berhasil diambil", result)
}

func (h *AdminQuestionnaireHandler) GetTargetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID target role tidak valid")
	}

	result, err := h.service.GetTargetRoleByID(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Data target role berhasil diambil", result)
}

func (h *AdminQuestionnaireHandler) UpdateTargetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID target role tidak valid")
	}

	var req UpdateTargetRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.UpdateTargetRole(id, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Target role berhasil diperbarui", result)
}

func (h *AdminQuestionnaireHandler) DeleteTargetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID target role tidak valid")
	}

	err = h.service.DeleteTargetRole(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Target role berhasil dihapus", nil)
}

// Questionnaire Generation
func (h *AdminQuestionnaireHandler) GenerateQuestionnaire(c *fiber.Ctx) error {
	var req GenerateQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.GenerateQuestionnaire(req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Kuesioner berhasil dibuat", result)
}

// Get Generation Status
func (h *AdminQuestionnaireHandler) GetGenerationStatus(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("id")
	if questionnaireIDStr == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "ID kuesioner wajib diisi")
	}

	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format ID kuesioner tidak valid")
	}

	status, err := h.service.GetGenerationStatus(questionnaireID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	response := QuestionnaireGenerationResponse{
		QuestionnaireID: questionnaireID,
		Status:          status.Status,
		Progress:        status.Progress,
		Message:         status.Message,
	}

	return utils.SendSuccess(c, "Status generasi berhasil diambil", response)
}

// Questionnaire Management
func (h *AdminQuestionnaireHandler) GetQuestionnaires(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.service.GetQuestionnaires(page, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data kuesioner berhasil diambil", result)
}

func (h *AdminQuestionnaireHandler) GetQuestionnaireDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID kuesioner tidak valid")
	}

	result, err := h.service.GetQuestionnaireDetail(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Detail kuesioner berhasil diambil", result)
}

func (h *AdminQuestionnaireHandler) ActivateQuestionnaire(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID kuesioner tidak valid")
	}

	var req ActivateQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	err = h.service.ActivateQuestionnaire(id, req.Active)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	status := "dinonaktifkan"
	if req.Active {
		status = "diaktifkan"
	}

	return utils.SendSuccess(c, "Kuesioner berhasil "+status, nil)
}

// Response Management
func (h *AdminQuestionnaireHandler) GetQuestionnaireResponses(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var questionnaireID *uuid.UUID
	if qIDStr := c.Query("questionnaire_id"); qIDStr != "" {
		if id, err := uuid.Parse(qIDStr); err == nil {
			questionnaireID = &id
		}
	}

	result, err := h.service.GetQuestionnaireResponses(page, limit, search, questionnaireID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data respons kuesioner berhasil diambil", result)
}

func (h *AdminQuestionnaireHandler) GetResponseDetail(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID respons tidak valid")
	}

	result, err := h.service.GetResponseDetail(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Detail respons berhasil diambil", result)
}
