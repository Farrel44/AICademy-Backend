package auth

import (
	"os"
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
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.Login(req)
	if err != nil {
		switch err.Error() {
		case "invalid email or password":
			return utils.SendError(c, 401, "Invalid email or password")
		case "failed to generate token":
			return utils.SendError(c, 500, "Internal server error")
		default:
			return utils.SendError(c, 500, "Internal server error")
		}
	}

	// Set cookies dengan access token
	h.setAuthCookies(c, result.AccessToken, result.User.Role, result.RefreshToken)

	return utils.SendSuccess(c, "Login successful", result)

}

func (h *CommonAuthHandler) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.SendError(c, fiber.StatusUnauthorized, "Unauthorized")
	}

	profile, err := h.service.GetMe(user.UserID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return utils.SendError(c, fiber.StatusNotFound, "User not found")
		case "student profile not found":
			return utils.SendError(c, fiber.StatusNotFound, "Student profile not found")
		case "alumni profile not found":
			return utils.SendError(c, fiber.StatusNotFound, "Alumni profile not found")
		case "teacher profile not found":
			return utils.SendError(c, fiber.StatusNotFound, "Teacher profile not found")
		case "company profile not found":
			return utils.SendError(c, fiber.StatusNotFound, "Company profile not found")
		case "invalid user role":
			return utils.SendError(c, fiber.StatusBadRequest, "Invalid user role")
		default:
			return utils.SendError(c, fiber.StatusInternalServerError, "Internal server error")
		}
	}

	return utils.SendSuccess(c, "Profile retrieved successfully", profile)
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

	return utils.SendSuccess(c, "Logout successful", MessageResponse{
		Message: "Logout successful",
	})
}

func (h *CommonAuthHandler) ChangePassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*middleware.UserClaims)
	if user == nil {
		return utils.SendError(c, 401, "Unauthorized")
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ChangePassword(user.UserID, req)
	if err != nil {
		switch err.Error() {
		case "password confirmation does not match":
			return utils.SendError(c, 422, "Password confirmation does not match")
		case "user not found":
			return utils.SendError(c, 404, "User not found")
		case "current password is incorrect":
			return utils.SendError(c, 401, "Current password is incorrect")
		default:
			return utils.SendError(c, 500, "Internal server error")
		}
	}

	return utils.SendSuccess(c, "Password changed successfully", MessageResponse{
		Message: "Password changed successfully",
	})
}

func (h *CommonAuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ForgotPassword(req)
	if err != nil {
		return utils.SendError(c, 500, "Failed to send reset email")
	}

	return utils.SendSuccess(c, "Reset email sent", MessageResponse{
		Message: "If the email exists, a reset link has been sent",
	})
}

func (h *CommonAuthHandler) ResetPassword(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return utils.SendError(c, 400, "Reset token is required")
	}

	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, 400, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	err := h.service.ResetPassword(token, req)
	if err != nil {
		switch err.Error() {
		case "password confirmation does not match":
			return utils.SendError(c, 422, "Password confirmation does not match")
		case "invalid or expired reset token":
			return utils.SendError(c, 400, "Invalid or expired reset token")
		case "reset token has expired":
			return utils.SendError(c, 400, "Reset token has expired")
		default:
			return utils.SendError(c, 500, "Internal server error")
		}
	}

	return utils.SendSuccess(c, "Password reset successful", MessageResponse{
		Message: "Password reset successfully",
	})
}

// Helper function to check if we're in production
func isProduction() bool {
	env := os.Getenv("APP_ENV")
	return env == "production"
}

// Helper methods - UPDATED
func (h *CommonAuthHandler) setAuthCookies(c *fiber.Ctx, token, role, refresh string) {
	secure := isProduction()
	sameSite := "Lax"
	if secure {
		sameSite = "None"
	}

	// Access token (readable by FE jika perlu; bisa dipertimbangkan HttpOnly=false/true sesuai desain)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		Secure:   secure,
		HTTPOnly: false,
		SameSite: sameSite,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	// Legacy token
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		Secure:   secure,
		HTTPOnly: false,
		SameSite: sameSite,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	// Role (opsional)
	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    role,
		Path:     "/",
		Secure:   secure,
		HTTPOnly: false,
		SameSite: sameSite,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Refresh token (HttpOnly)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refresh,
		Path:     "/",
		Secure:   secure,
		HTTPOnly: true,
		SameSite: sameSite,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *CommonAuthHandler) clearAuthCookies(c *fiber.Ctx) {
	cookieNames := []string{"access_token", "token", "role", "refresh_token"}

	isSecure := isProduction()
	sameSite := "Lax"
	if isSecure {
		sameSite = "None"
	}

	for _, name := range cookieNames {
		httpOnly := name == "refresh_token"
		c.Cookie(&fiber.Cookie{
			Name:     name,
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: httpOnly,
			Secure:   isSecure,
			SameSite: sameSite,
			Path:     "/",
			Domain:   "",
		})
	}
}

func (h *CommonAuthHandler) RefreshToken(c *fiber.Ctx) error {
	var refreshToken string

	var reqBody struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&reqBody); err == nil && reqBody.RefreshToken != "" {
		refreshToken = reqBody.RefreshToken
	} else {
		refreshToken = c.Cookies("refresh_token")
	}

	if refreshToken == "" {
		return utils.SendError(c, 400, "Refresh token is required")
	}

	req := RefreshTokenRequest{RefreshToken: refreshToken}
	if err := utils.ValidateStruct(req); err != nil {
		return utils.ValidationErrorResponse(c, err)
	}

	result, err := h.service.RefreshToken(req)
	if err != nil {
		switch err.Error() {
		case "invalid or expired refresh token":
			return utils.SendError(c, 401, "Invalid or expired refresh token")
		case "user not found":
			return utils.SendError(c, 404, "User not found")
		case "failed to generate access token":
			return utils.SendError(c, 500, "Internal server error")
		default:
			return utils.SendError(c, 500, "Internal server error")
		}
	}

	// Determine security settings based on environment
	isSecure := isProduction()
	sameSite := "Lax"
	if isSecure {
		sameSite = "None"
	}

	// Update access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: false,
		Secure:   isSecure,
		SameSite: sameSite,
		Path:     "/",
		Domain:   "",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    result.AccessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: false,
		Secure:   isSecure,
		SameSite: sameSite,
		Path:     "/",
		Domain:   "",
	})

	return utils.SendSuccess(c, "Token refreshed successfully", result)
}
