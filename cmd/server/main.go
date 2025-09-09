package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"aicademy-backend/internal/config"
	"aicademy-backend/internal/domain/auth"
	"aicademy-backend/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	authRepo := auth.NewAuthRepository(db)

	authService := auth.NewAuthService(authRepo)

	authHandler := auth.NewAuthHandler(authService)

	app := fiber.New(fiber.Config{
		AppName: "AICademy API v1.0",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3001",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "AICademy API is running!",
			"version": "1.0.0",
		})
	})

	api := app.Group("/api/v1")

	auth_routes := api.Group("/auth")
	auth_routes.Post("/register", authHandler.Register)
	auth_routes.Post("/login", authHandler.Login)

	protected := api.Group("/user")
	protected.Use(middleware.JWTProtected())

	protected.Get("/profile", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		userEmail := c.Locals("user_email")
		userRole := c.Locals("user_role")

		return c.JSON(fiber.Map{
			"user_id":    userID,
			"user_email": userEmail,
			"user_role":  userRole,
		})
	})

	admin := api.Group("/admin")
	admin.Use(middleware.JWTProtected())
	admin.Use(middleware.RequireRole("Admin"))
	admin.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Admin dashboard"})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
