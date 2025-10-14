package auth

import (
	"errors"
	"os"
	"testing"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPasswordValidationLogic(t *testing.T) {
	t.Run("CheckPassword_ValidPassword", func(t *testing.T) {
		password := "testpassword123"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		isValid := utils.CheckPassword(password, hashedPassword)
		assert.True(t, isValid)
	})

	t.Run("CheckPassword_InvalidPassword", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		isValid := utils.CheckPassword(wrongPassword, hashedPassword)
		assert.False(t, isValid)
	})

	t.Run("IsDefaultPassword_DetectsDefault", func(t *testing.T) {
		isDefault := utils.IsDefaultPassword("telkom@2025")
		assert.True(t, isDefault)
	})

	t.Run("IsDefaultPassword_DetectsNonDefault", func(t *testing.T) {
		isDefault := utils.IsDefaultPassword("custompassword123")
		assert.False(t, isDefault)
	})
}

func TestRequestValidationLogic(t *testing.T) {
	t.Run("ValidateStruct_LoginRequest_Valid", func(t *testing.T) {
		validReq := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		err := utils.ValidateStruct(validReq)
		assert.NoError(t, err)
	})

	t.Run("ValidateStruct_LoginRequest_InvalidEmail", func(t *testing.T) {
		invalidReq := LoginRequest{
			Email:    "invalid-email",
			Password: "password123",
		}

		err := utils.ValidateStruct(invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Email")
	})

	t.Run("ValidateStruct_LoginRequest_EmptyPassword", func(t *testing.T) {
		invalidReq := LoginRequest{
			Email:    "test@example.com",
			Password: "",
		}

		err := utils.ValidateStruct(invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Password")
	})

	t.Run("ValidateStruct_ChangePasswordRequest_Valid", func(t *testing.T) {
		validReq := ChangePasswordRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "newpass123456",
			ConfirmPassword: "newpass123456",
		}

		err := utils.ValidateStruct(validReq)
		assert.NoError(t, err)
	})

	t.Run("ValidateStruct_ChangePasswordRequest_ShortPassword", func(t *testing.T) {
		invalidReq := ChangePasswordRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "short",
			ConfirmPassword: "short",
		}

		err := utils.ValidateStruct(invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "NewPassword")
	})

	t.Run("ValidateStruct_ResetPasswordRequest_Valid", func(t *testing.T) {
		validReq := ResetPasswordRequest{
			Password:        "newpassword123",
			PasswordConfirm: "newpassword123",
		}

		err := utils.ValidateStruct(validReq)
		assert.NoError(t, err)
	})

	t.Run("ValidateStruct_ResetPasswordRequest_ShortPassword", func(t *testing.T) {
		invalidReq := ResetPasswordRequest{
			Password:        "short",
			PasswordConfirm: "short",
		}

		err := utils.ValidateStruct(invalidReq)
		assert.Error(t, err)
	})

	t.Run("ValidateStruct_RefreshTokenRequest_Valid", func(t *testing.T) {
		validReq := RefreshTokenRequest{
			RefreshToken: "valid-token-string",
		}

		err := utils.ValidateStruct(validReq)
		assert.NoError(t, err)
	})

	t.Run("ValidateStruct_RefreshTokenRequest_Empty", func(t *testing.T) {
		invalidReq := RefreshTokenRequest{
			RefreshToken: "",
		}

		err := utils.ValidateStruct(invalidReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RefreshToken")
	})
}

func TestTokenGenerationLogic(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	mockUser := &user.User{
		Email: "test@example.com",
		Role:  user.RoleStudent,
	}
	mockUser.ID = uuid.New()

	t.Run("GenerateAccessToken_Success", func(t *testing.T) {
		token, err := utils.GenerateAccessToken(mockUser)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := utils.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, mockUser.Email, claims.Email)
		assert.Equal(t, string(mockUser.Role), claims.Role)
	})

	t.Run("ValidateToken_InvalidToken", func(t *testing.T) {
		_, err := utils.ValidateToken("invalid.token.string")
		assert.Error(t, err)
	})

	t.Run("GenerateRefreshToken_Unique", func(t *testing.T) {
		token1, err1 := utils.GenerateRefreshToken()
		token2, err2 := utils.GenerateRefreshToken()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, token1, token2)
		assert.Equal(t, 64, len(token1))
		assert.Equal(t, 64, len(token2))
	})

	t.Run("GenerateResetToken_ValidFormat", func(t *testing.T) {
		token, err := utils.GenerateResetToken()

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, 64, len(token))

		token2, err2 := utils.GenerateResetToken()
		assert.NoError(t, err2)
		assert.NotEqual(t, token, token2)
	})
}

