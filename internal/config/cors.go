package config

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupCors() fiber.Handler {
	allowMethods := os.Getenv("CORS_ALLOW_METHODS")
	allowHeaders := os.Getenv("CORS_ALLOW_HEADERS")
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"

	rawOrigins := os.Getenv("CORS_ALLOW_ORIGINS")
	origins := strings.Split(rawOrigins, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}

	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			for _, o := range origins {
				if strings.Contains(origin, o) {
					return true
				}
			}
			return false
		},
		AllowMethods:     allowMethods,
		AllowHeaders:     allowHeaders,
		AllowCredentials: allowCredentials,
	})
}
