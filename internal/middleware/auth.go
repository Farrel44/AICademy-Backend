// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserClaims struct {
	UserID uuid.UUID     `json:"user_id"`
	Email  string        `json:"email"`
	Role   user.UserRole `json:"role"`
}

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token = tokenParts[1]
			}
		}

		if token == "" {
			token = c.Cookies("token")
		}

		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization token required",
			})
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		userClaims := &UserClaims{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   user.UserRole(claims.Role),
		}

		c.Locals("user", userClaims)
		c.Locals("user_id", claims.UserID.String())
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", user.UserRole(claims.Role))

		return c.Next()
	}
}

func AdminRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("user_role")
		if role != user.RoleAdmin {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Admin access required",
			})
		}
		return c.Next()
	}
}

func TeacherOrAdminRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("user_role")
		if role != user.RoleAdmin && role != user.RoleTeacher {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Teacher or Admin access required",
			})
		}
		return c.Next()
	}
}

func StudentRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("user_role")
		if role != user.RoleStudent {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Student access required",
			})
		}
		return c.Next()
	}
}

func AlumniRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("user_role")
		if role != user.RoleAlumni {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Alumni access required",
			})
		}
		return c.Next()
	}
}

func CompanyRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("user_role")
		if role != user.RoleCompany {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Company access required",
			})
		}
		return c.Next()
	}
}
