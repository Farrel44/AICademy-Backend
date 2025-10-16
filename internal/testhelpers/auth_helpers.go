package testhelpers

import (
	"net/http"
	"os"
	"testing"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type MockUser struct {
	ID    uuid.UUID
	Email string
	Role  user.UserRole
}

func (m *MockUser) GetID() uuid.UUID {
	return m.ID
}

func (m *MockUser) GetEmail() string {
	return m.Email
}

func (m *MockUser) GetRole() string {
	return string(m.Role)
}

type TestUser struct {
	ID    uuid.UUID
	Email string
	Role  user.UserRole
	Token string
}

func NewMockUser(email string, role user.UserRole) *MockUser {
	return &MockUser{
		ID:    uuid.New(),
		Email: email,
		Role:  role,
	}
}

func CreateTestUser(email string, role user.UserRole) (*TestUser, error) {
	mockUser := NewMockUser(email, role)

	token, err := utils.GenerateAccessToken(mockUser)
	if err != nil {
		return nil, err
	}

	return &TestUser{
		ID:    mockUser.ID,
		Email: mockUser.Email,
		Role:  mockUser.Role,
		Token: token,
	}, nil
}

func SetupTestEnvironment() func() {
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")

	return func() {
		os.Unsetenv("JWT_SECRET")
	}
}

func CreateAuthenticatedRequest(method, url string, user *TestUser) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+user.Token)
	return req, nil
}

func CreateFiberAppWithAuth() *fiber.App {
	app := fiber.New()
	return app
}

func AssertSuccessResponse(t *testing.T, resp *http.Response) {
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func AssertUnauthorizedResponse(t *testing.T, resp *http.Response) {
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func AssertForbiddenResponse(t *testing.T, resp *http.Response) {
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

type TestUsers struct {
	Admin   *TestUser
	Teacher *TestUser
	Student *TestUser
	Alumni  *TestUser
	Company *TestUser
}

func CreateAllRoleUsers() (*TestUsers, error) {
	admin, err := CreateTestUser("admin@test.com", user.RoleAdmin)
	if err != nil {
		return nil, err
	}

	teacher, err := CreateTestUser("teacher@test.com", user.RoleTeacher)
	if err != nil {
		return nil, err
	}

	student, err := CreateTestUser("student@test.com", user.RoleStudent)
	if err != nil {
		return nil, err
	}

	alumni, err := CreateTestUser("alumni@test.com", user.RoleAlumni)
	if err != nil {
		return nil, err
	}

	company, err := CreateTestUser("company@test.com", user.RoleCompany)
	if err != nil {
		return nil, err
	}

	return &TestUsers{
		Admin:   admin,
		Teacher: teacher,
		Student: student,
		Alumni:  alumni,
		Company: company,
	}, nil
}

func CreateValidLoginRequest() map[string]interface{} {
	return map[string]interface{}{
		"email":    "test@example.com",
		"password": "validpassword123",
	}
}

func CreateInvalidLoginRequest() map[string]interface{} {
	return map[string]interface{}{
		"email":    "invalid-email",
		"password": "",
	}
}

func CreateValidChangePasswordRequest() map[string]interface{} {
	return map[string]interface{}{
		"current_password": "oldpassword123",
		"new_password":     "newpassword123",
		"confirm_password": "newpassword123",
	}
}

func CreateInvalidChangePasswordRequest() map[string]interface{} {
	return map[string]interface{}{
		"current_password": "oldpassword123",
		"new_password":     "short",
		"confirm_password": "different",
	}
}

func CreateValidResetPasswordRequest() map[string]interface{} {
	return map[string]interface{}{
		"password":        "newpassword123",
		"passwordConfirm": "newpassword123",
	}
}

func CreateInvalidResetPasswordRequest() map[string]interface{} {
	return map[string]interface{}{
		"password":        "short",
		"passwordConfirm": "different",
	}
}

func CreateValidRefreshTokenRequest() map[string]interface{} {
	token, _ := utils.GenerateRefreshToken()
	return map[string]interface{}{
		"refresh_token": token,
	}
}

func CreateInvalidRefreshTokenRequest() map[string]interface{} {
	return map[string]interface{}{
		"refresh_token": "",
	}
}
