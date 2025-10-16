package utils

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MockUser implements JWTUser interface for testing
type MockUser struct {
	ID    uuid.UUID
	Email string
	Role  string
}

func (m *MockUser) GetID() uuid.UUID {
	return m.ID
}

func (m *MockUser) GetEmail() string {
	return m.Email
}

func (m *MockUser) GetRole() string {
	return m.Role
}

// JWTTestSuite groups all JWT-related tests
type JWTTestSuite struct {
	suite.Suite
	originalJWTSecret string
	testUser          *MockUser
}

func (suite *JWTTestSuite) SetupTest() {
	// Store original JWT_SECRET
	suite.originalJWTSecret = os.Getenv("JWT_SECRET")

	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-testing")

	// Create test user
	suite.testUser = &MockUser{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "student",
	}
}

func (suite *JWTTestSuite) TearDownTest() {
	// Restore original JWT_SECRET
	if suite.originalJWTSecret != "" {
		os.Setenv("JWT_SECRET", suite.originalJWTSecret)
	} else {
		os.Unsetenv("JWT_SECRET")
	}
}

func (suite *JWTTestSuite) TestGenerateAccessToken_Success() {
	token, err := GenerateAccessToken(suite.testUser)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)

	// Verify token can be parsed
	claims, err := ValidateToken(token)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUser.GetID(), claims.UserID)
	assert.Equal(suite.T(), suite.testUser.GetEmail(), claims.Email)
	assert.Equal(suite.T(), suite.testUser.GetRole(), claims.Role)
	assert.Equal(suite.T(), "aicademy", claims.Issuer)
}

func (suite *JWTTestSuite) TestGenerateAccessToken_NoJWTSecret() {
	os.Unsetenv("JWT_SECRET")

	token, err := GenerateAccessToken(suite.testUser)

	assert.Error(suite.T(), err)
	assert.Empty(suite.T(), token)
	assert.Equal(suite.T(), "JWT_SECRET not set", err.Error())
}

func (suite *JWTTestSuite) TestGenerateRefreshToken_Success() {
	token, err := GenerateRefreshToken()

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
	assert.Equal(suite.T(), 64, len(token)) // 32 bytes * 2 (hex encoding)
}

func (suite *JWTTestSuite) TestGenerateTokenPair_Success() {
	tokenPair, err := GenerateTokenPair(suite.testUser)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), tokenPair)
	assert.NotEmpty(suite.T(), tokenPair.AccessToken)
	assert.NotEmpty(suite.T(), tokenPair.RefreshToken)
	assert.Equal(suite.T(), int64(15*60), tokenPair.ExpiresIn)

	// Verify access token
	claims, err := ValidateToken(tokenPair.AccessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testUser.GetID(), claims.UserID)
}

func (suite *JWTTestSuite) TestGenerateTokenPair_AccessTokenError() {
	os.Unsetenv("JWT_SECRET")

	tokenPair, err := GenerateTokenPair(suite.testUser)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), tokenPair)
}

func (suite *JWTTestSuite) TestLegacyGenerateToken() {
	token, err := GenerateToken(suite.testUser)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)

	// Should be same as GenerateAccessToken
	accessToken, err := GenerateAccessToken(suite.testUser)
	assert.NoError(suite.T(), err)

	// Both tokens should be valid and contain same claims
	claims1, err := ValidateToken(token)
	assert.NoError(suite.T(), err)

	claims2, err := ValidateToken(accessToken)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), claims1.UserID, claims2.UserID)
	assert.Equal(suite.T(), claims1.Email, claims2.Email)
	assert.Equal(suite.T(), claims1.Role, claims2.Role)
}

func (suite *JWTTestSuite) TestValidateToken_Success() {
	token, err := GenerateAccessToken(suite.testUser)
	assert.NoError(suite.T(), err)

	claims, err := ValidateToken(token)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), claims)
	assert.Equal(suite.T(), suite.testUser.GetID(), claims.UserID)
	assert.Equal(suite.T(), suite.testUser.GetEmail(), claims.Email)
	assert.Equal(suite.T(), suite.testUser.GetRole(), claims.Role)
	assert.Equal(suite.T(), "aicademy", claims.Issuer)
}

func (suite *JWTTestSuite) TestValidateToken_NoJWTSecret() {
	token, err := GenerateAccessToken(suite.testUser)
	assert.NoError(suite.T(), err)

	os.Unsetenv("JWT_SECRET")

	claims, err := ValidateToken(token)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), claims)
	assert.Equal(suite.T(), "JWT_SECRET not set", err.Error())
}

func (suite *JWTTestSuite) TestValidateToken_InvalidToken() {
	claims, err := ValidateToken("invalid.token.string")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), claims)
}

func (suite *JWTTestSuite) TestValidateToken_ExpiredToken() {
	// Create token with past expiry time
	expireTime := time.Now().Add(-1 * time.Hour)

	claims := Claims{
		UserID: suite.testUser.GetID(),
		Email:  suite.testUser.GetEmail(),
		Role:   suite.testUser.GetRole(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "aicademy",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-testing"))
	assert.NoError(suite.T(), err)

	validatedClaims, err := ValidateToken(tokenString)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

func (suite *JWTTestSuite) TestValidateToken_WrongSecret() {
	// Create token with different secret to sign the token
	claims := Claims{
		UserID: suite.testUser.GetID(),
		Email:  suite.testUser.GetEmail(),
		Role:   suite.testUser.GetRole(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aicademy",
		},
	}

	// Use a different secret to sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret-key"))
	assert.NoError(suite.T(), err)

	validatedClaims, err := ValidateToken(tokenString)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

func TestJWTTestSuite(t *testing.T) {
	suite.Run(t, new(JWTTestSuite))
}

// Individual test functions for specific scenarios
func TestTokenExpiryTime(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	user := &MockUser{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "student",
	}

	startTime := time.Now()
	token, err := GenerateAccessToken(user)
	assert.NoError(t, err)

	claims, err := ValidateToken(token)
	assert.NoError(t, err)

	// Check that expiry is approximately 15 minutes from now
	expectedExpiry := startTime.Add(15 * time.Minute)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 1 second tolerance
	assert.WithinDuration(t, expectedExpiry, actualExpiry, time.Second)
}

func TestRefreshTokenUniqueness(t *testing.T) {
	token1, err := GenerateRefreshToken()
	assert.NoError(t, err)

	token2, err := GenerateRefreshToken()
	assert.NoError(t, err)

	assert.NotEqual(t, token1, token2)
}
