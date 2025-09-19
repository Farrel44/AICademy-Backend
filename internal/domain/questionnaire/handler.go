package questionnaire

import (
	"aicademy-backend/internal/utils"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QuestionnaireHandler struct {
	service   *QuestionnaireService
	validator *validator.Validate
}

func NewQuestionnaireHandler(service *QuestionnaireService) *QuestionnaireHandler {
	return &QuestionnaireHandler{
		service:   service,
		validator: validator.New(),
	}
}

func (h *QuestionnaireHandler) GetActiveQuestionnaire(c *fiber.Ctx) error {
	questionnaire, err := h.service.GetActiveQuestionnaire()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Kuesioner aktif tidak ditemukan")
	}

	return utils.SuccessResponse(c, questionnaire, "Kuesioner aktif berhasil diambil")
}

func (h *QuestionnaireHandler) SubmitQuestionnaire(c *fiber.Ctx) error {
	studentIDStr := c.Locals("user_id").(string)
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID siswa tidak valid")
	}

	var req SubmitQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	result, err := h.service.SubmitQuestionnaire(studentID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, result, "Kuesioner berhasil dikirim")
}

func (h *QuestionnaireHandler) GetQuestionnaireResult(c *fiber.Ctx) error {
	responseIDStr := c.Params("responseId")
	responseID, err := uuid.Parse(responseIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID respons tidak valid")
	}

	result, err := h.service.GetQuestionnaireResult(responseID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Hasil tidak ditemukan")
	}

	return utils.SuccessResponse(c, result, "Hasil berhasil diambil")
}

func (h *QuestionnaireHandler) GetLatestResultByStudent(c *fiber.Ctx) error {
	studentIDStr := c.Locals("user_id").(string)
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID siswa tidak valid")
	}

	result, err := h.service.GetLatestResultByStudent(studentID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Tidak ada hasil ditemukan")
	}

	return utils.SuccessResponse(c, result, "Hasil terbaru berhasil diambil")
}

func (h *QuestionnaireHandler) GenerateQuestionnaire(c *fiber.Ctx) error {
	var req GenerateQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	result, err := h.service.GenerateQuestionnaire(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, result, "Generasi kuesioner dimulai")
}

func (h *QuestionnaireHandler) GetGenerationStatus(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	status, err := h.service.GetGenerationStatus(questionnaireID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Kuesioner tidak ditemukan")
	}

	return utils.SuccessResponse(c, status, "Status generasi berhasil diambil")
}

func (h *QuestionnaireHandler) GetAllQuestionnaires(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	questionnaires, total, err := h.service.GetAllQuestionnaires(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil kuesioner")
	}

	response := map[string]interface{}{
		"questionnaires": questionnaires,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	return utils.SuccessResponse(c, response, "Kuesioner berhasil diambil")
}

func (h *QuestionnaireHandler) GetQuestionnaireByID(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	questionnaire, err := h.service.GetQuestionnaireByID(questionnaireID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Kuesioner tidak ditemukan")
	}

	return utils.SuccessResponse(c, questionnaire, "Kuesioner berhasil diambil")
}

func (h *QuestionnaireHandler) ActivateQuestionnaire(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	err = h.service.ActivateQuestionnaire(questionnaireID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengaktifkan kuesioner")
	}

	return utils.SuccessResponse(c, nil, "Kuesioner berhasil diaktifkan")
}

func (h *QuestionnaireHandler) DeactivateQuestionnaire(c *fiber.Ctx) error {
	err := h.service.DeactivateAllQuestionnaires()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menonaktifkan kuesioner")
	}

	return utils.SuccessResponse(c, nil, "Semua kuesioner berhasil dinonaktifkan")
}

func (h *QuestionnaireHandler) DeleteQuestionnaire(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	err = h.service.DeleteQuestionnaire(questionnaireID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus kuesioner")
	}

	return utils.SuccessResponse(c, nil, "Kuesioner berhasil dihapus")
}

