package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func SuccessResponse(c *fiber.Ctx, data interface{}, message string) error {
	return c.JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error:   message,
	})
}

func ValidationErrorResponse(c *fiber.Ctx, err error) error {
	var errors []ValidationErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationErrorDetail{
				Field:   e.Field(),
				Message: getValidationMessage(e),
			})
		}
	} else {
		return ErrorResponse(c, 400, err.Error())
	}

	return c.Status(400).JSON(Response{
		Success: false,
		Error:   "Validasi gagal",
		Data:    errors,
	})
}

func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Field ini wajib diisi"
	case "min":
		return "Field ini minimal " + e.Param() + " karakter"
	case "max":
		return "Field ini maksimal " + e.Param() + " karakter"
	case "oneof":
		return "Field ini harus salah satu dari: " + e.Param()
	case "uuid":
		return "Field ini harus berupa UUID yang valid"
	case "dive":
		return "Item array tidak valid"
	case "email":
		return "Format email tidak valid"
	default:
		return "Nilai tidak valid"
	}
}
