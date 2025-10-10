package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/Farrel44/AICademy-Backend/internal/config"
	"github.com/Farrel44/AICademy-Backend/internal/domain/admin"
	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	authAlumni "github.com/Farrel44/AICademy-Backend/internal/domain/auth/alumni"
	authStudent "github.com/Farrel44/AICademy-Backend/internal/domain/auth/student"
	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	pklAdmin "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/admin"
	pklAlumni "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/alumni"
	pklCompany "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/company"
	pklStudent "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/student"
	pklTeacher "github.com/Farrel44/AICademy-Backend/internal/domain/pkl/teacher"
	"github.com/Farrel44/AICademy-Backend/internal/domain/project"

	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	adminQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire/admin"
	studentQuestionnaire "github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire/student"
	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	adminRoadmap "github.com/Farrel44/AICademy-Backend/internal/domain/roadmap/admin"
	studentRoadmap "github.com/Farrel44/AICademy-Backend/internal/domain/roadmap/student"
	teacherRoadmap "github.com/Farrel44/AICademy-Backend/internal/domain/roadmap/teacher"

	// "github.com/Farrel44/AICademy-Backend/internal/domain/trend"
	// trendAdmin "github.com/Farrel44/AICademy-Backend/internal/domain/trend/admin"
	// trendStudent "github.com/Farrel44/AICademy-Backend/internal/domain/trend/student"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/middleware"
	"github.com/Farrel44/AICademy-Backend/internal/services/ai"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	challengeRepo "github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	adminChallenge "github.com/Farrel44/AICademy-Backend/internal/domain/challenge/admin"
	studentChallenge "github.com/Farrel44/AICademy-Backend/internal/domain/challenge/student"
	teacherChallenge "github.com/Farrel44/AICademy-Backend/internal/domain/challenge/teacher"
)

// RouteList prints all registered routes in a formatted table
func RouteList(app *fiber.App, filterPrefix string) {
	type RouteInfo struct {
		Method string
		Path   string
		Name   string
	}

	var routes []RouteInfo
	allowedMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	}

	for _, stack := range app.Stack() {
		for _, route := range stack {
			if !allowedMethods[route.Method] {
				continue
			}

			if filterPrefix == "" || strings.HasPrefix(route.Path, filterPrefix) {
				routes = append(routes, RouteInfo{
					Method: route.Method,
					Path:   route.Path,
					Name:   route.Name,
				})
			}
		}
	}

	// Sort routes by path, then by method
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	// Print header
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🚀 AICademy API - Registered Routes")
	fmt.Println(strings.Repeat("=", 80))

	if filterPrefix != "" {
		fmt.Printf("📝 Filter: %s\n", filterPrefix)
		fmt.Println(strings.Repeat("-", 80))
	}

	// Print table header
	fmt.Printf("%-8s | %-50s | %s\n", "METHOD", "PATH", "HANDLER")
	fmt.Println(strings.Repeat("-", 80))

	// Group routes by prefix for better readability
	currentGroup := ""
	for _, route := range routes {
		// Determine route group
		pathParts := strings.Split(strings.TrimPrefix(route.Path, "/"), "/")
		var group string
		if len(pathParts) >= 3 && pathParts[0] == "api" && pathParts[1] == "v1" {
			group = pathParts[2]
		} else if len(pathParts) >= 1 {
			group = pathParts[0]
		}

		// Print group separator
		if group != currentGroup && group != "" {
			if currentGroup != "" {
				fmt.Println(strings.Repeat("-", 80))
			}
			fmt.Printf("📁 %s\n", strings.ToUpper(group))
			fmt.Println(strings.Repeat("-", 80))
			currentGroup = group
		}

		// Truncate handler name if too long
		handlerName := route.Name
		if len(handlerName) > 25 {
			handlerName = handlerName[:22] + "..."
		}

		// Color code by method
		methodColor := getMethodColor(route.Method)
		fmt.Printf("%-8s | %-50s | %s\n",
			methodColor+route.Method+"\033[0m",
			route.Path,
			handlerName)
	}

	// Print footer
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("📊 Total Routes: %d\n", len(routes))

	// Group statistics
	methodStats := make(map[string]int)
	for _, route := range routes {
		methodStats[route.Method]++
	}

	fmt.Println("\n📈 Method Statistics:")
	for method, count := range methodStats {
		fmt.Printf("  %s: %d\n", method, count)
	}

	fmt.Println(strings.Repeat("=", 80))
}

// getMethodColor returns ANSI color code for HTTP methods
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[32m" // Green
	case "POST":
		return "\033[34m" // Blue
	case "PUT":
		return "\033[33m" // Yellow
	case "DELETE":
		return "\033[31m" // Red
	case "PATCH":
		return "\033[35m" // Magenta
	case "OPTIONS":
		return "\033[36m" // Cyan
	default:
		return "\033[0m" // Default
	}
}

