package student_challenge

import (
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	var req SearchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 10
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	students, err := h.service.SearchStudents(claims, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Students retrieved successfully", students)
}

func (h *StudentChallengeHandler) CreateTeam(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	var req CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	team, err := h.service.CreateTeam(claims, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Team created successfully", team)
}

func (h *StudentChallengeHandler) GetMyTeams(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	teams, err := h.service.GetMyTeams(claims)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Teams retrieved successfully", teams)
}

func (h *StudentChallengeHandler) GetAvailableChallenges(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	search := c.Query("search", "")

	challenges, err := h.service.GetAvailableChallenges(claims, page, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data tantangan berhasil diambil", challenges)
}

func (h *StudentChallengeHandler) RegisterTeamToChallenge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	var req RegisterChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.RegisterTeamToChallenge(claims, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, result.Message, result)
}

func (h *StudentChallengeHandler) AutoRegisterToChallenge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	var req AutoRegisterChallengeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.AutoRegisterToChallenge(claims, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Team berhasil terdaftar untuk challenge", result)
}

func (h *StudentChallengeHandler) GetChallengeByID(c *fiber.Ctx) error {
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

func (h *StudentChallengeHandler) SubmitChallenge(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return utils.SendError(c, 400, "Failed to parse form data")
	}

	// Get challenge_id from form
	challengeIDs := form.Value["challenge_id"]
	if len(challengeIDs) == 0 {
		return utils.SendError(c, 400, "challenge_id is required")
	}

	challengeID, err := uuid.Parse(challengeIDs[0])
	if err != nil {
		return utils.SendError(c, 400, "Invalid challenge_id format")
	}

	// Get title from form
	titles := form.Value["title"]
	if len(titles) == 0 {
		return utils.SendError(c, 400, "title is required")
	}

	req := &SubmitChallengeRequest{
		ChallengeID: challengeID,
		Title:       titles[0],
	}

	// Get optional repo_url
	if repoURLs := form.Value["repo_url"]; len(repoURLs) > 0 && repoURLs[0] != "" {
		req.RepoURL = &repoURLs[0]
	}

	// Get files
	if files := form.File["docs_file"]; len(files) > 0 {
		req.DocsFile = files[0]
	}

	if files := form.File["image_file"]; len(files) > 0 {
		req.ImageFile = files[0]
	}

	if err := utils.ValidateStruct(*req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.SubmitChallenge(claims, req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Submission berhasil", result)
}

func (h *StudentChallengeHandler) GetMySubmissions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role").(string)

	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	submissions, err := h.service.GetMySubmissions(claims, page, limit)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Submissions retrieved successfully", submissions)
}
