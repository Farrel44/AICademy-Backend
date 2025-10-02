package student

import (
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StudentQuestionnaireHandler struct {
	service *StudentQuestionnaireService
}

func NewStudentQuestionnaireHandler(service *StudentQuestionnaireService) *StudentQuestionnaireHandler {
	return &StudentQuestionnaireHandler{service: service}
}

func (h *StudentQuestionnaireHandler) GetActiveQuestionnaire(c *fiber.Ctx) error {
	result, err := h.service.GetActiveQuestionnaire()
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Kuesioner aktif berhasil diambil", result)
}

func (h *StudentQuestionnaireHandler) SubmitQuestionnaire(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	studentProfileID, err := h.getStudentProfileID(user.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Student profile not found")
	}

	var req SubmitQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	result, err := h.service.SubmitQuestionnaire(studentProfileID, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Kuesioner berhasil disubmit", result)
}

func (h *StudentQuestionnaireHandler) GetStudentRole(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	studentProfileID, err := h.getStudentProfileID(user.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Student profile not found")
	}

	result, err := h.service.GetStudentRole(studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Role siswa berhasil diambil", result)
}

func (h *StudentQuestionnaireHandler) GetQuestionnaireResult(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	studentProfileID, err := h.getStudentProfileID(user.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Student profile not found")
	}

	responseIDStr := c.Params("id")
	responseID, err := uuid.Parse(responseIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid response ID")
	}

	result, err := h.service.GetQuestionnaireResult(studentProfileID, responseID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Hasil kuesioner berhasil diambil", result)
}

func (h *StudentQuestionnaireHandler) GetLatestQuestionnaireResult(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	studentProfileID, err := h.getStudentProfileID(user.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Profil siswa tidak ditemukan")
	}

	result, err := h.service.GetLatestQuestionnaireResult(studentProfileID)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Hasil kuesioner terbaru berhasil diambil", result)
}

func (h *StudentQuestionnaireHandler) getStudentProfileID(userID uuid.UUID) (uuid.UUID, error) {
	return h.service.repo.GetStudentProfileIDByUserID(userID)
}
