package questionnaire

import (
	"aicademy-backend/internal/middleware"
	"aicademy-backend/internal/utils"
	"fmt"
	"log"
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
	userClaims, ok := c.Locals("user").(*middleware.UserClaims)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak ditemukan")
	}

	var req SubmitQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	log.Printf("Submit request - QuestionnaireID: %s, User: %s, Answers count: %d",
		req.QuestionnaireID.String(), userClaims.UserID.String(), len(req.Answers))

	questionnaire, err := h.service.GetActiveQuestionnaire()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Tidak ada kuesioner aktif")
	}

	if questionnaire.ID != req.QuestionnaireID {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Kuesioner yang dipilih tidak aktif")
	}

	if len(req.Answers) != len(questionnaire.Questions) {
		return utils.ErrorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("Jumlah jawaban tidak sesuai. Expected: %d, Got: %d",
				len(questionnaire.Questions), len(req.Answers)))
	}

	response, err := h.service.SubmitQuestionnaire(userClaims.UserID, req)
	if err != nil {
		log.Printf("Error submitting questionnaire: %v", err)
		return utils.ErrorResponse(c, http.StatusInternalServerError,
			"Gagal memproses jawaban: "+err.Error())
	}

	return utils.SuccessResponse(c, SubmitQuestionnaireResponse{
		ResponseID:      response.ID,
		QuestionnaireID: response.QuestionnaireID,
		Message:         "Jawaban berhasil disimpan dan sedang diproses",
		ProcessingTime:  "2-5 detik",
	}, "Kuesioner berhasil disubmit")
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
	userClaims, ok := c.Locals("user").(*middleware.UserClaims)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak ditemukan")
	}

	log.Printf("Getting latest result for user: %s", userClaims.UserID.String())

	studentProfile, err := h.service.GetStudentProfileByUserID(userClaims.UserID)
	if err != nil {
		log.Printf("Error getting student profile: %v", err)
		return utils.ErrorResponse(c, http.StatusNotFound, "Profil siswa tidak ditemukan")
	}

	log.Printf("Student profile found: %s", studentProfile.ID.String())

	result, err := h.service.GetLatestResultByStudentProfile(studentProfile.ID)
	if err != nil {
		log.Printf("Error getting latest result: %v", err)
		return utils.ErrorResponse(c, http.StatusNotFound, "Belum ada hasil kuesioner yang tersedia")
	}

	return utils.SuccessResponse(c, result, "Hasil terbaru berhasil diambil")
}

func (h *QuestionnaireHandler) GenerateQuestionnaire(c *fiber.Ctx) error {
	var req GenerateQuestionnaireRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format request tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
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

func (h *QuestionnaireHandler) CreateRole(c *fiber.Ctx) error {
	var req struct {
		RoleName    string `json:"role_name" validate:"required,min=2,max=100"`
		Description string `json:"description" validate:"required,min=10"`
		Category    string `json:"category" validate:"required,min=2,max=50"`
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

	role, err := h.service.CreateRole(req.RoleName, req.Description, req.Category)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, role, "Role berhasil dibuat")
}

func (h *QuestionnaireHandler) GetAllRoles(c *fiber.Ctx) error {
	roles, err := h.service.GetAllRoles()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil daftar role")
	}

	return utils.SuccessResponse(c, roles, "Daftar role berhasil diambil")
}

func (h *QuestionnaireHandler) DeleteRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID role tidak valid")
	}

	err = h.service.DeleteRole(roleID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, nil, "Role berhasil dihapus")
}

func (h *QuestionnaireHandler) formatValidationErrors(err error) []ValidationError {
	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		var errorMsg string
		switch err.Tag() {
		case "required":
			errorMsg = fmt.Sprintf("%s wajib diisi", err.Field())
		case "min":
			errorMsg = fmt.Sprintf("%s minimal %s karakter", err.Field(), err.Param())
		case "max":
			errorMsg = fmt.Sprintf("%s maksimal %s karakter", err.Field(), err.Param())
		case "email":
			errorMsg = fmt.Sprintf("%s harus berupa email yang valid", err.Field())
		case "oneof":
			errorMsg = fmt.Sprintf("%s harus salah satu dari: %s", err.Field(), err.Param())
		default:
			errorMsg = fmt.Sprintf("%s tidak valid", err.Field())
		}

		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Message: errorMsg,
		})
	}
	return errors
}

func (h *QuestionnaireHandler) UpdateRole(c *fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "ID role tidak valid")
	}

	var req struct {
		RoleName    string `json:"role_name" validate:"required,min=2,max=100"`
		Description string `json:"description" validate:"required,min=10"`
		Category    string `json:"category" validate:"required,min=2,max=50"`
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

	role, err := h.service.UpdateRole(roleID, req.RoleName, req.Description, req.Category)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, role, "Role berhasil diupdate")
}
