package testhelpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func CreateJSONRequest(method, url string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func CreateJSONRequestWithAuth(method, url string, body interface{}, user *TestUser) (*http.Request, error) {
	req, err := CreateJSONRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+user.Token)
	return req, nil
}

func CreateCookieRequest(method, url string, cookieName, cookieValue string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: cookieValue,
	})

	return req, nil
}

func CreateFiberAppWithMiddleware(middlewares ...fiber.Handler) *fiber.App {
	app := fiber.New()

	for _, mw := range middlewares {
		app.Use(mw)
	}

	return app
}

func CreateAuthenticatedFiberApp() *fiber.App {
	return CreateFiberAppWithMiddleware(middleware.AuthRequired())
}

func CreateRoleBasedFiberApp(role user.UserRole) *fiber.App {
	app := CreateFiberAppWithMiddleware(middleware.AuthRequired())

	switch role {
	case user.RoleAdmin:
		app.Use(middleware.AdminRequired())
	case user.RoleTeacher:
		app.Use(middleware.TeacherOrAdminRequired())
	case user.RoleStudent:
		app.Use(middleware.StudentRequired())
	case user.RoleAlumni:
		app.Use(middleware.AlumniRequired())
	case user.RoleCompany:
		app.Use(middleware.CompanyRequired())
	}

	return app
}

func CreateMockLocalsMiddleware(userRole user.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("user_role", userRole)
		return c.Next()
	}
}

func AddTestRoute(app *fiber.App, method, path string) {
	handler := func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	}

	switch method {
	case "GET":
		app.Get(path, handler)
	case "POST":
		app.Post(path, handler)
	case "PUT":
		app.Put(path, handler)
	case "DELETE":
		app.Delete(path, handler)
	case "PATCH":
		app.Patch(path, handler)
	}
}

func TestFiberResponse(t *testing.T, app *fiber.App, req *http.Request, expectedStatus int) *http.Response {
	resp, err := app.Test(req)
	assert.NoError(t, err, "Request should not fail")
	assert.Equal(t, expectedStatus, resp.StatusCode, "Status code mismatch")

	return resp
}

func TestFiberJSONResponse(t *testing.T, app *fiber.App, req *http.Request, expectedStatus int, result interface{}) *http.Response {
	resp := TestFiberResponse(t, app, req, expectedStatus)

	if result != nil {
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err, "Reading response body should not fail")

		err = json.Unmarshal(body, result)
		assert.NoError(t, err, "Parsing JSON response should not fail")
	}

	return resp
}

type ResponseAssertion struct {
	StatusCode int
	Headers    map[string]string
	JSONPath   map[string]interface{}
}

func AssertResponse(t *testing.T, resp *http.Response, expected ResponseAssertion) {
	assert.Equal(t, expected.StatusCode, resp.StatusCode, "Status code mismatch")

	for header, expectedValue := range expected.Headers {
		actualValue := resp.Header.Get(header)
		assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", header)
	}

	if len(expected.JSONPath) > 0 {
		var jsonResp map[string]interface{}
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err, "Reading response body should not fail")

		err = json.Unmarshal(body, &jsonResp)
		assert.NoError(t, err, "Parsing JSON response should not fail")

		for path, expectedValue := range expected.JSONPath {
			actualValue, exists := jsonResp[path]
			assert.True(t, exists, "JSON path %s should exist", path)
			assert.Equal(t, expectedValue, actualValue, "JSON path %s value mismatch", path)
		}
	}
}
