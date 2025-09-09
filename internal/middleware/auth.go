package middleware

import (
	"aicademy-backend/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   "Missing authorization header",
			})
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid authorization format",
			})
		}

		claims, err := utils.ValidateToken(tokenParts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("user_role").(string)

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"success": false,
			"error":   "Insufficient permissions",
		})
	}
}
