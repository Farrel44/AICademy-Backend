package student

import (
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetStudentDashboard(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*middleware.UserClaims)
	if userClaims == nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	dashboard, err := h.service.GetStudentDashboard(userClaims.UserID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Student dashboard data retrieved successfully", dashboard)
}
