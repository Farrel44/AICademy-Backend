package auth

import (
	"testing"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestAuthLogicWithHelpers(t *testing.T) {
	cleanup := testhelpers.SetupTestEnvironment()
	defer cleanup()

	t.Run("PasswordValidation_WithHelpers", func(t *testing.T) {
		testhelpers.TestPasswordHashing(t, "testpassword123")
		testhelpers.TestDefaultPasswordDetection(t, "telkom@2025", true)
		testhelpers.TestDefaultPasswordDetection(t, "custompassword", false)
	})

	t.Run("RequestValidation_WithHelpers", func(t *testing.T) {
		validLogin := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		testhelpers.AssertValidationPass(t, validLogin)

		invalidLogin := LoginRequest{
			Email:    "invalid-email",
			Password: "",
		}
		testhelpers.AssertValidationFail(t, invalidLogin, "Email")

		validChangePassword := ChangePasswordRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "newpass123456",
			ConfirmPassword: "newpass123456",
		}
		testhelpers.AssertValidationPass(t, validChangePassword)

		invalidChangePassword := ChangePasswordRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "short",
			ConfirmPassword: "short",
		}
		testhelpers.AssertValidationFail(t, invalidChangePassword, "NewPassword")
	})

	t.Run("TokenGeneration_WithHelpers", func(t *testing.T) {
		mockUser := testhelpers.NewMockUser("test@example.com", user.RoleStudent)
		token, err := testhelpers.TestTokenGeneration(t, mockUser)
		assert.NoError(t, err)

		testhelpers.TestTokenValidation(t, token, true)
		testhelpers.TestTokenValidation(t, "invalid.token.string", false)

		testhelpers.TestRefreshTokenGeneration(t)
		testhelpers.TestResetTokenGeneration(t)

		testhelpers.TestTokenUniqueness(t, func() (string, error) {
			return testhelpers.TestRefreshTokenGeneration(t), nil
		})
	})

	t.Run("CreateTestUsers_WithHelpers", func(t *testing.T) {
		testUsers, err := testhelpers.CreateAllRoleUsers()
		assert.NoError(t, err)
		assert.NotNil(t, testUsers.Admin)
		assert.NotNil(t, testUsers.Teacher)
		assert.NotNil(t, testUsers.Student)
		assert.NotNil(t, testUsers.Alumni)
		assert.NotNil(t, testUsers.Company)

		assert.Equal(t, user.RoleAdmin, testUsers.Admin.Role)
		assert.Equal(t, user.RoleTeacher, testUsers.Teacher.Role)
		assert.Equal(t, user.RoleStudent, testUsers.Student.Role)
		assert.Equal(t, user.RoleAlumni, testUsers.Alumni.Role)
		assert.Equal(t, user.RoleCompany, testUsers.Company.Role)
	})

	t.Run("RequestPayloads_WithHelpers", func(t *testing.T) {
		validLogin := testhelpers.CreateValidLoginRequest()
		assert.Equal(t, "test@example.com", validLogin["email"])
		assert.Equal(t, "validpassword123", validLogin["password"])

		invalidLogin := testhelpers.CreateInvalidLoginRequest()
		assert.Equal(t, "invalid-email", invalidLogin["email"])
		assert.Equal(t, "", invalidLogin["password"])

		validChangePassword := testhelpers.CreateValidChangePasswordRequest()
		assert.Equal(t, "oldpassword123", validChangePassword["current_password"])
		assert.Equal(t, "newpassword123", validChangePassword["new_password"])
		assert.Equal(t, "newpassword123", validChangePassword["confirm_password"])

		invalidChangePassword := testhelpers.CreateInvalidChangePasswordRequest()
		assert.Equal(t, "short", invalidChangePassword["new_password"])
		assert.Equal(t, "different", invalidChangePassword["confirm_password"])
	})
}
