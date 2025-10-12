package teacher

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherHandler struct {
	service *TeacherService
}

func NewTeacherHandler(service *TeacherService) *TeacherHandler {
	return &TeacherHandler{service: service}
}

func (h *TeacherHandler) GetPendingSubmissions(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	teacherClaims := c.Locals("user")
	if teacherClaims == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid teacher claims")
	}

	var teacherID uuid.UUID
	if claims, ok := teacherClaims.(*middleware.UserClaims); ok {
		teacherID = claims.UserID
	}

	if teacherID == uuid.Nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid teacher ID")
	}

	result, err := h.service.GetPendingSubmissions(teacherID, page, limit, search)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, result, "Pending submissions retrieved successfully")
}

func (h *TeacherHandler) ReviewSubmission(c *fiber.Ctx) error {
	submissionID, err := uuid.Parse(c.Params("submissionId"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	var req ReviewSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	if err := utils.ValidateStruct(&req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	teacherClaims := c.Locals("user")
	if teacherClaims == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid teacher claims")
	}

	var teacherID uuid.UUID
	if claims, ok := teacherClaims.(*middleware.UserClaims); ok {
		teacherID = claims.UserID
	}

	if teacherID == uuid.Nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid teacher ID")
	}

	result, err := h.service.ReviewSubmission(submissionID, teacherID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, result, "Submission reviewed successfully")
}
