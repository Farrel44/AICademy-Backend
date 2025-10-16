package testhelpers

import (
	"testing"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
	cleanup   func()
	testUsers *TestUsers
}

func (suite *AuthTestSuite) SetupSuite() {
	suite.cleanup = SetupTestEnvironment()

	testUsers, err := CreateAllRoleUsers()
	suite.Require().NoError(err)
	suite.testUsers = testUsers
}

func (suite *AuthTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *AuthTestSuite) TestUserCreation() {
	student, err := CreateTestUser("student@test.com", user.RoleStudent)
	suite.NoError(err)
	suite.Equal(user.RoleStudent, student.Role)
	suite.NotEmpty(student.Token)
	suite.NotEmpty(student.ID)
}

func (suite *AuthTestSuite) TestTokenValidation() {
	TestTokenValidation(suite.T(), suite.testUsers.Admin.Token, true)
	TestTokenValidation(suite.T(), "invalid.token.string", false)
}

func (suite *AuthTestSuite) TestValidationHelpers() {
	validLogin := CreateValidLoginRequest()
	AssertValidationPass(suite.T(), validLogin)

	invalidLogin := CreateInvalidLoginRequest()
	AssertValidationFail(suite.T(), invalidLogin, "Email")
}

func (suite *AuthTestSuite) TestPasswordHelpers() {
	TestPasswordHashing(suite.T(), "testpassword123")
	TestDefaultPasswordDetection(suite.T(), "telkom@2025", true)
	TestDefaultPasswordDetection(suite.T(), "custompassword", false)
}

func (suite *AuthTestSuite) TestTokenGenerationHelpers() {
	mockUser := NewMockUser("test@example.com", user.RoleStudent)
	TestTokenGeneration(suite.T(), mockUser)

	TestRefreshTokenGeneration(suite.T())
	TestResetTokenGeneration(suite.T())
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func TestIndividualHelpers(t *testing.T) {
	cleanup := SetupTestEnvironment()
	defer cleanup()

	t.Run("MockUser Creation", func(t *testing.T) {
		mockUser := NewMockUser("test@example.com", user.RoleAdmin)
		assert.NotEmpty(t, mockUser.ID)
		assert.Equal(t, "test@example.com", mockUser.Email)
		assert.Equal(t, user.RoleAdmin, mockUser.Role)
	})

	t.Run("HTTP Request Creation", func(t *testing.T) {
		testUser, err := CreateTestUser("test@example.com", user.RoleStudent)
		assert.NoError(t, err)

		req, err := CreateJSONRequestWithAuth("GET", "/test", nil, testUser)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer "+testUser.Token, req.Header.Get("Authorization"))
	})

	t.Run("Fiber App Creation", func(t *testing.T) {
		app := CreateAuthenticatedFiberApp()
		assert.NotNil(t, app)

		roleApp := CreateRoleBasedFiberApp(user.RoleAdmin)
		assert.NotNil(t, roleApp)
	})

	t.Run("Request Payloads", func(t *testing.T) {
		validLogin := CreateValidLoginRequest()
		assert.NotEmpty(t, validLogin["email"])
		assert.NotEmpty(t, validLogin["password"])

		invalidLogin := CreateInvalidLoginRequest()
		assert.Equal(t, "invalid-email", invalidLogin["email"])
		assert.Equal(t, "", invalidLogin["password"])
	})
}
