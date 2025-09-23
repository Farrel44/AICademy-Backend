package alumni

import (
	"aicademy-backend/internal/utils"

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
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.RegisterAlumni(req)
	if err != nil {
		switch err.Error() {
		case "user with this email already exists":
			return utils.ErrorResponse(c, 409, "Email already registered")
		case "failed to hash password":
			return utils.ErrorResponse(c, 500, "Internal server error")
		case "failed to create user account":
			return utils.ErrorResponse(c, 500, "Failed to create user account")
		case "failed to create alumni profile":
			return utils.ErrorResponse(c, 500, "Failed to create alumni profile")
		case "failed to generate authentication token":
			return utils.ErrorResponse(c, 500, "Internal server error")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, result, "Alumni registration successful")
}
