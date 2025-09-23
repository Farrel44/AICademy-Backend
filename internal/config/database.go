package config

import (
	"aicademy-backend/internal/domain/questionnaire"
	"aicademy-backend/internal/domain/user"
	"fmt"
	"log"
	"os"
	"strconv"

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

	// Auto migrate semua model
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
		&questionnaire.RoleRecommendation{},
		&questionnaire.QuestionGenerationTemplate{},
		&questionnaire.TargetRole{},
		&questionnaire.QuestionnaireTargetRole{},
	)

	if err != nil {
		return nil, err
	}

	// Tambahkan indexes untuk performa
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

	// Refresh token indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)")

	// Questionnaire indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questionnaires_active ON profiling_questionnaires(active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questionnaires_generated_by ON profiling_questionnaires(generated_by)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questions_questionnaire_id ON questionnaire_questions(questionnaire_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_questions_order ON questionnaire_questions(question_order)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_responses_student_id ON questionnaire_responses(student_profile_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_responses_questionnaire_id ON questionnaire_responses(questionnaire_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_responses_submitted_at ON questionnaire_responses(submitted_at)")

	// Role recommendation indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_role_recommendations_active ON role_recommendations(active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_role_recommendations_category ON role_recommendations(category)")
}
