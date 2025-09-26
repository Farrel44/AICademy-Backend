package questionnaire

import (
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CommonQuestionnaireHandler struct {
	service *CommonQuestionnaireService
}

func NewCommonQuestionnaireHandler(service *CommonQuestionnaireService) *CommonQuestionnaireHandler {
	return &CommonQuestionnaireHandler{service: service}
}

// GetActiveQuestionnaire - Public endpoint, anyone can access
func (h *CommonQuestionnaireHandler) GetActiveQuestionnaire(c *fiber.Ctx) error {
	response, err := h.service.GetActiveQuestionnaire()
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Active questionnaire retrieved successfully", response)
}

// SubmitQuestionnaire - Students only
func (h *CommonQuestionnaireHandler) SubmitQuestionnaire(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid token")
	}

	var req SubmitQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if req.QuestionnaireID == uuid.Nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Questionnaire ID is required")
	}

	if len(req.Answers) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "Answers are required")
	}

	result, err := h.service.SubmitQuestionnaire(userID, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SendSuccess(c, "Questionnaire submitted successfully", result)
}

// GetQuestionnaireResult - Students can get their own results
func (h *CommonQuestionnaireHandler) GetQuestionnaireResult(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid token")
	}

	responseIDStr := c.Params("responseId")
	responseID, err := uuid.Parse(responseIDStr)
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid response ID")
	}

	result, err := h.service.GetQuestionnaireResult(userID, responseID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Questionnaire result retrieved successfully", result)
}

// GetLatestResult - Students can get their latest result
func (h *CommonQuestionnaireHandler) GetLatestResult(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Invalid token")
	}

	result, err := h.service.GetLatestResultByStudent(userID)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SendSuccess(c, "Latest questionnaire result retrieved successfully", result)
}

// SetupCommonQuestionnaireRoutes - Setup routes untuk common questionnaire endpoints
func SetupCommonQuestionnaireRoutes(router fiber.Router, handler *CommonQuestionnaireHandler) {
	questionnaire := router.Group("/questionnaire")

	// Public endpoints (no auth required)
	questionnaire.Get("/active", handler.GetActiveQuestionnaire)

	// Protected endpoints (require authentication)
	protected := questionnaire.Group("/protected")
	protected.Post("/submit", handler.SubmitQuestionnaire)
	protected.Get("/result/:responseId", handler.GetQuestionnaireResult)
	protected.Get("/latest", handler.GetLatestResult)
}
