package pkl

import (
	"fmt"
	"strconv"

	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherPklHandler struct {
	service *TeacherPklService
}

func NewTeacherPklHandler(service *TeacherPklService) *TeacherPklHandler {
	return &TeacherPklHandler{
		service: service,
	}
}

func (h *TeacherPklHandler) GetAllInternships(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.service.GetInternshipPositions(page, limit, search)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to get internship positions")
	}

	return utils.SuccessResponse(c, result, "Internship positions retrieved successfully")
}

func (h *TeacherPklHandler) GetInternshipsWithSubmissionsByCompanyID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	companyID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid company ID")
	}

	internships, err := h.service.GetInternshipsWithSubmissionsByCompanyID(companyID)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to get internships with submissions")
	}

	return utils.SuccessResponse(c, internships, "Internships with submissions retrieved successfully")
}

func (h *TeacherPklHandler) GetSubmissionByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	submissionID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid submission ID")
	}

	submission, err := h.service.GetSubmissionByID(submissionID)
	if err != nil {
		return utils.ErrorResponse(c, 404, "Submission not found")
	}

	return utils.SuccessResponse(c, submission, "Submission retrieved successfully")
}

func (h *TeacherPklHandler) GetApplicationByID(c *fiber.Ctx) error {
	applicationIDParam := c.Params("id")
	applicationID, err := uuid.Parse(applicationIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid application ID")
	}

	application, err := h.service.GetApplicationByID(applicationID)
	if err != nil {
		return utils.ErrorResponse(c, 404, "Application not found")
	}

	return utils.SuccessResponse(c, application, "Application retrieved successfully")
}

func (h *TeacherPklHandler) UpdateApplicationStatus(c *fiber.Ctx) error {
	applicationIDParam := c.Params("id")
	applicationID, err := uuid.Parse(applicationIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid application ID")
	}

	var req UpdateApplicationStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	userIDRaw, ok := c.Locals("user_id").(string)
	if !ok || userIDRaw == "" {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}
	teacherID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}

	if err := h.service.UpdateApplicationStatus(applicationID, req.Status, teacherID); err != nil {
		switch err.Error() {
		case "invalid status":
			return utils.ErrorResponse(c, 400, "Invalid status value")
		case "submission not found":
			return utils.ErrorResponse(c, 404, "Application not found")
		case "failed to update submission status":
			return utils.ErrorResponse(c, 500, "Failed to update application status")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	statusMessage := "approved"
	if req.Status == "rejected" {
		statusMessage = "rejected"
	}

	return utils.SuccessResponse(c, commonAuth.MessageResponse{
		Message: fmt.Sprintf("Application %s successfully", statusMessage),
	}, fmt.Sprintf("Application %s successfully", statusMessage))
}

func (h *TeacherPklHandler) GetInternshipApplications(c *fiber.Ctx) error {
	internshipIDParam := c.Params("id")
	internshipID, err := uuid.Parse(internshipIDParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	applications, err := h.service.GetInternshipApplications(internshipID)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to get applications")
	}

	return utils.SuccessResponse(c, applications, "Applications retrieved successfully")
}

func (h *TeacherPklHandler) GetSubmissionsByInternshipID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	internshipID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	submissions, err := h.service.GetSubmissionsByInternshipID(internshipID)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to get submissions")
	}

	return utils.SuccessResponse(c, submissions, "Submissions retrieved successfully")
}

func (h *TeacherPklHandler) UpdateSubmissionStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")
	submissionID, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid submission ID")
	}

	var req UpdateSubmissionStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	userIDRaw, ok := c.Locals("user_id").(string)
	if !ok || userIDRaw == "" {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}
	teacherID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}

	if err := h.service.UpdateSubmissionStatus(submissionID, req.Status, teacherID); err != nil {
		switch err.Error() {
		case "invalid status":
			return utils.ErrorResponse(c, 400, "Invalid status value")
		case "submission not found":
			return utils.ErrorResponse(c, 404, "Submission not found")
		case "failed to update submission status":
			return utils.ErrorResponse(c, 500, "Failed to update submission status")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	statusMessage := "approved"
	if req.Status == "rejected" {
		statusMessage = "rejected"
	}

	return utils.SuccessResponse(c, commonAuth.MessageResponse{
		Message: fmt.Sprintf("Submission %s successfully", statusMessage),
	}, fmt.Sprintf("Submission %s successfully", statusMessage))
}
