package auth

import (
	"aicademy-backend/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service   *AuthService
	validator *validator.Validate
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service:   service,
		validator: validator.New(),
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ErrorResponse(c, 400, "Validation failed: "+err.Error())
	}

	result, err := h.service.Register(req)
	if err != nil {
		return utils.ErrorResponse(c, 400, err.Error())
	}

	return utils.SuccessResponse(c, result, "Registration successful")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		return utils.ErrorResponse(c, 400, "Validation failed: "+err.Error())
	}

	result, err := h.service.Login(req)
	if err != nil {
		return utils.ErrorResponse(c, 401, err.Error())
	}

	return utils.SuccessResponse(c, result, "Login successful")
}
