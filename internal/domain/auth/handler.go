package auth

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service   *AuthService
	validator *validator.Validate
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service:   service,
		validator: validator.New(),
	}
}

func (h *AuthHandler) formatValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErr {
			var message string

			switch fieldError.Tag() {
			case "required":
				message = fieldError.Field() + " wajib diisi"
			case "email":
				message = "Format email tidak valid"
			case "min":
				message = fieldError.Field() + " minimal " + fieldError.Param() + " karakter"
			case "max":
				message = fieldError.Field() + " maksimal " + fieldError.Param() + " karakter"
			case "regexp":
				if strings.Contains(strings.ToLower(fieldError.Field()), "password") {
					message = "Password harus mengandung minimal satu huruf kecil, satu huruf besar, satu angka, dan satu karakter khusus"
				} else {
					message = "Format " + fieldError.Field() + " tidak valid"
				}
			default:
				message = fieldError.Field() + " tidak valid"
			}

			validationErrors = append(validationErrors, ValidationError{
				Field:   strings.ToLower(fieldError.Field()),
				Message: message,
			})
		}
	}

	return validationErrors
}

func (h *AuthHandler) RegisterAlumni(c *fiber.Ctx) error {
	var req RegisterAlumniRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	result, err := h.service.RegisterAlumni(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    result.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    result.User.Role,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})

	return utils.SuccessResponse(c, result, "Registrasi alumni berhasil")
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	result, err := h.service.Login(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    result.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    result.User.Role,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})

	return utils.SuccessResponse(c, result, "Login berhasil")
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Logout berhasil",
	}, "Logout berhasil")
}

func (h *AuthHandler) CreateTeacher(c *fiber.Ctx) error {
	var req CreateTeacherRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	err := h.service.CreateTeacher(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Guru berhasil dibuat",
	}, "Guru berhasil dibuat")
}

func (h *AuthHandler) CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	err := h.service.CreateStudent(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: fmt.Sprintf("Siswa berhasil dibuat dengan password default: %s", DefaultStudentPassword),
	}, "Siswa berhasil dibuat")
}

func (h *AuthHandler) CreateCompany(c *fiber.Ctx) error {
	var req CreateCompanyRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	err := h.service.CreateCompany(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Perusahaan berhasil dibuat",
	}, "Perusahaan berhasil dibuat")
}

func (h *AuthHandler) UploadStudentsCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("csv_file")
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "CSV file is required")
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return utils.ErrorResponse(c, http.StatusBadRequest, "File must be a CSV")
	}

	src, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to open file")
	}
	defer src.Close()

	csvReader := csv.NewReader(src)
	csvReader.FieldsPerRecord = 4

	createdCount, validationErrors, err := h.service.CreateStudentsFromCSV(csvReader)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	response := map[string]interface{}{
		"created_count": createdCount,
		"message":       fmt.Sprintf("Berhasil membuat %d siswa dengan password default: %s", createdCount, DefaultStudentPassword),
	}

	if len(validationErrors) > 0 {
		response["validation_errors"] = validationErrors
		response["message"] = fmt.Sprintf("Berhasil membuat %d siswa, tetapi %d baris mengalami error", createdCount, len(validationErrors))
	}

	return utils.SuccessResponse(c, response, "Upload CSV berhasil diproses")
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req ChangePasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	err := h.service.ChangePassword(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Password berhasil diubah",
	}, "Password berhasil diubah")
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	err := h.service.ForgotPassword(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Anda akan menerima email reset jika pengguna dengan email tersebut ada",
	}, "Link reset telah dikirim")
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Token reset diperlukan")
	}

	var req ResetPasswordRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		validationErrors := h.formatValidationErrors(err)
		return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
			Success: false,
			Error:   "Validasi gagal",
			Details: validationErrors,
		})
	}

	err := h.service.ResetPassword(token, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Password berhasil direset",
	}, "Password berhasil direset")
}
