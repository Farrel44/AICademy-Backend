package project

import (
	"encoding/json"
	"strconv"
	"time"

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

func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req CreateProjectRequest

	req.ProjectName = c.FormValue("project_name")
	req.Description = c.FormValue("description")
	req.LinkURL = nil
	if linkURL := c.FormValue("link_url"); linkURL != "" {
		req.LinkURL = &linkURL
	}

	// Parse dates
	if startDateStr := c.FormValue("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = startDate
		} else {
			return utils.SendError(c, 400, "Invalid start_date format. Use YYYY-MM-DD")
		}
	}

	if endDateStr := c.FormValue("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = endDate
		} else {
			return utils.SendError(c, 400, "Invalid end_date format. Use YYYY-MM-DD")
		}
	}

	// Parse contributors JSON
	if contributorsStr := c.FormValue("contributors"); contributorsStr != "" {
		var contributors []CreateContributorRequest
		if err := json.Unmarshal([]byte(contributorsStr), &contributors); err == nil {
			req.Contributors = contributors
		} else {
			return utils.SendError(c, 400, "Invalid contributors format. Must be valid JSON array")
		}
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err == nil && form.File["photos"] != nil {
		req.Photos = form.File["photos"]
	}

	// Validate required fields manually
	if req.ProjectName == "" {
		return utils.SendError(c, 400, "project_name is required")
	}
	if req.Description == "" {
		return utils.SendError(c, 400, "description is required")
	}
	if req.StartDate.IsZero() {
		return utils.SendError(c, 400, "start_date is required")
	}
	if req.EndDate.IsZero() {
		return utils.SendError(c, 400, "end_date is required")
	}

	// Validate end date is after start date
	if req.EndDate.Before(req.StartDate) {
		return utils.SendError(c, 400, "end_date must be after start_date")
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
		return utils.SendError(c, 400, "Invalid project ID")
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
		return utils.SendError(c, 400, "Invalid project ID")
	}

	var req UpdateProjectRequest

	// Parse form data manually
	if projectName := c.FormValue("project_name"); projectName != "" {
		req.ProjectName = projectName
	}
	if description := c.FormValue("description"); description != "" {
		req.Description = description
	}
	if linkURL := c.FormValue("link_url"); linkURL != "" {
		req.LinkURL = &linkURL
	}

	// Parse dates
	if startDateStr := c.FormValue("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			req.StartDate = &startDate
		} else {
			return utils.SendError(c, 400, "Invalid start_date format. Use YYYY-MM-DD")
		}
	}

	if endDateStr := c.FormValue("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			req.EndDate = &endDate
		} else {
			return utils.SendError(c, 400, "Invalid end_date format. Use YYYY-MM-DD")
		}
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
		return utils.SendError(c, 400, "Invalid project ID")
	}

	if err := h.service.DeleteProject(id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Project berhasil dihapus", nil)
}

// Certification handlers
func (h *ProjectHandler) CreateCertification(c *fiber.Ctx) error {
	var req CreateCertificationRequest

	// Parse form data manually
	req.Name = c.FormValue("name")
	req.IssuingOrganization = c.FormValue("issuing_organization")

	if credentialID := c.FormValue("credential_id"); credentialID != "" {
		req.CredentialID = &credentialID
	}
	if credentialURL := c.FormValue("credential_url"); credentialURL != "" {
		req.CredentialURL = &credentialURL
	}

	// Parse dates
	if issueDateStr := c.FormValue("issue_date"); issueDateStr != "" {
		if issueDate, err := time.Parse("2006-01-02", issueDateStr); err == nil {
			req.IssueDate = issueDate
		} else {
			return utils.SendError(c, 400, "Invalid issue_date format. Use YYYY-MM-DD")
		}
	}

	if expirationDateStr := c.FormValue("expiration_date"); expirationDateStr != "" {
		if expirationDate, err := time.Parse("2006-01-02", expirationDateStr); err == nil {
			req.ExpirationDate = &expirationDate
		}
	}

	// Get files from form
	form, err := c.MultipartForm()
	if err == nil && form.File["photos"] != nil {
		req.Photos = form.File["photos"]
	}

	// Validate required fields manually
	if req.Name == "" {
		return utils.SendError(c, 400, "name is required")
	}
	if req.IssuingOrganization == "" {
		return utils.SendError(c, 400, "issuing_organization is required")
	}
	if req.IssueDate.IsZero() {
		return utils.SendError(c, 400, "issue_date is required")
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
		return utils.SendError(c, 400, "Invalid certification ID")
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
		return utils.SendError(c, 400, "Invalid certification ID")
	}

	var req UpdateCertificationRequest

	// Parse form data manually
	if name := c.FormValue("name"); name != "" {
		req.Name = &name
	}
	if issuingOrg := c.FormValue("issuing_organization"); issuingOrg != "" {
		req.IssuingOrganization = &issuingOrg
	}
	if credentialID := c.FormValue("credential_id"); credentialID != "" {
		req.CredentialID = &credentialID
	}
	if credentialURL := c.FormValue("credential_url"); credentialURL != "" {
		req.CredentialURL = &credentialURL
	}

	// Parse dates
	if issueDateStr := c.FormValue("issue_date"); issueDateStr != "" {
		if issueDate, err := time.Parse("2006-01-02", issueDateStr); err == nil {
			req.IssueDate = &issueDate
		} else {
			return utils.SendError(c, 400, "Invalid issue_date format. Use YYYY-MM-DD")
		}
	}

	if expirationDateStr := c.FormValue("expiration_date"); expirationDateStr != "" {
		if expirationDate, err := time.Parse("2006-01-02", expirationDateStr); err == nil {
			req.ExpirationDate = &expirationDate
		}
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
		return utils.SendError(c, 400, "Invalid certification ID")
	}

	if err := h.service.DeleteCertification(id); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SendSuccess(c, "Sertifikasi berhasil dihapus", nil)
}
