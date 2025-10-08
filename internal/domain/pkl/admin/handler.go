package pkl

import (
	"fmt"
	"strconv"

	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PklHandler struct {
	service *AdminPklService
}

func NewPklHandler(service *AdminPklService) *PklHandler {
	return &PklHandler{service: service}
}

func (h *PklHandler) CreateInternshipPosition(c *fiber.Ctx) error {
	var req CreateInternshipRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	internship, err := h.service.CreateInternshipPosition(&req)
	if err != nil {
		switch err.Error() {
		case "Companies Not Found":
			return utils.ErrorResponse(c, 404, "Company not found")
		case "failed to create intership position":
			return utils.ErrorResponse(c, 500, "Failed to create internship position")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c.Status(fiber.StatusCreated), internship, "Internship position created successfully")
}

func (h *PklHandler) GetInternshipPositions(c *fiber.Ctx) error {
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

func (h *PklHandler) GetInternshipByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	internship, err := h.service.GetInternshipByID(id)
	if err != nil {
		return utils.ErrorResponse(c, 404, "Internship position not found")
	}

	return utils.SuccessResponse(c, internship, "Internship position retrieved successfully")
}

func (h *PklHandler) UpdateInternshipPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	var req UpdateInternshipRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	updatedInternship := h.service.UpdateInternshipPosition(id, &req)
	if err != nil {
		switch err.Error() {
		case "internship position not found":
			return utils.ErrorResponse(c, 404, "Internship position not found")
		case "company not found":
			return utils.ErrorResponse(c, 404, "Company not found")
		case "failed to update internship position":
			return utils.ErrorResponse(c, 500, "Failed to update internship position")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, updatedInternship, "Internship position updated successfully")
}

func (h *PklHandler) DeleteInternshipPosition(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid internship ID")
	}

	err = h.service.DeleteInternshipPosition(id)
	if err != nil {
		switch err.Error() {
		case "internship position not found":
			return utils.ErrorResponse(c, 404, "Internship position not found")
		case "failed to delete internship position":
			return utils.ErrorResponse(c, 500, "Failed to delete internship position")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, commonAuth.MessageResponse{
		Message: "Internship position deleted successfully",
	}, "Internship position deleted successfully")
}

func (h *PklHandler) GetSubmissionsByInternshipID(c *fiber.Ctx) error {
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

func (h *PklHandler) GetInternshipsWithSubmissionsByCompanyID(c *fiber.Ctx) error {
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

func (h *PklHandler) GetSubmissionByID(c *fiber.Ctx) error {
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

func (h *PklHandler) UpdateSubmissionStatus(c *fiber.Ctx) error {
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
	adminID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}

	if err := h.service.UpdateSubmissionStatus(submissionID, req.Status, adminID); err != nil {
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
