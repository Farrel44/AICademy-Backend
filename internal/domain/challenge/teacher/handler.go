package teacher_challenge

import (
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherChallengeHandler struct {
	service *TeacherChallengeService
}

func NewTeacherChallengeHandler(service *TeacherChallengeService) *TeacherChallengeHandler {
	return &TeacherChallengeHandler{
		service: service,
	}
}

func (h *TeacherChallengeHandler) CreateChallenge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	var req CreateChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	challenge, err := h.service.CreateChallenge(claims, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Challenge created successfully", challenge)
}

func (h *TeacherChallengeHandler) UpdateChallenge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	challengeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid challenge ID")
	}

	var req UpdateChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	challenge, err := h.service.UpdateChallenge(claims, challengeID, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Challenge updated successfully", challenge)
}

func (h *TeacherChallengeHandler) DeleteChallenge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	challengeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid challenge ID")
	}

	err = h.service.DeleteChallenge(claims, challengeID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Challenge deleted successfully", nil)
}

func (h *TeacherChallengeHandler) GetMyChallenges(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	challenges, err := h.service.GetMyChallenges(claims, page, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data tantangan berhasil diambil", challenges)
}

func (h *TeacherChallengeHandler) GetChallengeByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	challengeID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid challenge ID")
	}

	challenge, err := h.service.GetChallengeByID(claims, challengeID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Challenge retrieved successfully", challenge)
}

func (h *TeacherChallengeHandler) GetMySubmissions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var challengeID *uuid.UUID
	if challengeIDStr := c.Query("challenge_id"); challengeIDStr != "" {
		id, err := uuid.Parse(challengeIDStr)
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, "Invalid challenge ID")
		}
		challengeID = &id
	}

	submissions, err := h.service.GetMySubmissions(claims, page, limit, search, challengeID)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data submission berhasil diambil", submissions)
}

func (h *TeacherChallengeHandler) ScoreSubmission(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	var req ScoreSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ScoreSubmission(claims, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Submission scored successfully", nil)
}
