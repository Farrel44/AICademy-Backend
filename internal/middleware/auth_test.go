package middleware

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

func TestAuthRequired(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	app := fiber.New()
	app.Use(AuthRequired())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	t.Run("ValidToken_BearerHeader", func(t *testing.T) {
		mockUser := &user.User{
			Email: "test@example.com",
			Role:  user.RoleStudent,
		}
		mockUser.ID = uuid.New()

		token, err := utils.GenerateAccessToken(mockUser)
		assert.NoError(t, err)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ValidToken_Cookie", func(t *testing.T) {
		mockUser := &user.User{
			Email: "test@example.com",
			Role:  user.RoleAdmin,
		}
		mockUser.ID = uuid.New()

		token, err := utils.GenerateAccessToken(mockUser)
		assert.NoError(t, err)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: token})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NoToken_Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("InvalidToken_Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("MalformedHeader_Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("ExpiredToken_Unauthorized", func(t *testing.T) {
		oldSecret := os.Getenv("JWT_SECRET")
		os.Setenv("JWT_SECRET", "different-secret")

		mockUser := &user.User{
			Email: "test@example.com",
			Role:  user.RoleStudent,
		}
		mockUser.ID = uuid.New()

		token, err := utils.GenerateAccessToken(mockUser)
		assert.NoError(t, err)

		os.Setenv("JWT_SECRET", oldSecret)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAdminRequired(t *testing.T) {
	t.Run("AdminRole_Allowed", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleAdmin)
			return c.Next()
		})
		app.Use(AdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("StudentRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleStudent)
			return c.Next()
		})
		app.Use(AdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("TeacherRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleTeacher)
			return c.Next()
		})
		app.Use(AdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestTeacherOrAdminRequired(t *testing.T) {
	t.Run("AdminRole_Allowed", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleAdmin)
			return c.Next()
		})
		app.Use(TeacherOrAdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("TeacherRole_Allowed", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleTeacher)
			return c.Next()
		})
		app.Use(TeacherOrAdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("StudentRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleStudent)
			return c.Next()
		})
		app.Use(TeacherOrAdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("AlumniRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleAlumni)
			return c.Next()
		})
		app.Use(TeacherOrAdminRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestStudentRequired(t *testing.T) {
	t.Run("StudentRole_Allowed", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleStudent)
			return c.Next()
		})
		app.Use(StudentRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AdminRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleAdmin)
			return c.Next()
		})
		app.Use(StudentRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestAlumniRequired(t *testing.T) {
	t.Run("AlumniRole_Allowed", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleAlumni)
			return c.Next()
		})
		app.Use(AlumniRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("CompanyRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleCompany)
			return c.Next()
		})
		app.Use(AlumniRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestCompanyRequired(t *testing.T) {
	t.Run("CompanyRole_Allowed", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleCompany)
			return c.Next()
		})
		app.Use(CompanyRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("StudentRole_Forbidden", func(t *testing.T) {
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_role", user.RoleStudent)
			return c.Next()
		})
		app.Use(CompanyRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"success": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestAuthRequiredWithLocals(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	t.Run("ValidToken_SetsLocals", func(t *testing.T) {
		app := fiber.New()
		app.Use(AuthRequired())
		app.Get("/test", func(c *fiber.Ctx) error {
			userClaims := c.Locals("user").(*UserClaims)
			userID := c.Locals("user_id").(string)
			userEmail := c.Locals("user_email").(string)
			userRole := c.Locals("user_role").(user.UserRole)

			return c.JSON(fiber.Map{
				"user_id":    userID,
				"user_email": userEmail,
				"user_role":  userRole,
				"claims_id":  userClaims.UserID.String(),
			})
		})

		mockUser := &user.User{
			Email: "test@example.com",
			Role:  user.RoleTeacher,
		}
		mockUser.ID = uuid.New()

		token, err := utils.GenerateAccessToken(mockUser)
		assert.NoError(t, err)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
