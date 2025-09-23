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
	authAlumni "aicademy-backend/internal/domain/auth/alumni"
	authStudent "aicademy-backend/internal/domain/auth/student"
	commonAuth "aicademy-backend/internal/domain/common/auth"
	"aicademy-backend/internal/domain/questionnaire"
	"aicademy-backend/internal/middleware"
	"aicademy-backend/internal/services/ai"
)

func main() {
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db, err := config.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Seeding database...")
	if err := config.SeedData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	var aiService ai.AIService

	if geminiAPIKey != "" {
		aiService, err = ai.NewGeminiService(geminiAPIKey)
		if err != nil {
			log.Printf("Failed to initialize Gemini AI service: %v", err)
			log.Println("Falling back to NoAI service")
			aiService = ai.NewNoAIService()
		}
	} else {
		log.Println("GEMINI_API_KEY not found in environment")
		log.Println("Using NoAI service - AI features will be disabled")
		aiService = ai.NewNoAIService()
	}

	authRepo := auth.NewAuthRepository(db)

	commonAuthService := commonAuth.NewCommonAuthService(authRepo)
	alumniAuthService := authAlumni.NewAlumniAuthService(authRepo)
	studentAuthService := authStudent.NewStudentAuthService(authRepo)

	commonAuthHandler := commonAuth.NewCommonAuthHandler(commonAuthService)
	alumniAuthHandler := authAlumni.NewAlumniAuthHandler(alumniAuthService)
	studentAuthHandler := authStudent.NewStudentAuthHandler(studentAuthService)

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

	// Common Auth
	authRoutes := api.Group("/auth")
	authRoutes.Post("/login", commonAuthHandler.Login)
	authRoutes.Post("/forgot-password", commonAuthHandler.ForgotPassword)
	authRoutes.Post("/reset-password/:token", commonAuthHandler.ResetPassword)
	authRoutes.Post("/refresh", commonAuthHandler.RefreshToken)

	// Alumni Auth
	authRoutes.Post("/register/alumni", alumniAuthHandler.RegisterAlumni)

	// Protected common auth
	protectedAuth := authRoutes.Group("/protected", middleware.AuthRequired())
	protectedAuth.Post("/change-password", commonAuthHandler.ChangePassword)
	protectedAuth.Post("/logout", commonAuthHandler.Logout)

	// Student auth
	studentAuth := authRoutes.Group("/student", middleware.AuthRequired())
	studentAuth.Post("/change-default-password", studentAuthHandler.ChangeDefaultPassword)

	// Admin Auth
	adminAuth := api.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())
	adminAuth.Post("/students", studentAuthHandler.CreateStudent)
	adminAuth.Post("/students/upload-csv", studentAuthHandler.UploadStudentsCSV)

	questionnaireRoutes := api.Group("/questionnaire", middleware.AuthRequired())
	questionnaireRoutes.Get("/active", questionnaireHandler.GetActiveQuestionnaire)
	questionnaireRoutes.Post("/submit", questionnaireHandler.SubmitQuestionnaire)
	questionnaireRoutes.Get("/result/:responseId", questionnaireHandler.GetQuestionnaireResult)
	questionnaireRoutes.Get("/result/latest", questionnaireHandler.GetLatestResultByStudent)

	adminRoutes := api.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())

	adminRoutes.Post("/questionnaires/generate", questionnaireHandler.GenerateQuestionnaire)
	adminRoutes.Get("/questionnaires/generate/status/:questionnaireId", questionnaireHandler.GetGenerationStatus)
	adminRoutes.Get("/questionnaires", questionnaireHandler.GetAllQuestionnaires)
	adminRoutes.Get("/questionnaires/:questionnaireId", questionnaireHandler.GetQuestionnaireByID)
	adminRoutes.Put("/questionnaires/:questionnaireId/activate", questionnaireHandler.ActivateQuestionnaire)
	adminRoutes.Put("/questionnaires/deactivate", questionnaireHandler.DeactivateQuestionnaire)
	adminRoutes.Delete("/questionnaires/:questionnaireId", questionnaireHandler.DeleteQuestionnaire)
	adminRoutes.Get("/questionnaires/:questionnaireId/responses", questionnaireHandler.GetQuestionnaireResponses)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("API at: http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
