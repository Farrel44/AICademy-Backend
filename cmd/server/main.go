package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"aicademy-backend/internal/config"
	"aicademy-backend/internal/domain/auth"
	"aicademy-backend/internal/domain/questionnaire"
	"aicademy-backend/internal/middleware"
	"aicademy-backend/internal/services/ai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	log.Println("Starting AICademy Backend Server...")

	log.Println("Connecting to database...")
	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")

	log.Println("Seeding database...")
	if err := config.SeedData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	} else {
		log.Println("Database seeded successfully")
	}

	log.Println("Initializing AI service...")
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	var aiService ai.AIService

	if geminiAPIKey != "" {
		aiService, err = ai.NewGeminiService(geminiAPIKey)
		if err != nil {
			log.Printf("Failed to initialize Gemini AI service: %v", err)
			log.Println("Falling back to NoAI service")
			aiService = ai.NewNoAIService()
		} else {
			log.Println("Gemini AI service initialized successfully")
		}
	} else {
		log.Println("GEMINI_API_KEY not found in environment")
		log.Println("Using NoAI service - AI features will be disabled")
		aiService = ai.NewNoAIService()
	}

	log.Println("Initializing services...")
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	questionnaireRepo := questionnaire.NewQuestionnaireRepository(db)
	questionnaireService := questionnaire.NewQuestionnaireService(questionnaireRepo, aiService)
	questionnaireHandler := questionnaire.NewQuestionnaireHandler(questionnaireService)

	app := fiber.New(fiber.Config{
		AppName:      "AICademy API v1.0",
		ServerHeader: "Fiber",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			log.Printf("Error: %v", err)

			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3001,http://localhost:3000,http://127.0.0.1:3000",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "AICademy API is running!",
			"version": "1.0.0",
			"status":  "OK",
			"ai_service": func() string {
				if geminiAPIKey != "" {
					return "Gemini AI"
				}
				return "Disabled"
			}(),
		})
	})

	api := app.Group("/api/v1")

	authRoutes := api.Group("/auth")
	authRoutes.Post("/register/alumni", authHandler.RegisterAlumni)
	authRoutes.Post("/login", authHandler.Login)
	authRoutes.Post("/forgot-password", authHandler.ForgotPassword)
	authRoutes.Post("/reset-password/:token", authHandler.ResetPassword)

	protectedAuth := authRoutes.Group("/", middleware.AuthRequired())
	protectedAuth.Post("/change-password", authHandler.ChangePassword)
	protectedAuth.Post("/logout", authHandler.Logout)

	adminRoutes := api.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())
	adminRoutes.Post("/create-teacher", authHandler.CreateTeacher)
	adminRoutes.Post("/create-student", authHandler.CreateStudent)
	adminRoutes.Post("/create-company", authHandler.CreateCompany)
	adminRoutes.Post("/upload-students-csv", authHandler.UploadStudentsCSV)

	questionnaireRoutes := api.Group("/questionnaire", middleware.AuthRequired())
	questionnaireRoutes.Get("/active", questionnaireHandler.GetActiveQuestionnaire)
	questionnaireRoutes.Post("/submit", questionnaireHandler.SubmitQuestionnaire)
	questionnaireRoutes.Get("/result/:responseId", questionnaireHandler.GetQuestionnaireResult)
	questionnaireRoutes.Get("/result/latest", questionnaireHandler.GetLatestResultByStudent)

	adminQuestionnaireRoutes := questionnaireRoutes.Group("/admin", middleware.AdminRequired())
	adminQuestionnaireRoutes.Post("/generate", questionnaireHandler.GenerateQuestionnaire)
	adminQuestionnaireRoutes.Get("/generate/status/:questionnaireId", questionnaireHandler.GetGenerationStatus)
	adminQuestionnaireRoutes.Get("/", questionnaireHandler.GetAllQuestionnaires)
	adminQuestionnaireRoutes.Get("/:questionnaireId", questionnaireHandler.GetQuestionnaireByID)
	adminQuestionnaireRoutes.Put("/:questionnaireId/activate", questionnaireHandler.ActivateQuestionnaire)
	adminQuestionnaireRoutes.Put("/deactivate", questionnaireHandler.DeactivateQuestionnaire)
	adminQuestionnaireRoutes.Delete("/:questionnaireId", questionnaireHandler.DeleteQuestionnaire)
	adminQuestionnaireRoutes.Post("/:questionnaireId/clone", questionnaireHandler.CloneQuestionnaire)
	adminQuestionnaireRoutes.Get("/:questionnaireId/responses", questionnaireHandler.GetQuestionnaireResponses)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("API available at: http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
