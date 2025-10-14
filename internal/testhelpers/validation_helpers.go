package testhelpers

import (
	"testing"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T, req interface{}, shouldBeValid bool, expectedErrorField string) {
	err := utils.ValidateStruct(req)

	if shouldBeValid {
		assert.NoError(t, err, "Expected validation to pass")
	} else {
		assert.Error(t, err, "Expected validation to fail")
		if expectedErrorField != "" {
			assert.Contains(t, err.Error(), expectedErrorField, "Expected error to contain field name")
		}
	}
}

func AssertValidationPass(t *testing.T, req interface{}) {
	TestValidation(t, req, true, "")
}

func AssertValidationFail(t *testing.T, req interface{}, expectedField string) {
	TestValidation(t, req, false, expectedField)
}

func TestPasswordHashing(t *testing.T, password string) (string, bool) {
	hashedPassword, err := utils.HashPassword(password)
	assert.NoError(t, err, "Password hashing should not fail")
	assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")

	isValid := utils.CheckPassword(password, hashedPassword)
	assert.True(t, isValid, "Password verification should pass")

	return hashedPassword, isValid
}

func TestDefaultPasswordDetection(t *testing.T, password string, shouldBeDefault bool) {
	isDefault := utils.IsDefaultPassword(password)
	assert.Equal(t, shouldBeDefault, isDefault, "Default password detection mismatch")
}

func TestTokenGeneration(t *testing.T, user *MockUser) (string, error) {
	token, err := utils.GenerateAccessToken(user)
	assert.NoError(t, err, "Token generation should not fail")
	assert.NotEmpty(t, token, "Generated token should not be empty")

	return token, err
}

func TestTokenValidation(t *testing.T, token string, shouldBeValid bool) {
	claims, err := utils.ValidateToken(token)

	if shouldBeValid {
		assert.NoError(t, err, "Token validation should pass")
		assert.NotNil(t, claims, "Claims should not be nil")
		assert.NotEmpty(t, claims.Email, "Email claim should not be empty")
		assert.NotEmpty(t, claims.Role, "Role claim should not be empty")
	} else {
		assert.Error(t, err, "Token validation should fail")
		assert.Nil(t, claims, "Claims should be nil for invalid token")
	}
}

func TestRefreshTokenGeneration(t *testing.T) string {
	token, err := utils.GenerateRefreshToken()
	assert.NoError(t, err, "Refresh token generation should not fail")
	assert.NotEmpty(t, token, "Refresh token should not be empty")
	assert.Equal(t, 64, len(token), "Refresh token should be 64 characters")

	return token
}

func TestResetTokenGeneration(t *testing.T) string {
	token, err := utils.GenerateResetToken()
	assert.NoError(t, err, "Reset token generation should not fail")
	assert.NotEmpty(t, token, "Reset token should not be empty")
	assert.Equal(t, 64, len(token), "Reset token should be 64 characters")

	return token
}

func TestTokenUniqueness(t *testing.T, tokenGenerator func() (string, error)) {
	token1, err1 := tokenGenerator()
	token2, err2 := tokenGenerator()

	assert.NoError(t, err1, "First token generation should not fail")
	assert.NoError(t, err2, "Second token generation should not fail")
	assert.NotEqual(t, token1, token2, "Tokens should be unique")
}