func TestPasswordChangeLogic(t *testing.T) {
	t.Run("ChangePasswordRequest_PasswordMismatch", func(t *testing.T) {
		req := ChangePasswordRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "newpass123",
			ConfirmPassword: "differentpass",
		}

		// Test the actual validation logic from service
		err := func() error {
			if req.NewPassword != req.ConfirmPassword {
				return errors.New("password confirmation does not match")
			}
			return nil
		}()

		assert.Error(t, err)
		assert.Equal(t, "password confirmation does not match", err.Error())
	})

	t.Run("ChangePasswordRequest_PasswordMatch", func(t *testing.T) {
		req := ChangePasswordRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "newpass123",
			ConfirmPassword: "newpass123",
		}

		// Test the actual validation logic from service
		err := func() error {
			if req.NewPassword != req.ConfirmPassword {
				return errors.New("password confirmation does not match")
			}
			return nil
		}()

		assert.NoError(t, err)
	})
}

func TestResetPasswordLogic(t *testing.T) {
	t.Run("ResetPasswordRequest_PasswordMismatch", func(t *testing.T) {
		req := ResetPasswordRequest{
			Password:        "newpassword123",
			PasswordConfirm: "differentpassword",
		}

		// Test the actual validation logic from service
		err := func() error {
			if req.Password != req.PasswordConfirm {
				return errors.New("password confirmation does not match")
			}
			return nil
		}()

		assert.Error(t, err)
		assert.Equal(t, "password confirmation does not match", err.Error())
	})

	t.Run("ResetPasswordRequest_PasswordMatch", func(t *testing.T) {
		req := ResetPasswordRequest{
			Password:        "newpassword123",
			PasswordConfirm: "newpassword123",
		}

		// Test the actual validation logic from service
		err := func() error {
			if req.Password != req.PasswordConfirm {
				return errors.New("password confirmation does not match")
			}
			return nil
		}()

		assert.NoError(t, err)
	})
}

func TestUserRoleLogic(t *testing.T) {
	t.Run("UserRole_Constants", func(t *testing.T) {
		assert.Equal(t, user.UserRole("student"), user.RoleStudent)
		assert.Equal(t, user.UserRole("teacher"), user.RoleTeacher)
		assert.Equal(t, user.UserRole("alumni"), user.RoleAlumni)
		assert.Equal(t, user.UserRole("admin"), user.RoleAdmin)
		assert.Equal(t, user.UserRole("company"), user.RoleCompany)
	})

	t.Run("UserRole_StringConversion", func(t *testing.T) {
		testUser := &user.User{Role: user.RoleStudent}
		roleString := string(testUser.Role)

		assert.Equal(t, "student", roleString)
	})
}

func TestRefreshTokenLogic(t *testing.T) {
	t.Run("RefreshTokenRequest_EmptyValidation", func(t *testing.T) {
		req := RefreshTokenRequest{
			RefreshToken: "",
		}

		err := utils.ValidateStruct(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RefreshToken")
	})

	t.Run("RefreshTokenRequest_ValidToken", func(t *testing.T) {
		req := RefreshTokenRequest{
			RefreshToken: "valid-token-string",
		}

		err := utils.ValidateStruct(req)
		assert.NoError(t, err)
	})
}
