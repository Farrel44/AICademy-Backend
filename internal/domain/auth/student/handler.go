package student

import (
	"encoding/csv"
	"fmt"

	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type StudentAuthHandler struct {
	service *StudentAuthService
}

func NewStudentAuthHandler(service *StudentAuthService) *StudentAuthHandler {
	return &StudentAuthHandler{service: service}
}

func (h *StudentAuthHandler) CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.CreateStudent(req)
	if err != nil {
		switch err.Error() {
		case "user with this email already exists":
			return utils.ErrorResponse(c, 409, "Email already registered")
		case "student with this NIS already exists":
			return utils.ErrorResponse(c, 409, "NIS already exists")
		case "failed to check NIS":
			return utils.ErrorResponse(c, 500, "Internal server error")
		case "failed to generate password":
			return utils.ErrorResponse(c, 500, "Internal server error")
		case "failed to create user account":
			return utils.ErrorResponse(c, 500, "Failed to create user account")
		case "failed to create student profile":
			return utils.ErrorResponse(c, 500, "Failed to create student profile")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c.Status(fiber.StatusCreated), commonAuth.MessageResponse{
		Message: "Student created successfully",
	}, "Student created successfully")
}

func (h *StudentAuthHandler) ChangeDefaultPassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}

	var req ChangeDefaultPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.ChangeDefaultPassword(user.UserID, req)
	if err != nil {
		switch err.Error() {
		case "password confirmation does not match":
			return utils.ErrorResponse(c, 422, "Password confirmation does not match")
		case "user not found":
			return utils.ErrorResponse(c, 404, "User not found")
		case "user is not a student":
			return utils.ErrorResponse(c, 403, "Access denied")
		case "current password is incorrect":
			return utils.ErrorResponse(c, 401, "Current password is incorrect")
		case "password has already been changed from default":
			return utils.ErrorResponse(c, 400, "Password has already been changed from default")
		case "failed to hash new password":
			return utils.ErrorResponse(c, 500, "Internal server error")
		case "failed to update password":
			return utils.ErrorResponse(c, 500, "Internal server error")
		case "failed to generate token":
			return utils.ErrorResponse(c, 500, "Internal server error")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, result, "Password changed successfully")
}

func (h *StudentAuthHandler) UploadStudentsCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, 400, "No file uploaded")
	}

	src, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to open file")
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to parse CSV file")
	}

	if len(records) < 2 {
		return utils.ErrorResponse(c, 400, "CSV file must contain header and at least one data row")
	}

	var successCount, failedCount int
	var errors []string

	// Skip header (first row)
	for i, record := range records[1:] {
		if len(record) < 4 {
			failedCount++
			errors = append(errors, fmt.Sprintf("Row %d: Insufficient columns", i+2))
			continue
		}

		req := CreateStudentRequest{
			NIS:      record[0],
			Class:    record[1],
			Email:    record[2],
			Fullname: record[3],
		}

		err := h.service.CreateStudent(req)
		if err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+2, err.Error()))
		} else {
			successCount++
		}
	}

	result := BulkCreateResult{
		SuccessCount: successCount,
		FailedCount:  failedCount,
		Errors:       errors,
	}

	return utils.SuccessResponse(c, result, "Bulk student creation completed")
}
