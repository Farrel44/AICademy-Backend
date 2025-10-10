package student_challenge

import (
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type StudentChallengeHandler struct {
	service *StudentChallengeService
}

func NewStudentChallengeHandler(service *StudentChallengeService) *StudentChallengeHandler {
	return &StudentChallengeHandler{
		service: service,
	}
}

func (h *StudentChallengeHandler) SearchStudents(c *fiber.Ctx) error {
	var req SearchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 10
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	students, err := h.service.SearchStudents(c, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Students retrieved successfully", students)
}

func (h *StudentChallengeHandler) CreateTeam(c *fiber.Ctx) error {
	var req CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	team, err := h.service.CreateTeam(c, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Team created successfully", team)
}

func (h *StudentChallengeHandler) GetMyTeams(c *fiber.Ctx) error {
	teams, err := h.service.GetMyTeams(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Teams retrieved successfully", teams)
}

func (h *StudentChallengeHandler) GetAvailableChallenges(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")

	challenges, err := h.service.GetAvailableChallenges(c, page, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data tantangan berhasil diambil", challenges)
}

func (h *StudentChallengeHandler) RegisterTeamToChallenge(c *fiber.Ctx) error {
	var req RegisterChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.RegisterTeamToChallenge(c, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, result.Message, result)
}
