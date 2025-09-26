package main

import (
	"log"
	"os"

	"github.com/Farrel44/AICademy-Backend/internal/config"
	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	authAlumni "github.com/Farrel44/AICademy-Backend/internal/domain/auth/alumni"
	authStudent "github.com/Farrel44/AICademy-Backend/internal/domain/auth/student"
	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	commonQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/common/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	adminQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire/admin"
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/services/ai"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/Farrel44/AICademy-Backend/internal/config"
	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	authAlumni "github.com/Farrel44/AICademy-Backend/internal/domain/auth/alumni"
	authStudent "github.com/Farrel44/AICademy-Backend/internal/domain/auth/student"
	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	commonQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/common/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	adminQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire/admin"
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/services/ai"
	"gorm.io/gorm/logger"
)

func main() {
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	rdb := utils.NewRedisClient()
	defer func() {
		_ = rdb.Close()
	}()

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

	// Questionnaire services and handlers
	questionnaireRepo := questionnaire.NewQuestionnaireRepository(db)

	// Common questionnaire service and handler
	commonQuestionnaireService := commonQuestionnaire.NewCommonQuestionnaireService(questionnaireRepo, aiService)
	commonQuestionnaireHandler := commonQuestionnaire.NewCommonQuestionnaireHandler(commonQuestionnaireService)

	// Admin questionnaire service and handler
	adminQuestionnaireService := adminQuestionnaire.NewAdminQuestionnaireService(questionnaireRepo, aiService)
	adminQuestionnaireHandler := adminQuestionnaire.NewAdminQuestionnaireHandler(adminQuestionnaireService)

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

	// Admin Questionnaire Routes
	adminAuth.Post("/questionnaires/target-roles", adminQuestionnaireHandler.CreateTargetRole)
	adminAuth.Get("/questionnaires/target-roles", adminQuestionnaireHandler.GetTargetRoles)
	adminAuth.Get("/questionnaires/target-roles/:id", adminQuestionnaireHandler.GetTargetRole)
	adminAuth.Put("/questionnaires/target-roles/:id", adminQuestionnaireHandler.UpdateTargetRole)
	adminAuth.Delete("/questionnaires/target-roles/:id", adminQuestionnaireHandler.DeleteTargetRole)
	adminAuth.Post("/questionnaires/generate", adminQuestionnaireHandler.GenerateQuestionnaire)
	adminAuth.Get("/questionnaires", adminQuestionnaireHandler.GetQuestionnaires)
	adminAuth.Get("/responses", adminQuestionnaireHandler.GetQuestionnaireResponses)
	adminAuth.Get("/responses/:id", adminQuestionnaireHandler.GetResponseDetail)
	// Parameterized routes must be last
	adminAuth.Put("/questionnaires/:id/activate", adminQuestionnaireHandler.ActivateQuestionnaire)
	adminAuth.Get("/questionnaires/:id", adminQuestionnaireHandler.GetQuestionnaireDetail)

	// Common Questionnaire
	questionnaireRoutes := api.Group("/questionnaire", middleware.AuthRequired())
	questionnaireRoutes.Get("/active", commonQuestionnaireHandler.GetActiveQuestionnaire)
	questionnaireRoutes.Post("/submit", commonQuestionnaireHandler.SubmitQuestionnaire)
	questionnaireRoutes.Get("/result/:responseId", commonQuestionnaireHandler.GetQuestionnaireResult)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("API at: http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
