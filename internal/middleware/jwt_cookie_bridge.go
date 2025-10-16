package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTCookieBridge injects Authorization header from cookies if missing.
func JWTCookieBridge() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if strings.TrimSpace(auth) == "" {
			// Try cookies set by login
			token := c.Cookies("access_token")
			if token == "" {
				token = c.Cookies("token")
			}
			if token != "" {
				c.Request().Header.Set("Authorization", "Bearer "+token)
			}
		}
		return c.Next()
	}
}
