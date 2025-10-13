package config

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func SetupCors() fiber.Handler {
	allowOrigins := strings.ReplaceAll(os.Getenv("CORS_ALLOW_ORIGINS"), " ", "")
	allowMethods := os.Getenv("CORS_ALLOW_METHODS")
	allowHeaders := os.Getenv("CORS_ALLOW_HEADERS")
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"

	// Tambahkan header yang umum dipakai browser & file upload
	if allowHeaders == "" {
		allowHeaders = "Origin, Content-Type, Accept, Authorization, Cookie, X-Requested-With, Content-Length, Accept-Language"
	} else {
		allowHeaders += ",Cookie,X-Requested-With,Content-Length,Accept-Language"
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins, // contoh: http://localhost:3000,http://localhost:3001,https://your-prod-domain.com
		AllowMethods:     allowMethods, // GET,POST,PUT,DELETE,OPTIONS
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    "Set-Cookie, Content-Disposition",
		AllowCredentials: allowCredentials, // true
		MaxAge:           86400,            // cache preflight 24 jam
	})
}
