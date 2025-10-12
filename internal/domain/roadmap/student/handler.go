package student

import (
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

func (h *StudentRoadmapHandler) GetMyRoadmap(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	result, err := h.service.GetMyRoadmap(studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	// If no roadmap found but service returned success with message
	if result.Roadmap == nil {
		return utils.SendSuccess(c, result.Message, nil)
	}

	return utils.SendSuccess(c, result.Message, result.Roadmap)
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

func (h *StudentRoadmapHandler) GetStepProgress(c *fiber.Ctx) error {
	userID, _ := uuid.Parse(c.Locals("user_id").(string))
	stepID, err := uuid.Parse(c.Params("stepId"))
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "ID step tidak valid")
	}

	studentProfileID, err := h.getStudentProfileID(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profile siswa tidak ditemukan")
	}

	result, err := h.service.GetStepProgress(stepID, studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Step progress berhasil diambil", result)
}

func (h *StudentRoadmapHandler) getStudentProfileID(userID uuid.UUID) (uuid.UUID, error) {
	profile, err := h.service.repo.GetStudentProfile(userID)
	if err != nil {
		return uuid.Nil, err
	}
	return profile.ID, nil
}