// CLI flags for route listing
var (
	routeListFlag   = flag.Bool("route:list", false, "print all registered Fiber routes and exit")
	routePrefixFlag = flag.String("route:prefix", "", "optional prefix filter, e.g. /api/v1/admin")
)

func main() {
	// Parse CLI flags early
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")

	rdb := utils.NewRedisClient()
	defer func() {
		_ = rdb.Close()
	}()

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
	adminRoadmapService := adminRoadmap.NewAdminRoadmapService(roadmapRepo, rdb.Client)
	adminRoadmapHandler := adminRoadmap.NewAdminRoadmapHandler(adminRoadmapService)

	// Student roadmap service and handler
	studentRoadmapService := studentRoadmap.NewStudentRoadmapService(roadmapRepo)
	studentRoadmapHandler := studentRoadmap.NewStudentRoadmapHandler(studentRoadmapService)

	// Teacher service and handler
	teacherService := teacherRoadmap.NewTeacherService(roadmapRepo, rdb.Client)
	teacherHandler := teacherRoadmap.NewTeacherHandler(teacherService)

	userRepo := user.NewUserRepository(db, rdb.Client)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	// Project services and handlers
	projectRepo := project.NewProjectRepository(db, rdb.Client)
	projectService := project.NewProjectService(projectRepo, rdb)
	projectHandler := project.NewProjectHandler(projectService)

	pklStudentRepo := pkl.NewPklRepository(db, rdb.Client)
	pklStudentService := pklStudent.NewStudentPklService(pklStudentRepo, rdb)
	pklStudentHandler := pklStudent.NewStudentPklHandler(pklStudentService)

	pklAlumniRepo := pkl.NewPklRepository(db, rdb.Client)
	pklAlumniService := pklAlumni.NewAlumniPklService(pklAlumniRepo, rdb)
	pklAlumniHandler := pklAlumni.NewAlumniPklHandler(pklAlumniService)

	pklCompanyRepo := pkl.NewPklRepository(db, rdb.Client)
	pklCompanyService := pklCompany.NewCompanyPklService(pklCompanyRepo, rdb)
	pklCompanyHandler := pklCompany.NewCompanyPklHandler(pklCompanyService)

	pklTeacherRepo := pkl.NewPklRepository(db, rdb.Client)
	pklTeacherService := pklTeacher.NewTeacherPklService(pklTeacherRepo, rdb)
	pklTeacherHandler := pklTeacher.NewTeacherPklHandler(pklTeacherService)

	// Challenge services and handlers
	challengeRepository := challengeRepo.NewChallengeRepository(db, rdb.Client)

	// Admin challenge
	adminChallengeService := adminChallenge.NewAdminChallengeService(challengeRepository, rdb.Client)
	adminChallengeHandler := adminChallenge.NewAdminChallengeHandler(adminChallengeService)

	// // Trend services and handlers
	// trendRepo := trend.NewTrendRepository(db, rdb.Client)
	// adminTrendService := trendAdmin.NewAdminTrendService(trendRepo, rdb.Client, aiService)
	// adminTrendHandler := trendAdmin.NewAdminTrendHandler(adminTrendService)

	// studentTrendService := trendStudent.NewTrendService(trendRepo, rdb.Client)
	// studentTrendHandler := trendStudent.NewTrendHandler(studentTrendService)

	// Teacher challenge
	teacherChallengeService := teacherChallenge.NewTeacherChallengeService(challengeRepository, rdb.Client)
	teacherChallengeHandler := teacherChallenge.NewTeacherChallengeHandler(teacherChallengeService)

	// Student challenge
	studentChallengeService := studentChallenge.NewStudentChallengeService(challengeRepository, rdb)
	studentChallengeHandler := studentChallenge.NewStudentChallengeHandler(studentChallengeService)

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

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"message": "AICademy API is healthy",
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

	// Submission routes
	adminAuth.Get("/internship/:id/submissions", pklAdminHandler.GetSubmissionsByInternshipID)
	adminAuth.Get("/company/:id/internships-with-submissions", pklAdminHandler.GetInternshipsWithSubmissionsByCompanyID)
	adminAuth.Get("/submission/:id", pklAdminHandler.GetSubmissionByID)
	adminAuth.Put("/submission/:id/status", pklAdminHandler.UpdateSubmissionStatus)

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
	adminAuth.Get("/roadmaps/:roadmapId/progress", adminRoadmapHandler.GetStudentProgress)

	// Admin Challenge Routes
	adminAuth.Post("/challenges", adminChallengeHandler.CreateChallenge)
	adminAuth.Get("/challenges", adminChallengeHandler.GetAllChallenges)
	adminAuth.Get("/challenges/submissions", adminChallengeHandler.GetAllSubmissions)
	adminAuth.Post("/challenges/submissions/score", adminChallengeHandler.ScoreSubmission)
	adminAuth.Get("/challenges/:id", adminChallengeHandler.GetChallengeByID)
	adminAuth.Put("/challenges/:id", adminChallengeHandler.UpdateChallenge)
	adminAuth.Delete("/challenges/:id", adminChallengeHandler.DeleteChallenge)

	// Admin Trend Routes
	// adminAuth.Post("/trends/collect", adminTrendHandler.TriggerDataCollection)
	// adminAuth.Get("/trends/status", adminTrendHandler.GetCollectionStatus)
	// adminAuth.Get("/trends/data", adminTrendHandler.GetAllTrendData)
	// adminAuth.Get("/challenges/leaderboard", adminChallengeHandler.GetLeaderboard)

	// Teacher Routes (for reviewing submissions)
	teacherAuth := api.Group("/teacher", middleware.AuthRequired(), middleware.TeacherOrAdminRequired())
	teacherAuth.Get("/roadmaps/submissions", teacherHandler.GetPendingSubmissions)
	teacherAuth.Post("/roadmaps/submissions/:submissionId/review", teacherHandler.ReviewSubmission)

	teacherAuth.Get("/internships", pklTeacherHandler.GetAllInternships)
	teacherAuth.Get("/internship/:id/submissions", pklTeacherHandler.GetSubmissionsByInternshipID)
	teacherAuth.Get("/company/:id/internships-with-submissions", pklTeacherHandler.GetInternshipsWithSubmissionsByCompanyID)
	teacherAuth.Get("/submission/:id", pklTeacherHandler.GetSubmissionByID)
	teacherAuth.Put("/submission/:id/status", pklTeacherHandler.UpdateSubmissionStatus)

	// Teacher Challenge Routes
	teacherAuth.Post("/challenges", teacherChallengeHandler.CreateChallenge)
	teacherAuth.Get("/challenges", teacherChallengeHandler.GetMyChallenges)
	teacherAuth.Get("/challenges/:id", teacherChallengeHandler.GetChallengeByID)
	teacherAuth.Put("/challenges/:id", teacherChallengeHandler.UpdateChallenge)
	teacherAuth.Delete("/challenges/:id", teacherChallengeHandler.DeleteChallenge)
	teacherAuth.Get("/challenges/submissions", teacherChallengeHandler.GetMySubmissions)
	teacherAuth.Post("/challenges/submissions/score", teacherChallengeHandler.ScoreSubmission)
	teacherAuth.Get("/challenges/leaderboard", teacherChallengeHandler.GetLeaderboard)

	// Student Questionnaire Routes
	studentRoutes := api.Group("/student", middleware.AuthRequired(), middleware.StudentRequired())
	studentRoutes.Get("/questionnaire/active", studentQuestionnaireHandler.GetActiveQuestionnaire)
	studentRoutes.Post("/questionnaire/submit", studentQuestionnaireHandler.SubmitQuestionnaire)
	studentRoutes.Get("/questionnaire/latest-result", studentQuestionnaireHandler.GetLatestQuestionnaireResult)
	studentRoutes.Get("/questionnaire/result/:id", studentQuestionnaireHandler.GetQuestionnaireResult)
	studentRoutes.Get("/role", studentQuestionnaireHandler.GetStudentRole)

	// Student Roadmap Routes
	studentRoutes.Get("/my-roadmap", studentRoadmapHandler.GetMyRoadmap)
	studentRoutes.Post("/roadmaps/start", studentRoadmapHandler.StartRoadmap)
	studentRoutes.Get("/roadmaps/:roadmapId/progress", studentRoadmapHandler.GetRoadmapProgress)
	studentRoutes.Get("/roadmaps/steps/:stepId/progress", studentRoadmapHandler.GetStepProgress)
	studentRoutes.Post("/roadmaps/steps/start", studentRoadmapHandler.StartStep)
	studentRoutes.Post("/roadmaps/steps/submit", studentRoadmapHandler.SubmitEvidence)

	studentRoutes.Get("/me", userHandler.GetUserByToken)
	studentRoutes.Put("/profile", userHandler.UpdateUserProfile)

	// Project Routes for Students
	studentRoutes.Post("/projects", projectHandler.CreateProject)
	studentRoutes.Get("/projects", projectHandler.GetMyProjects)
	studentRoutes.Get("/projects/:id", projectHandler.GetProjectByID)
	studentRoutes.Put("/projects/:id", projectHandler.UpdateProject)
	studentRoutes.Delete("/projects/:id", projectHandler.DeleteProject)
	studentRoutes.Post("/projects/:id/contributors", projectHandler.AddProjectContributor)

	// Certification Routes for Students
	studentRoutes.Post("/certifications", projectHandler.CreateCertification)
	studentRoutes.Get("/certifications", projectHandler.GetMyCertifications)
	studentRoutes.Get("/certifications/:id", projectHandler.GetCertificationByID)
	studentRoutes.Put("/certifications/:id", projectHandler.UpdateCertification)
	studentRoutes.Delete("/certifications/:id", projectHandler.DeleteCertification)

	// Student PKL Routes
	studentRoutes.Get("/internships", pklStudentHandler.GetInternships)
	studentRoutes.Post("/internship/apply", pklStudentHandler.ApplyPklPosition)

	// Student Challenge Routes
	studentRoutes.Post("/challenges/teams", studentChallengeHandler.CreateTeam)
	studentRoutes.Get("/challenges/teams", studentChallengeHandler.GetMyTeams)
	studentRoutes.Get("/challenges", studentChallengeHandler.GetAvailableChallenges)
	studentRoutes.Post("/challenges/register", studentChallengeHandler.RegisterTeamToChallenge)
	studentRoutes.Post("/challenges/students/search", studentChallengeHandler.SearchStudents)

	// Student Trend Routes
	// studentRoutes.Get("/trends/dashboard", studentTrendHandler.GetTrendDashboard)
	// studentRoutes.Get("/trends", studentTrendHandler.GetCareerTrends)

	// Alumni Routes
	alumniRoutes := api.Group("/alumni", middleware.AuthRequired(), middleware.AlumniRequired())
	alumniRoutes.Get("/internships", pklAlumniHandler.GetAvailablePositions)
	alumniRoutes.Post("/internship/apply", pklAlumniHandler.ApplyPklPosition)
	alumniRoutes.Get("/applications", pklAlumniHandler.GetMyApplications)
	alumniRoutes.Get("/applications/:id", pklAlumniHandler.GetApplicationByID)
	alumniRoutes.Get("/me", userHandler.GetUserByToken)
	alumniRoutes.Put("/profile", userHandler.UpdateUserProfile)

	// Project Routes for Alumni
	alumniRoutes.Post("/projects", projectHandler.CreateProject)
	alumniRoutes.Get("/projects", projectHandler.GetMyProjects)
	alumniRoutes.Get("/projects/:id", projectHandler.GetProjectByID)
	alumniRoutes.Put("/projects/:id", projectHandler.UpdateProject)
	alumniRoutes.Delete("/projects/:id", projectHandler.DeleteProject)
	alumniRoutes.Post("/projects/:id/contributors", projectHandler.AddProjectContributor)

	// Certification Routes for Alumni
	alumniRoutes.Post("/certifications", projectHandler.CreateCertification)
	alumniRoutes.Get("/certifications", projectHandler.GetMyCertifications)
	alumniRoutes.Get("/certifications/:id", projectHandler.GetCertificationByID)
	alumniRoutes.Put("/certifications/:id", projectHandler.UpdateCertification)
	alumniRoutes.Delete("/certifications/:id", projectHandler.DeleteCertification)

	// Company Routes
	companyRoutes := api.Group("/company", middleware.AuthRequired(), middleware.CompanyRequired())
	companyRoutes.Get("/internships", pklCompanyHandler.GetMyInternships)
	companyRoutes.Post("/internships", pklCompanyHandler.CreateInternship)
	companyRoutes.Put("/internships/:id", pklCompanyHandler.UpdateInternship)
	companyRoutes.Delete("/internships/:id", pklCompanyHandler.DeleteInternship)
	companyRoutes.Get("/internships/:id/applications", pklCompanyHandler.GetInternshipApplications)
	companyRoutes.Put("/applications/:id/status", pklCompanyHandler.UpdateApplicationStatus)
	companyRoutes.Get("/me", userHandler.GetUserByToken)
	companyRoutes.Put("/profile", userHandler.UpdateUserProfile)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8000"
	}

	// Check for route listing flags/env vars AFTER all routes are registered
	if *routeListFlag || os.Getenv("ROUTE_LIST") == "1" {
		prefix := *routePrefixFlag
		if envPrefix := os.Getenv("ROUTE_PREFIX"); envPrefix != "" {
			prefix = envPrefix
		}
		RouteList(app, prefix)
		return // Exit after displaying routes
	}

	// Start server
	log.Printf("🚀 AICademy API starting...")
	log.Printf("📡 Server running at: http://localhost:%s", port)
	log.Printf("📚 API Documentation: http://localhost:%s/api/v1", port)
	log.Printf("🔧 Environment: %s", os.Getenv("APP_ENV"))

	// Show route count on startup
	routeCount := 0
	for _, stack := range app.Stack() {
		routeCount += len(stack)
	}
	log.Printf("📋 Total registered routes: %d", routeCount)

	log.Fatal(app.Listen(":" + port))
}