func (h *QuestionnaireHandler) CloneQuestionnaire(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	var req struct {
		Name string `json:"name" validate:"required,min=2"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	clone, err := h.service.CloneQuestionnaire(questionnaireID, req.Name)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, clone, "Kuesioner berhasil digandakan")
}

func (h *QuestionnaireHandler) GetQuestionnaireResponses(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	responses, total, err := h.service.GetQuestionnaireResponses(questionnaireID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil respons")
	}

	response := map[string]interface{}{
		"responses": responses,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	return utils.SuccessResponse(c, response, "Respons berhasil diambil")
}

func (h *QuestionnaireHandler) GetQuestionnaireStats(c *fiber.Ctx) error {
	stats, err := h.service.GetQuestionnaireStats()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil statistik")
	}

	return utils.SuccessResponse(c, stats, "Statistik berhasil diambil")
}

func (h *QuestionnaireHandler) GetResponseAnalytics(c *fiber.Ctx) error {
	var questionnaireID *uuid.UUID
	if id := c.Query("questionnaire_id"); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
		}
		questionnaireID = &parsed
	}

	analytics, err := h.service.GetResponseAnalytics(questionnaireID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil analitik")
	}

	return utils.SuccessResponse(c, analytics, "Analitik berhasil diambil")
}

func (h *QuestionnaireHandler) SearchQuestionnaires(c *fiber.Ctx) error {
	keyword := c.Query("q")
	if keyword == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Kata kunci pencarian diperlukan")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	questionnaires, total, err := h.service.SearchQuestionnaires(keyword, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Pencarian gagal")
	}

	response := map[string]interface{}{
		"questionnaires": questionnaires,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"keyword": keyword,
	}

	return utils.SuccessResponse(c, response, "Hasil pencarian berhasil diambil")
}

func (h *QuestionnaireHandler) GetQuestionnairesByType(c *fiber.Ctx) error {
	generatedBy := c.Query("generated_by", "ai")
	if generatedBy != "ai" && generatedBy != "manual" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Tipe generasi tidak valid. Harus 'ai' atau 'manual'")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	questionnaires, total, err := h.service.GetQuestionnairesByGeneratedBy(generatedBy, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil kuesioner")
	}

	response := map[string]interface{}{
		"questionnaires": questionnaires,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"generated_by": generatedBy,
	}

	return utils.SuccessResponse(c, response, "Kuesioner berhasil diambil")
}

func (h *QuestionnaireHandler) AddQuestionToQuestionnaire(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	var req struct {
		QuestionText  string           `json:"question_text" validate:"required,min=5"`
		QuestionType  QuestionType     `json:"question_type" validate:"required,oneof=mcq likert case text"`
		Options       []QuestionOption `json:"options,omitempty"`
		Category      string           `json:"category" validate:"required"`
		QuestionOrder int              `json:"question_order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	if req.QuestionType == QuestionTypeMCQ && len(req.Options) == 0 {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Pertanyaan MCQ harus memiliki pilihan")
	}

	err = h.service.AddQuestionToQuestionnaire(questionnaireID, req.QuestionText, req.QuestionType, req.Options, req.Category, req.QuestionOrder)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Pertanyaan berhasil ditambahkan")
}

func (h *QuestionnaireHandler) UpdateQuestion(c *fiber.Ctx) error {
	questionIDStr := c.Params("questionId")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID pertanyaan tidak valid")
	}

	var req struct {
		QuestionText  string           `json:"question_text" validate:"required,min=5"`
		QuestionType  QuestionType     `json:"question_type" validate:"required,oneof=mcq likert case text"`
		Options       []QuestionOption `json:"options,omitempty"`
		Category      string           `json:"category" validate:"required"`
		QuestionOrder int              `json:"question_order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	err = h.service.UpdateQuestion(questionID, req.QuestionText, req.QuestionType, req.Options, req.Category, req.QuestionOrder)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Pertanyaan berhasil diperbarui")
}

func (h *QuestionnaireHandler) DeleteQuestion(c *fiber.Ctx) error {
	questionIDStr := c.Params("questionId")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID pertanyaan tidak valid")
	}

	err = h.service.DeleteQuestion(questionID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus pertanyaan")
	}

	return utils.SuccessResponse(c, nil, "Pertanyaan berhasil dihapus")
}

func (h *QuestionnaireHandler) UpdateQuestionOrder(c *fiber.Ctx) error {
	questionnaireIDStr := c.Params("questionnaireId")
	questionnaireID, err := uuid.Parse(questionnaireIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID kuesioner tidak valid")
	}

	var req struct {
		Questions []struct {
			ID    uuid.UUID `json:"id" validate:"required"`
			Order int       `json:"order" validate:"required,min=1"`
		} `json:"questions" validate:"required,dive"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	err = h.service.UpdateQuestionOrder(questionnaireID, req.Questions)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Urutan pertanyaan berhasil diperbarui")
}

func (h *QuestionnaireHandler) CreateQuestionTemplate(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name" validate:"required,min=2"`
		Description string `json:"description"`
		Prompt      string `json:"prompt" validate:"required,min=10"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.validator.Struct(req); err != nil {
		errors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(utils.Response{
			Success: false,
			Error:   "Validasi gagal",
			Data:    errors,
		})
	}

	template, err := h.service.CreateQuestionTemplate(req.Name, req.Description, req.Prompt)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, template, "Template berhasil dibuat")
}

func (h *QuestionnaireHandler) GetQuestionTemplates(c *fiber.Ctx) error {
	templates, err := h.service.GetQuestionTemplates()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil template")
	}

	return utils.SuccessResponse(c, templates, "Template berhasil diambil")
}

func (h *QuestionnaireHandler) formatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: h.getValidationMessage(e),
			})
		}
	}

	return errors
}

func (h *QuestionnaireHandler) getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "Field ini wajib diisi"
	case "min":
		return "Field ini minimal " + e.Param() + " karakter"
	case "max":
		return "Field ini maksimal " + e.Param() + " karakter"
	case "oneof":
		return "Field ini harus salah satu dari: " + e.Param()
	case "uuid":
		return "Field ini harus berupa UUID yang valid"
	case "dive":
		return "Item array tidak valid"
	default:
		return "Nilai tidak valid"
	}
}
