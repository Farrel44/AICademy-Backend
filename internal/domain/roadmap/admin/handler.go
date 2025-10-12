package admin

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminRoadmapHandler struct {
	service *AdminRoadmapService
}

func NewAdminRoadmapHandler(service *AdminRoadmapService) *AdminRoadmapHandler {
	return &AdminRoadmapHandler{service: service}
}

func (h *AdminRoadmapHandler) CreateRoadmap(c *fiber.Ctx) error {
	adminID, _ := uuid.Parse(c.Locals("user_id").(string))

	var req CreateRoadmapRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.CreateRoadmap(req, adminID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Roadmap berhasil dibuat", result)
}

func (h *AdminRoadmapHandler) GetRoadmaps(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	status := c.Query("status")
	roleIDStr := c.Query("profiling_role_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var profilingRoleID *uuid.UUID
	if roleIDStr != "" {
		if id, err := uuid.Parse(roleIDStr); err == nil {
			profilingRoleID = &id
		}
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	result, err := h.service.GetRoadmaps(page, limit, search, profilingRoleID, statusPtr)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data roadmap berhasil diambil", result)
}

func (h *AdminRoadmapHandler) GetRoadmapByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID roadmap tidak valid")
	}

	result, err := h.service.GetRoadmapByID(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Detail roadmap berhasil diambil", result)
}

func (h *AdminRoadmapHandler) UpdateRoadmap(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID roadmap tidak valid")
	}

	var req UpdateRoadmapRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.UpdateRoadmap(id, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Roadmap berhasil diperbarui", result)
}

func (h *AdminRoadmapHandler) DeleteRoadmap(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID roadmap tidak valid")
	}

	err = h.service.DeleteRoadmap(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Roadmap berhasil dihapus", nil)
}

func (h *AdminRoadmapHandler) CreateRoadmapStep(c *fiber.Ctx) error {
	roadmapIDStr := c.Params("roadmapId")
	roadmapID, err := uuid.Parse(roadmapIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID roadmap tidak valid")
	}

	var req CreateStepRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.CreateRoadmapStep(roadmapID, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Step berhasil dibuat", result)
}

func (h *AdminRoadmapHandler) UpdateRoadmapStep(c *fiber.Ctx) error {
	stepIDStr := c.Params("stepId")
	stepID, err := uuid.Parse(stepIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID step tidak valid")
	}

	var req UpdateStepRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.UpdateRoadmapStep(stepID, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Step berhasil diperbarui", result)
}

func (h *AdminRoadmapHandler) DeleteRoadmapStep(c *fiber.Ctx) error {
	stepIDStr := c.Params("stepId")
	stepID, err := uuid.Parse(stepIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID step tidak valid")
	}

	err = h.service.DeleteRoadmapStep(stepID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Step berhasil dihapus", nil)
}

func (h *AdminRoadmapHandler) GetStudentProgress(c *fiber.Ctx) error {
	roadmapIDStr := c.Params("roadmapId")
	roadmapID, err := uuid.Parse(roadmapIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID roadmap tidak valid")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.service.GetAllStudentProgress(roadmapID, page, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data progress siswa berhasil diambil", result)
}

func (h *AdminRoadmapHandler) GetPendingSubmissions(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")
	teacherIDStr := c.Query("teacher_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var teacherID *uuid.UUID
	if teacherIDStr != "" {
		if id, err := uuid.Parse(teacherIDStr); err == nil {
			teacherID = &id
		}
	}

	result, err := h.service.GetPendingSubmissions(page, limit, search, teacherID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data submission pending berhasil diambil", result)
}

func (h *AdminRoadmapHandler) GetStatistics(c *fiber.Ctx) error {
	result, err := h.service.GetStatistics()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Statistik roadmap berhasil diambil", result)
}
