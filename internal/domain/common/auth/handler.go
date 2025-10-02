package auth

import (
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type CommonAuthHandler struct {
	service *CommonAuthService
}

func NewCommonAuthHandler(service *CommonAuthService) *CommonAuthHandler {
	return &CommonAuthHandler{service: service}
}

func (h *CommonAuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.Login(req)
	if err != nil {
		switch err.Error() {
		case "invalid email or password":
			return utils.ErrorResponse(c, 401, "Invalid email or password")
		case "failed to generate token":
			return utils.ErrorResponse(c, 500, "Internal server error")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	// Set cookies dengan access token
	h.setAuthCookies(c, result.AccessToken, result.User.Role)

	return utils.SuccessResponse(c, result, "Login successful")
}

func (h *CommonAuthHandler) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	profile, err := h.service.GetMe(user.UserID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found")
		case "student profile not found":
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Student profile not found")
		case "alumni profile not found":
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Alumni profile not found")
		case "teacher profile not found":
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Teacher profile not found")
		case "company profile not found":
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Company profile not found")
		case "invalid user role":
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid user role")
		default:
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, profile, "Profil berhasil diambil")
}

func (h *CommonAuthHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Try to get refresh token from request body
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		// Logout dengan refresh token (hapus dari database)
		if err := h.service.Logout(req.RefreshToken); err != nil {
			// Log error tapi tetap clear cookies
		}
	}

	// Clear cookies
	h.clearAuthCookies(c)

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Logout successful",
	}, "Logout successful")
}

func (h *CommonAuthHandler) ChangePassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.ErrorResponse(c, 401, "Unauthorized")
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ChangePassword(user.UserID, req)
	if err != nil {
		switch err.Error() {
		case "password confirmation does not match":
			return utils.ErrorResponse(c, 422, "Password confirmation does not match")
		case "user not found":
			return utils.ErrorResponse(c, 404, "User not found")
		case "current password is incorrect":
			return utils.ErrorResponse(c, 401, "Current password is incorrect")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Password berhasil diubah",
	}, "Password berhasil diubah")
}

func (h *CommonAuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ForgotPassword(req)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to send reset email")
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "If the email exists, a reset link has been sent",
	}, "Reset email sent")
}

func (h *CommonAuthHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return utils.ErrorResponse(c, 400, "Reset token is required")
	}

	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ResetPassword(token, req)
	if err != nil {
		switch err.Error() {
		case "password confirmation does not match":
			return utils.ErrorResponse(c, 422, "Password confirmation does not match")
		case "invalid or expired reset token":
			return utils.ErrorResponse(c, 400, "Invalid or expired reset token")
		case "reset token has expired":
			return utils.ErrorResponse(c, 400, "Reset token has expired")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, MessageResponse{
		Message: "Password berhasil direset",
	}, "Password reset successful")
} // Helper methods
func (h *CommonAuthHandler) setAuthCookies(c *fiber.Ctx, token, role string) {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    role,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})
}

func (h *CommonAuthHandler) clearAuthCookies(c *fiber.Ctx) {
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
		HTTPOnly: false,
		Secure:   false,
		SameSite: "Lax",
	})
}

func (h *CommonAuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Format data tidak valid")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.RefreshToken(req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired refresh token":
			return utils.ErrorResponse(c, 401, "Invalid or expired refresh token")
		case "user not found":
			return utils.ErrorResponse(c, 404, "User not found")
		case "failed to generate access token":
			return utils.ErrorResponse(c, 500, "Internal server error")
		default:
			return utils.ErrorResponse(c, 500, "Internal server error")
		}
	}

	return utils.SuccessResponse(c, result, "Token berhasil direfresh")
}
