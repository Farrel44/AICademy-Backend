package experience

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service  Service
	userRepo *user.UserRepository
}

func NewHandler(service Service, userRepo *user.UserRepository) *Handler {
	return &Handler{
		service:  service,
		userRepo: userRepo,
	}
}

// getStudentProfileID helper function to get student profile ID from token
func (h *Handler) getStudentProfileID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return uuid.Nil, err
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		return uuid.Nil, err
	}

	if user.StudentProfile == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusForbidden, "Student profile required")
	}

	return user.StudentProfile.ID, nil
}

func (h *Handler) CreateExperience(c *fiber.Ctx) error {
	// Get student profile ID from token
	studentProfileID, err := h.getStudentProfileID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	var req CreateExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
	}

	response, err := h.service.CreateExperience(studentProfileID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Experience created successfully",
		"data":    response,
	})
}

func (h *Handler) GetExperienceByID(c *fiber.Ctx) error {
	// Get student profile ID from token
	studentProfileID, err := h.getStudentProfileID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Parse experience ID from URL parameter
	experienceIDStr := c.Params("id")
	experienceID, err := uuid.Parse(experienceIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid experience ID",
		})
	}

	response, err := h.service.GetExperienceByID(experienceID, studentProfileID)
	if err != nil {
		if err.Error() == "experience not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Experience not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Experience retrieved successfully",
		"data":    response,
	})
}

func (h *Handler) GetExperiences(c *fiber.Ctx) error {
	// Get student profile ID from token
	studentProfileID, err := h.getStudentProfileID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	response, err := h.service.GetExperiencesByStudentID(studentProfileID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Experiences retrieved successfully",
		"data":        response.Data,
		"total":       response.Total,
		"page":        response.Page,
		"limit":       response.Limit,
		"total_pages": response.TotalPages,
	})
}

func (h *Handler) UpdateExperience(c *fiber.Ctx) error {
	// Get student profile ID from token
	studentProfileID, err := h.getStudentProfileID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Parse experience ID from URL parameter
	experienceIDStr := c.Params("id")
	experienceID, err := uuid.Parse(experienceIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid experience ID",
		})
	}

	var req UpdateExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate request
	if err := utils.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Validation failed",
			"error":   err.Error(),
		})
	}

	response, err := h.service.UpdateExperience(experienceID, studentProfileID, &req)
	if err != nil {
		if err.Error() == "experience not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Experience not found",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Experience updated successfully",
		"data":    response,
	})
}

func (h *Handler) DeleteExperience(c *fiber.Ctx) error {
	// Get student profile ID from token
	studentProfileID, err := h.getStudentProfileID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	// Parse experience ID from URL parameter
	experienceIDStr := c.Params("id")
	experienceID, err := uuid.Parse(experienceIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid experience ID",
		})
	}

	err = h.service.DeleteExperience(experienceID, studentProfileID)
	if err != nil {
		if err.Error() == "experience not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Experience not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Experience deleted successfully",
	})
}
