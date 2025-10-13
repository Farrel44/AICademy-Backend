package config

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupCors() fiber.Handler {
	allowOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	allowMethods := os.Getenv("CORS_ALLOW_METHODS")
	allowHeaders := os.Getenv("CORS_ALLOW_HEADERS")
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"

	allowOrigins = strings.ReplaceAll(allowOrigins, " ", "")

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: allowCredentials,
	})
}
