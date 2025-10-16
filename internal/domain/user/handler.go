package user

import (
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GetUserByToken(c *fiber.Ctx) error {
	// Get enhanced profile with projects and certifications for /me endpoint
	enhancedUser, err := h.service.GetEnhancedUserProfile(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data user berhasil diambil", enhancedUser)
}

func (h *UserHandler) GetPublicStudentProfileByNIS(c *fiber.Ctx) error {
	nis := c.Params("nis")
	if nis == "" {
		return utils.SendError(c, 400, "NIS is required")
	}

	if len(nis) < 4 || len(nis) > 20 {
		return utils.SendError(c, 400, "Invalid NIS format")
	}

	profile, err := h.service.GetPublicStudentProfileByNIS(nis)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Student profile not found")
	}

	return utils.SendSuccess(c, "Student profile retrieved successfully", profile)
}

func (h *UserHandler) UpdateUserProfile(c *fiber.Ctx) error {
	user, err := h.service.UpdateUserProfile(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SendSuccess(c, "Data siswa berhasil di update", user)
}
