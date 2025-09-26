package auth

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func (h *AuthHandler) formatValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErr {
			var message string

			switch fieldError.Tag() {
			case "required":
				message = fieldError.Field() + " is required"
			case "email":
				message = "Invalid email format"
			case "min":
				message = fieldError.Field() + " must be at least " + fieldError.Param() + " characters"
			case "max":
				message = fieldError.Field() + " must be at most " + fieldError.Param() + " characters"
			case "regexp":
				if strings.Contains(strings.ToLower(fieldError.Field()), "password") {
					message = "Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character"
				} else {
					message = fieldError.Field() + " format is invalid"
				}
			default:
				message = fieldError.Field() + " is invalid"
			}

			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(fieldError.Field()),
				Message: message,
			})
		}
	}

	return validationErrors
}

func (h *AuthHandler) RegisterAlumni(c *fiber.Ctx) error {
	var req RegisterAlumniRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	result, err := h.service.RegisterAlumni(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    result.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    result.User.Role,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})

	return utils.SuccessResponse(c, result, "Alumni registration successful")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	result, err := h.service.Login(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    result.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    result.User.Role,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})

	return utils.SuccessResponse(c, result, "Login successful")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Logout successful",
	}, "Logout successful")
}

func (h *AuthHandler) CreateTeacher(c *fiber.Ctx) error {
	var req CreateTeacherRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	err := h.service.CreateTeacher(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Teacher created successfully",
	}, "Teacher created successfully")
}

func (h *AuthHandler) CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	err := h.service.CreateStudent(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: fmt.Sprintf("Student created successfully with default password: %s", DefaultStudentPassword),
	}, "Student created successfully")
}

func (h *AuthHandler) CreateCompany(c *fiber.Ctx) error {
	var req CreateCompanyRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	err := h.service.CreateCompany(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Company created successfully",
	}, "Company created successfully")
}

func (h *AuthHandler) UploadStudentsCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("csv_file")
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "CSV file is required")
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return utils.ErrorResponse(c, http.StatusBadRequest, "File must be a CSV")
	}

	src, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file")
	}
	defer src.Close()

	csvReader := csv.NewReader(src)
	csvReader.FieldsPerRecord = 4

	createdCount, validationErrors, err := h.service.CreateStudentsFromCSV(csvReader)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	response := map[string]interface{}{
		"created_count": createdCount,
		"message":       fmt.Sprintf("Successfully created %d students with default password: %s", createdCount, DefaultStudentPassword),
	}

	if len(validationErrors) > 0 {
		response["validation_errors"] = validationErrors
		response["message"] = fmt.Sprintf("Created %d students, but %d rows had errors", createdCount, len(validationErrors))
	}

	return utils.SuccessResponse(c, response, "CSV upload processed")
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	err := h.service.ChangePassword(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Password changed successfully",
	}, "Password changed successfully")
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	err := h.service.ForgotPassword(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "You will receive a reset email if user with that email exists",
	}, "Reset link sent")
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Reset token is required")
	}

	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validation failed",
			Details: validationErrors,
		})
	}

	err := h.service.ResetPassword(token, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Password reset successfully",
	}, "Password reset successfully")
}
