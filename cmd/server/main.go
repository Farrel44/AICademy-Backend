package main

import (
	"log"
	"os"

	"github.com/Farrel44/AICademy-Backend/internal/config"
	"github.com/Farrel44/AICademy-Backend/internal/domain/admin"
	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	authAlumni "github.com/Farrel44/AICademy-Backend/internal/domain/auth/alumni"
	authStudent "github.com/Farrel44/AICademy-Backend/internal/domain/auth/student"
	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	pklAdmin "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/admin"
	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	adminQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire/admin"
	studentQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire/student"
	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	adminRoadmap "github.com/Farrel44/AICademy-Backend/internal/domain/roadmap/admin"
	studentRoadmap "github.com/Farrel44/AICademy-Backend/internal/domain/roadmap/student"
	teacherRoadmap "github.com/Farrel44/AICademy-Backend/internal/domain/roadmap/teacher"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/services/ai"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
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
		log.Printf("Seeding failed: %v", err)
	}

	var aiService ai.AIService

	if geminiAPIKey != "" {
		aiService, err = ai.NewGeminiService(geminiAPIKey)
		if err != nil {
			log.Printf("Failed to initialize Gemini service: %v", err)
			aiService = ai.NewNoAIService()
		}
	} else {
		log.Println("GEMINI_API_KEY not found in environment")
		log.Println("Using NoAI service - AI features will be disabled")
		aiService = ai.NewNoAIService()
	}

	authRepo := auth.NewRepository(db, rdb.Client)

	commonAuthService := commonAuth.NewCommonAuthService(authRepo)
	alumniAuthService := authAlumni.NewAlumniAuthService(authRepo)
	studentAuthService := authStudent.NewStudentAuthService(authRepo)

	commonAuthHandler := commonAuth.NewCommonAuthHandler(commonAuthService)
	alumniAuthHandler := authAlumni.NewAlumniAuthHandler(alumniAuthService)
	studentAuthHandler := authStudent.NewStudentAuthHandler(studentAuthService)

	// Admin service and handler
	adminUserService := admin.NewAdminUserService(authRepo, rdb)
	adminUserHandler := admin.NewAdminUserHandler(adminUserService)

	// Questionnaire services and handlers
	questionnaireRepo := questionnaire.NewQuestionnaireRepository(db)

	// Student questionnaire service and handler
	studentQuestionnaireService := studentQuestionnaire.NewStudentQuestionnaireService(questionnaireRepo, aiService)
	studentQuestionnaireHandler := studentQuestionnaire.NewStudentQuestionnaireHandler(studentQuestionnaireService)

	// Admin questionnaire service and handler
	adminQuestionnaireService := adminQuestionnaire.NewAdminQuestionnaireService(questionnaireRepo, aiService, rdb.Client)
	adminQuestionnaireHandler := adminQuestionnaire.NewAdminQuestionnaireHandler(adminQuestionnaireService)

	// PKL services and handlers
	pklRepo := pkl.NewPklRepository(db, rdb.Client)
	pklAdminService := pklAdmin.NewAdminPklService(pklRepo, rdb)
	pklAdminHandler := pklAdmin.NewPklHandler(pklAdminService)

	// Roadmap services and handlers
	roadmapRepo := roadmap.NewRoadmapRepository(db)

	// Admin roadmap service and handler
	adminRoadmapService := adminRoadmap.NewAdminRoadmapService(roadmapRepo)
	adminRoadmapHandler := adminRoadmap.NewAdminRoadmapHandler(adminRoadmapService)

	// Student roadmap service and handler
	studentRoadmapService := studentRoadmap.NewStudentRoadmapService(roadmapRepo)
	studentRoadmapHandler := studentRoadmap.NewStudentRoadmapHandler(studentRoadmapService)

	// Teacher service and handler
	teacherService := teacherRoadmap.NewTeacherService(roadmapRepo)
	teacherHandler := teacherRoadmap.NewTeacherHandler(teacherService)

	userRepo := user.NewUserRepository(db, rdb.Client)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

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
			"message": "AICademy API v1.0",
			"status":  "OK",
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
	protectedAuth.Get("/me", userHandler.GetUserByToken)
	// Student auth
	studentAuth := authRoutes.Group("/student", middleware.AuthRequired())
	studentAuth.Post("/change-default-password", studentAuthHandler.ChangeDefaultPassword)

	// Admin Auth
	adminAuth := api.Group("/admin", middleware.AuthRequired(), middleware.AdminRequired())
	adminAuth.Post("/students", studentAuthHandler.CreateStudent)
	adminAuth.Post("/students/upload-csv", studentAuthHandler.UploadStudentsCSV)

	adminAuth.Get("/users/statistics", adminUserHandler.GetStatistics)
	// Students
	adminAuth.Get("/users/students", adminUserHandler.GetStudents)
	adminAuth.Get("/users/students/:id", adminUserHandler.GetStudentByID)
	adminAuth.Put("/users/students/:id", adminUserHandler.UpdateStudent)
	adminAuth.Delete("/users/students/:id", adminUserHandler.DeleteStudent)
	// Teachers
	adminAuth.Get("/users/teachers", adminUserHandler.GetTeachers)
	adminAuth.Get("/users/teachers/:id", adminUserHandler.GetTeacherByID)
	adminAuth.Post("/users/teachers", adminUserHandler.CreateTeacher)
	adminAuth.Put("/users/teachers/:id", adminUserHandler.UpdateTeacher)
	adminAuth.Delete("/users/teachers/:id", adminUserHandler.DeleteTeacher)
	// Companies
	adminAuth.Get("/users/companies", adminUserHandler.GetCompanies)
	adminAuth.Get("/users/companies/:id", adminUserHandler.GetCompanyByID)
	adminAuth.Post("/users/companies", adminUserHandler.CreateCompany)
	adminAuth.Put("/users/companies/:id", adminUserHandler.UpdateCompany)
	adminAuth.Delete("/users/companies/:id", adminUserHandler.DeleteCompany)
	// Alumni
	adminAuth.Get("/users/alumni", adminUserHandler.GetAlumni)
	adminAuth.Get("/users/alumni/:id", adminUserHandler.GetAlumniByID)
	adminAuth.Put("/users/alumni/:id", adminUserHandler.UpdateAlumni)
	adminAuth.Delete("/users/alumni/:id", adminUserHandler.DeleteAlumni)

	// Admin Questionnaire Routes
	adminAuth.Post("/questionnaires/target-roles", adminQuestionnaireHandler.CreateTargetRole)
	adminAuth.Get("/questionnaires/target-roles", adminQuestionnaireHandler.GetTargetRoles)
	adminAuth.Get("/questionnaires/target-roles/:id", adminQuestionnaireHandler.GetTargetRole)
	adminAuth.Put("/questionnaires/target-roles/:id", adminQuestionnaireHandler.UpdateTargetRole)
	adminAuth.Delete("/questionnaires/target-roles/:id", adminQuestionnaireHandler.DeleteTargetRole)
	adminAuth.Post("/questionnaires/generate", adminQuestionnaireHandler.GenerateQuestionnaire)
	adminAuth.Get("/questionnaires/generation-status/:id", adminQuestionnaireHandler.GetGenerationStatus)
	adminAuth.Get("/questionnaires", adminQuestionnaireHandler.GetQuestionnaires)
	adminAuth.Get("/responses", adminQuestionnaireHandler.GetQuestionnaireResponses)
	adminAuth.Get("/responses/:id", adminQuestionnaireHandler.GetResponseDetail)
	// Parameterized routes must be last
	adminAuth.Put("/questionnaires/:id/activate", adminQuestionnaireHandler.ActivateQuestionnaire)
	adminAuth.Get("/questionnaires/:id", adminQuestionnaireHandler.GetQuestionnaireDetail)

	// PKL Admin Routes
	adminAuth.Post("/users/internships", pklAdminHandler.CreateInternshipPosition)
	adminAuth.Get("/users/internships", pklAdminHandler.GetInternshipPositions)
	adminAuth.Get("/users/internships/:id", pklAdminHandler.GetInternshipByID)
	adminAuth.Put("/users/internships/:id", pklAdminHandler.UpdateInternshipPosition)
	adminAuth.Delete("/users/internships/:id", pklAdminHandler.DeleteInternshipPosition)

	// Admin Roadmap Routes
	adminAuth.Get("/roadmaps/statistics", adminRoadmapHandler.GetStatistics)
	adminAuth.Get("/roadmaps/submissions", adminRoadmapHandler.GetPendingSubmissions)
	adminAuth.Post("/roadmaps", adminRoadmapHandler.CreateRoadmap)
	adminAuth.Get("/roadmaps", adminRoadmapHandler.GetRoadmaps)
	adminAuth.Get("/roadmaps/:id", adminRoadmapHandler.GetRoadmapByID)
	adminAuth.Put("/roadmaps/:id", adminRoadmapHandler.UpdateRoadmap)
	adminAuth.Delete("/roadmaps/:id", adminRoadmapHandler.DeleteRoadmap)
	adminAuth.Post("/roadmaps/:roadmapId/steps", adminRoadmapHandler.CreateRoadmapStep)
	adminAuth.Put("/roadmaps/steps/:stepId", adminRoadmapHandler.UpdateRoadmapStep)
	adminAuth.Delete("/roadmaps/steps/:stepId", adminRoadmapHandler.DeleteRoadmapStep)
	adminAuth.Put("/roadmaps/steps/reorder", adminRoadmapHandler.UpdateStepOrders)
	adminAuth.Get("/roadmaps/:roadmapId/progress", adminRoadmapHandler.GetStudentProgress)

	// Teacher Routes (for reviewing submissions)
	teacherAuth := api.Group("/teacher", middleware.AuthRequired(), middleware.TeacherOrAdminRequired())
	teacherAuth.Get("/roadmaps/submissions", teacherHandler.GetPendingSubmissions)
	teacherAuth.Post("/roadmaps/submissions/:submissionId/review", teacherHandler.ReviewSubmission)

	// Student Questionnaire Routes
	studentRoutes := api.Group("/student", middleware.AuthRequired(), middleware.StudentRequired())
	studentRoutes.Get("/questionnaire/active", studentQuestionnaireHandler.GetActiveQuestionnaire)
	studentRoutes.Post("/questionnaire/submit", studentQuestionnaireHandler.SubmitQuestionnaire)
	studentRoutes.Get("/questionnaire/latest-result", studentQuestionnaireHandler.GetLatestQuestionnaireResult)
	studentRoutes.Get("/questionnaire/result/:id", studentQuestionnaireHandler.GetQuestionnaireResult)
	studentRoutes.Get("/role", studentQuestionnaireHandler.GetStudentRole)

	// Student Roadmap Routes
	studentRoutes.Get("/roadmaps", studentRoadmapHandler.GetAvailableRoadmaps)
	studentRoutes.Post("/roadmaps/start", studentRoadmapHandler.StartRoadmap)
	studentRoutes.Get("/roadmaps/:roadmapId/progress", studentRoadmapHandler.GetRoadmapProgress)
	studentRoutes.Post("/roadmaps/steps/start", studentRoadmapHandler.StartStep)
	studentRoutes.Post("/roadmaps/steps/submit", studentRoadmapHandler.SubmitEvidence)
	studentRoutes.Get("/roadmaps/my-progress", studentRoadmapHandler.GetMyProgress)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("API at: http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
