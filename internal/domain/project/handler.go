package project

import (
	"strconv"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	service *ProjectService
}

func NewProjectHandler(service *ProjectService) *ProjectHandler {
	return &ProjectHandler{
		service: service,
	}
}

// Project handlers
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req CreateProjectRequest

	// Parse multipart form
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err == nil && form.File["photos"] != nil {
		req.Photos = form.File["photos"]
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	project, err := h.service.CreateProject(c, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Project berhasil dibuat", project)
}

func (h *ProjectHandler) GetProjectByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid project ID")
	}

	project, err := h.service.GetProjectByID(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Project tidak ditemukan")
	}

	return utils.SendSuccess(c, "Project berhasil diambil", project)
}

func (h *ProjectHandler) GetMyProjects(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	projects, err := h.service.GetMyProjects(c, page, limit, search)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Data proyek berhasil diambil", projects)
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid project ID")
	}

	var req UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err == nil && form.File["photos"] != nil {
		req.Photos = form.File["photos"]
	}

	project, err := h.service.UpdateProject(id, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Project berhasil diupdate", project)
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid project ID")
	}

	if err := h.service.DeleteProject(id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Project berhasil dihapus", nil)
}

func (h *ProjectHandler) AddProjectContributor(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid project ID")
	}

	var req AddProjectContributorRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	if err := h.service.AddProjectContributor(c, projectID, &req); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Contributor berhasil ditambahkan", nil)
}

// Certification handlers
func (h *ProjectHandler) CreateCertification(c *fiber.Ctx) error {
	var req CreateCertificationRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err == nil && form.File["photos"] != nil {
		req.Photos = form.File["photos"]
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	certification, err := h.service.CreateCertification(c, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Sertifikasi berhasil dibuat", certification)
}

func (h *ProjectHandler) GetCertificationByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid certification ID")
	}

	certification, err := h.service.GetCertificationByID(id)
	if err != nil {
		return utils.SendError(c, fiber.StatusNotFound, "Sertifikasi tidak ditemukan")
	}

	return utils.SendSuccess(c, "Sertifikasi berhasil diambil", certification)
}

func (h *ProjectHandler) GetMyCertifications(c *fiber.Ctx) error {
	certifications, err := h.service.GetMyCertifications(c)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Sertifikasi berhasil diambil", certifications)
}

func (h *ProjectHandler) UpdateCertification(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid certification ID")
	}

	var req UpdateCertificationRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err == nil && form.File["photos"] != nil {
		req.Photos = form.File["photos"]
	}

	certification, err := h.service.UpdateCertification(id, &req)
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Sertifikasi berhasil diupdate", certification)
}

func (h *ProjectHandler) DeleteCertification(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid certification ID")
	}

	if err := h.service.DeleteCertification(id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Sertifikasi berhasil dihapus", nil)
}
