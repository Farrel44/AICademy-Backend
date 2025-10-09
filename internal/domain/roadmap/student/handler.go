package student

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentRoadmapHandler struct {
	service *StudentRoadmapService
}

func NewStudentRoadmapHandler(service *StudentRoadmapService) *StudentRoadmapHandler {
	return &StudentRoadmapHandler{service: service}
}

func (h *StudentRoadmapHandler) GetAvailableRoadmaps(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.service.GetAvailableRoadmaps(studentProfileID, page, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data roadmap tersedia berhasil diambil", result)
}

func (h *StudentRoadmapHandler) StartRoadmap(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	var req StartRoadmapRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.StartRoadmap(req.RoadmapID, studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Roadmap berhasil dimulai", result)
}

func (h *StudentRoadmapHandler) GetRoadmapProgress(c *fiber.Ctx) error {
	roadmapIDStr := c.Params("roadmapId")
	roadmapID, err := uuid.Parse(roadmapIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID roadmap tidak valid")
	}

	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	result, err := h.service.GetRoadmapProgress(roadmapID, studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Progress roadmap berhasil diambil", result)
}

func (h *StudentRoadmapHandler) StartStep(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	var req StartStepRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.StartStep(req.StepID, studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Step berhasil dimulai", result)
}

func (h *StudentRoadmapHandler) SubmitEvidence(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	var req SubmitEvidenceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Validasi gagal")
	}

	result, err := h.service.SubmitEvidence(req, studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Bukti berhasil dikirim", result)
}

func (h *StudentRoadmapHandler) GetMyProgress(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	result, err := h.service.GetMyProgress(studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Progress saya berhasil diambil", result)
}

func (h *StudentRoadmapHandler) getStudentProfileID(userID uuid.UUID) (uuid.UUID, error) {
	profile, err := h.service.repo.GetStudentProfile(userID)
	if err != nil {
		return uuid.Nil, err
	}
	return profile.ID, nil
}
