package alumni

import (
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type AlumniAuthHandler struct {
	service *AlumniAuthService
}

func NewAlumniAuthHandler(service *AlumniAuthService) *AlumniAuthHandler {
	return &AlumniAuthHandler{service: service}
}

func (h *AlumniAuthHandler) RegisterAlumni(c *fiber.Ctx) error {
	var req RegisterAlumniRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.RegisterAlumni(req)
	if err != nil {
		switch err.Error() {
		case "user with this email already exists":
			return utils.SendError(c, 409, "Email already registered")
		case "failed to hash password":
			return utils.SendError(c, 500, "Internal server error")
		case "failed to create user account":
			return utils.SendError(c, 500, "Failed to create user account")
		case "failed to create alumni profile":
			return utils.SendError(c, 500, "Failed to create alumni profile")
		case "failed to generate authentication token":
			return utils.SendError(c, 500, "Internal server error")
		default:
			return utils.SendError(c, 500, "Internal server error")
		}
	}

	return utils.SendSuccess(c, "Alumni berhasil mendaftar", result)
}
