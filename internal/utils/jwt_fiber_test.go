package utils

import (
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestGetClaimsFromHeader_Success(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	user := &MockUser{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "student",
	}

	token, err := GenerateAccessToken(user)
	assert.NoError(t, err)

	// Create Fiber app and context
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Request().Header.Set("Authorization", "Bearer "+token)
	defer app.ReleaseCtx(ctx)

	// Test
	claims, err := GetClaimsFromHeader(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.GetID(), claims.UserID)
	assert.Equal(t, user.GetEmail(), claims.Email)
	assert.Equal(t, user.GetRole(), claims.Role)
}

func TestGetClaimsFromHeader_MissingHeader(t *testing.T) {
	// Create Fiber app and context
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	// Test
	claims, err := GetClaimsFromHeader(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, "authorization header is required", err.Error())
}

func TestGetClaimsFromHeader_InvalidFormat(t *testing.T) {
	// Setup JWT secret
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Create Fiber app and context
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(ctx)

	testCases := []struct {
		name          string
		header        string
		expectedError string
	}{
		{"No Bearer prefix", "token123", "authorization header format must be Bearer {token}"},
		{"Wrong prefix", "Basic token123", "authorization header format must be Bearer {token}"},
		{"Missing token", "Bearer ", "authorization header format must be Bearer {token}"},
		{"Missing space", "Bearertoken123", "authorization header format must be Bearer {token}"},
		{"Multiple spaces", "Bearer  token123", "token is malformed: token contains an invalid number of segments"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx.Request().Header.Set("Authorization", tc.header)

			claims, err := GetClaimsFromHeader(ctx)

			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Equal(t, tc.expectedError, err.Error())
		})
	}
}

func TestGetClaimsFromHeader_InvalidToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	// Create Fiber app and context
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Request().Header.Set("Authorization", "Bearer invalid.token.here")
	defer app.ReleaseCtx(ctx)

	// Test
	claims, err := GetClaimsFromHeader(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestGetClaimsFromHeader_CaseInsensitiveBearer(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	user := &MockUser{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "student",
	}

	token, err := GenerateAccessToken(user)
	assert.NoError(t, err)

	// Test different cases
	testCases := []string{
		"bearer " + token,
		"BEARER " + token,
		"Bearer " + token,
		"BeArEr " + token,
	}

	for _, authHeader := range testCases {
		t.Run("Header: "+authHeader[:6], func(t *testing.T) {
			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			ctx.Request().Header.Set("Authorization", authHeader)
			defer app.ReleaseCtx(ctx)

			claims, err := GetClaimsFromHeader(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, user.GetID(), claims.UserID)
		})
	}
}
