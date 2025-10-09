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
	userId, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get user id")
	}

	user, err := h.service.repo.GetUserByID(userId)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to get user data")
	}

	if user.Role == RoleStudent {
		enhancedUser, err := h.service.GetStudentWithRecommendedRole(c)
		if err != nil {
			return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
		}
		return utils.SendSuccess(c, "Data siswa berhasil diambil", enhancedUser)
	}

	return utils.SendSuccess(c, "Data user berhasil diambil", user)
}

func (h *UserHandler) UpdateUserProfile(c *fiber.Ctx) error {
	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	user, err := h.service.UpdateUserProfile(c, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SendSuccess(c, "Data siswa berhasil di update", user)
}
