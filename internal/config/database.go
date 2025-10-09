package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	pkl_model "github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/Farrel44/AICademy-Backend/internal/domain/project"

	"github.com/Farrel44/AICademy-Backend/internal/domain/questionnaire"
	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDatabase() (*gorm.DB, error) {
	host := os.Getenv("DB_HOST")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslMode := os.Getenv("DB_SSLMODE")

	if host == "" {
		host = "localhost"
	}
	if username == "" {
		username = "postgres"
	}
	if password == "" {
		password = "password"
	}
	if databaseName == "" {
		databaseName = "aicademy"
	}
	if sslMode == "" {
		sslMode = "disable"
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		portInt = 5432
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Jakarta",
		host, username, password, databaseName, portInt, sslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		// User models
		&user.User{},
		&user.StudentProfile{},
		&user.AlumniProfile{},
		&user.TeacherProfile{},
		&user.CompanyProfile{},
		&user.ResetPasswordToken{},
		&user.RefreshToken{},

		// Questionnaire models
		&questionnaire.ProfilingQuestionnaire{},
		&questionnaire.QuestionnaireQuestion{},
		&questionnaire.QuestionnaireResponse{},

		&questionnaire.QuestionGenerationTemplate{},
		&project.TargetRole{},
		&questionnaire.QuestionnaireTargetRole{},

		&roadmap.FeatureRoadmap{},
		&roadmap.RoadmapStep{},
		&roadmap.StudentRoadmapProgress{},
		&roadmap.StudentStepProgress{},

		// PKL/Internship models
		&pkl_model.Internship{},
		&pkl_model.InternshipApplication{},
		&pkl_model.InternshipReview{},

		// Project models
		&project.Project{},
		&project.ProjectContributor{},
		&project.ProjectPhoto{},
		&project.Certification{},
		&project.CertificationPhoto{},
	)

	if err != nil {
		return nil, err
	}

	addIndexes(db)

	log.Println("Database migration completed successfully")
	return db, nil
}

func addIndexes(db *gorm.DB) {
	// User indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)")

	// Student indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_profiles_nis ON student_profiles(nis)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_profiles_user_id ON student_profiles(user_id)")

	// Alumni indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_alumni_profiles_user_id ON alumni_profiles(user_id)")

	// Teacher indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_teacher_profiles_user_id ON teacher_profiles(user_id)")

	// Company indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_company_profiles_user_id ON company_profiles(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_company_profiles_company_name ON company_profiles(company_name)")

	// Refresh token indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)")

	// Reset password token indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reset_password_tokens_user_id ON reset_password_tokens(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reset_password_tokens_token ON reset_password_tokens(token)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_reset_password_tokens_expires_at ON reset_password_tokens(expires_at)")

	// Questionnaire indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questionnaires_active ON profiling_questionnaires(active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questionnaires_generated_by ON profiling_questionnaires(generated_by)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questionnaires_generation_status ON profiling_questionnaires(generation_status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questionnaires_generation_updated ON profiling_questionnaires(generation_updated_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questions_questionnaire_id ON questionnaire_questions(questionnaire_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questions_order ON questionnaire_questions(question_order)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_responses_student_id ON questionnaire_responses(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_responses_questionnaire_id ON questionnaire_responses(questionnaire_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_responses_submitted_at ON questionnaire_responses(submitted_at)")

	// Role recommendation indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_role_recommendations_active ON role_recommendations(active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_role_recommendations_category ON role_recommendations(category)")

	// Roadmap indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_feature_roadmaps_profiling_role_id ON feature_roadmaps(profiling_role_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_feature_roadmaps_status ON feature_roadmaps(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_feature_roadmaps_visibility ON feature_roadmaps(visibility)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_roadmap_steps_roadmap_id ON roadmap_steps(roadmap_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_roadmap_steps_order ON roadmap_steps(step_order)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_roadmap_progress_student_id ON student_roadmap_progress(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_roadmap_progress_roadmap_id ON student_roadmap_progress(roadmap_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_step_progress_roadmap_progress_id ON student_step_progress(student_roadmap_progress_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_step_progress_step_id ON student_step_progress(roadmap_step_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_student_step_progress_status ON student_step_progress(status)")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_student_roadmap_progress_unique ON student_roadmap_progress(roadmap_id, student_profile_id)")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_student_step_progress_unique ON student_step_progress(student_roadmap_progress_id, roadmap_step_id)")

	// Target role indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_target_roles_active ON target_roles(active)")

	// PKL/Internship indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_company_profile_id ON internships(company_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_type ON internships(type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_posted_at ON internships(posted_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_deadline ON internships(deadline)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_created_at ON internships(created_at)")

	// Internship application indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_internship_id ON internship_applications(internship_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_student_profile_id ON internship_applications(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_status ON internship_applications(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_applied_at ON internship_applications(applied_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_reviewed_at ON internship_applications(reviewed_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_approved_by ON internship_applications(approved_by)")

	// Internship review indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_reviews_internship_id ON internship_reviews(internship_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_reviews_student_profile_id ON internship_reviews(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_reviews_rating ON internship_reviews(rating)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_reviews_created_at ON internship_reviews(created_at)")

	// Project indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_owner_student_profile_id ON projects(owner_student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_start_date ON projects(start_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_end_date ON projects(end_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_project_name ON projects(project_name)")

	// Project contributor indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_contributors_project_id ON project_contributors(project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_contributors_student_profile_id ON project_contributors(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_contributors_profiling_role_id ON project_contributors(profiling_role_id)")

	// Project photo indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_photos_project_id ON project_photos(project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_photos_is_primary ON project_photos(is_primary)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_photos_created_at ON project_photos(created_at)")

	// Certification indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_student_profile_id ON certifications(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_issue_date ON certifications(issue_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_expiration_date ON certifications(expiration_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_issuing_organization ON certifications(issuing_organization)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_name ON certifications(name)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_created_at ON certifications(created_at)")

	// Certification photo indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certification_photos_certification_id ON certification_photos(certification_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certification_photos_is_primary ON certification_photos(is_primary)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certification_photos_created_at ON certification_photos(created_at)")

	// Composite indexes untuk query yang sering digunakan
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_student_status ON internship_applications(student_profile_id, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internship_applications_internship_status ON internship_applications(internship_id, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_company_type ON internships(company_profile_id, type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_internships_type_deadline ON internships(type, deadline)")

	// Project composite indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_owner_date_range ON projects(owner_student_profile_id, start_date, end_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_project_contributors_student_project ON project_contributors(student_profile_id, project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_student_date ON certifications(student_profile_id, issue_date)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_certifications_expiring ON certifications(expiration_date) WHERE expiration_date IS NOT NULL")

	log.Println("Database indexes created successfully")
}
