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
	"aicademy-backend/internal/domain/questionnaire"
	"aicademy-backend/internal/middleware"
	"aicademy-backend/internal/services/ai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Seed data
	if err := config.SeedData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	var aiService ai.AIService
	if geminiAPIKey != "" {
		aiService, err = ai.NewGeminiService(geminiAPIKey)
		if err != nil {
			log.Printf("Warning: Failed to initialize AI service: %v", err)
		} else {
			log.Println("AI service initialized successfully")
		}
	} else {
		log.Println("Warning: GEMINI_API_KEY not found, AI features will be disabled")
	}

	// Initialize repositories and services
	authRepo := auth.NewAuthRepository(db)
	authService := auth.NewAuthService(authRepo)
	authHandler := auth.NewAuthHandler(authService)

	questionnaireRepo := questionnaire.NewQuestionnaireRepository(db)
	questionnaireService := questionnaire.NewQuestionnaireService(questionnaireRepo, aiService)
	questionnaireHandler := questionnaire.NewQuestionnaireHandler(questionnaireService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "AICademy API v1.0",
		ServerHeader: "Fiber",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3001,http://localhost:3000",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "AICademy API is running!",
			"version": "1.0.0",
			"status":  "OK",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes
	authRoutes := api.Group("/auth")
	authRoutes.Post("/register/alumni", authHandler.RegisterAlumni)
	authRoutes.Post("/login", authHandler.Login)
	authRoutes.Post("/forgot-password", authHandler.ForgotPassword)
	authRoutes.Post("/reset-password/:token", authHandler.ResetPassword)

	// Protected auth routes
	protectedAuth := authRoutes.Group("/", middleware.AuthRequired())
	protectedAuth.Post("/change-password", authHandler.ChangePassword)
	protectedAuth.Post("/logout", authHandler.Logout)

	// Admin routes
	adminRoutes := api.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())
	adminRoutes.Post("/create-teacher", authHandler.CreateTeacher)
	adminRoutes.Post("/create-student", authHandler.CreateStudent)
	adminRoutes.Post("/create-company", authHandler.CreateCompany)
	adminRoutes.Post("/upload-students-csv", authHandler.UploadStudentsCSV)

	// Student/General questionnaire routes
	questionnaireRoutes := api.Group("/questionnaire", middleware.AuthRequired())
	questionnaireRoutes.Get("/active", questionnaireHandler.GetActiveQuestionnaire)
	questionnaireRoutes.Post("/submit", questionnaireHandler.SubmitQuestionnaire)
	questionnaireRoutes.Get("/result/:responseId", questionnaireHandler.GetQuestionnaireResult)
	questionnaireRoutes.Get("/result/latest", questionnaireHandler.GetLatestResultByStudent)

	// Admin questionnaire routes
	adminQuestionnaireRoutes := questionnaireRoutes.Group("/admin", middleware.AdminRequired())
	adminQuestionnaireRoutes.Post("/generate", questionnaireHandler.GenerateQuestionnaire)
	adminQuestionnaireRoutes.Get("/generate/status/:questionnaireId", questionnaireHandler.GetGenerationStatus)
	adminQuestionnaireRoutes.Get("/", questionnaireHandler.GetAllQuestionnaires)
	adminQuestionnaireRoutes.Get("/search", questionnaireHandler.SearchQuestionnaires)
	adminQuestionnaireRoutes.Get("/type", questionnaireHandler.GetQuestionnairesByType)
	adminQuestionnaireRoutes.Get("/stats", questionnaireHandler.GetQuestionnaireStats)
	adminQuestionnaireRoutes.Get("/analytics", questionnaireHandler.GetResponseAnalytics)
	adminQuestionnaireRoutes.Get("/:questionnaireId", questionnaireHandler.GetQuestionnaireByID)
	adminQuestionnaireRoutes.Put("/:questionnaireId/activate", questionnaireHandler.ActivateQuestionnaire)
	adminQuestionnaireRoutes.Put("/deactivate", questionnaireHandler.DeactivateQuestionnaire)
	adminQuestionnaireRoutes.Delete("/:questionnaireId", questionnaireHandler.DeleteQuestionnaire)
	adminQuestionnaireRoutes.Post("/:questionnaireId/clone", questionnaireHandler.CloneQuestionnaire)
	adminQuestionnaireRoutes.Get("/:questionnaireId/responses", questionnaireHandler.GetQuestionnaireResponses)

	// Question management routes
	questionRoutes := adminQuestionnaireRoutes.Group("/:questionnaireId/questions")
	questionRoutes.Post("/", questionnaireHandler.AddQuestionToQuestionnaire)
	questionRoutes.Put("/:questionId", questionnaireHandler.UpdateQuestion)
	questionRoutes.Delete("/:questionId", questionnaireHandler.DeleteQuestion)
	questionRoutes.Put("/order", questionnaireHandler.UpdateQuestionOrder)

	// Template routes
	templateRoutes := adminQuestionnaireRoutes.Group("/templates")
	templateRoutes.Post("/", questionnaireHandler.CreateQuestionTemplate)
	templateRoutes.Get("/", questionnaireHandler.GetQuestionTemplates)

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
