package auth

import (
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewRepository(db *gorm.DB, rdb *redis.Client) *AuthRepository {
	return &AuthRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

func (r *AuthRepository) GetCompanyProfileByUserID(userID uuid.UUID) (*user.CompanyProfile, error) {
	var profile user.CompanyProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *AuthRepository) CreateUser(u *user.User) error {
	return r.db.Create(u).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*user.User, error) {
	var u user.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) GetUserByID(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	return r.db.Model(&user.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword).Error
}

func (r *AuthRepository) CreateAlumniProfile(profile *user.AlumniProfile) error {
	return r.db.Create(profile).Error
}

func (r *AuthRepository) CreateStudentProfile(profile *user.StudentProfile) error {
	return r.db.Create(profile).Error
}

func (r *AuthRepository) CreateTeacherProfile(profile *user.TeacherProfile) error {
	return r.db.Create(profile).Error
}

func (r *AuthRepository) CreateCompanyProfile(profile *user.CompanyProfile) error {
	return r.db.Create(profile).Error
}
func (r *AuthRepository) CreateStudentsBulk(users []user.User, profiles []user.StudentProfile) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&users).Error; err != nil {
			return err
		}
		if err := tx.Create(&profiles).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *AuthRepository) SaveResetToken(email, token string, expiresAt time.Time) error {
	return r.db.Model(&user.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"password_reset_token": token,
		"password_reset_at":    expiresAt,
	}).Error
}

func (r *AuthRepository) GetUserByResetToken(token string) (*user.User, error) {
	var u user.User
	err := r.db.Where("password_reset_token = ? AND password_reset_at > ?", token, time.Now()).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) ClearResetToken(userID uuid.UUID) error {
	return r.db.Model(&user.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_reset_token": "",
		"password_reset_at":    nil,
	}).Error
}

func (r *AuthRepository) CheckNISExists(nis string) (bool, error) {
	var count int64
	err := r.db.Model(&user.StudentProfile{}).Where("nis = ?", nis).Count(&count).Error
	return count > 0, err
}

// Additional methods for getting profiles by user ID
func (r *AuthRepository) GetStudentProfileByUserID(userID uuid.UUID) (*user.StudentProfile, error) {
	var profile user.StudentProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *AuthRepository) GetAlumniProfileByUserID(userID uuid.UUID) (*user.AlumniProfile, error) {
	var profile user.AlumniProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *AuthRepository) GetTeacherProfileByUserID(userID uuid.UUID) (*user.TeacherProfile, error) {
	var profile user.TeacherProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

// Refresh Token Management
func (r *AuthRepository) CreateRefreshToken(refreshToken *user.RefreshToken) error {
	return r.db.Create(refreshToken).Error
}

func (r *AuthRepository) GetRefreshTokenByToken(token string) (*user.RefreshToken, error) {
	var refreshToken user.RefreshToken
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *AuthRepository) DeleteRefreshToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&user.RefreshToken{}).Error
}

func (r *AuthRepository) DeleteAllRefreshTokensByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&user.RefreshToken{}).Error
}

func (r *AuthRepository) CleanupExpiredRefreshTokens() error {
	return r.db.Where("expires_at <= ?", time.Now()).Delete(&user.RefreshToken{}).Error
}

type StudentStatistics struct {
	TotalStudents int64
	TotalRPL      int64
	TotalTKJ      int64
}

func (r *AuthRepository) GetStudents(offset, limit int, search string) ([]user.StudentProfile, int64, error) {
	var students []user.StudentProfile
	var total int64

	query := r.db.Preload("User").Model(&user.StudentProfile{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Joins("JOIN users ON users.id = student_profiles.user_id").
			Where("LOWER(student_profiles.fullname) LIKE ? OR LOWER(users.email) LIKE ? OR student_profiles.nis LIKE ? OR LOWER(student_profiles.class) LIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("student_profiles.created_at DESC").Find(&students).Error
	return students, total, err
}

func (r *AuthRepository) GetStudentStatistics() (*StudentStatistics, error) {
	var stats StudentStatistics

	err := r.db.Model(&user.StudentProfile{}).Count(&stats.TotalStudents).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&user.StudentProfile{}).
		Where("LOWER(class) LIKE ?", "%rpl%").
		Count(&stats.TotalRPL).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&user.StudentProfile{}).
		Where("LOWER(class) LIKE ?", "%tkj%").
		Count(&stats.TotalTKJ).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *AuthRepository) GetStudentByID(id uuid.UUID) (*user.StudentProfile, error) {
	var student user.StudentProfile
	err := r.db.Preload("User").Where("id = ?", id).First(&student).Error
	return &student, err
}

func (r *AuthRepository) UpdateStudentProfile(student *user.StudentProfile) error {
	return r.db.Save(student).Error
}

func (r *AuthRepository) DeleteStudent(userID, profileID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&user.StudentProfile{}, "id = ?", profileID).Error; err != nil {
			return err
		}
		return tx.Delete(&user.User{}, "id = ?", userID).Error
	})
}

func (r *AuthRepository) GetTeachers(offset, limit int, search string) ([]user.TeacherProfile, int64, error) {
	var teachers []user.TeacherProfile
	var total int64

	query := r.db.Preload("User").Model(&user.TeacherProfile{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Joins("JOIN users ON users.id = teacher_profiles.user_id").
			Where("LOWER(teacher_profiles.fullname) LIKE ? OR LOWER(users.email) LIKE ?",
				searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("teacher_profiles.created_at DESC").Find(&teachers).Error
	return teachers, total, err
}

func (r *AuthRepository) GetTeacherByID(id uuid.UUID) (*user.TeacherProfile, error) {
	var teacher user.TeacherProfile
	err := r.db.Preload("User").Where("id = ?", id).First(&teacher).Error
	return &teacher, err
}

func (r *AuthRepository) UpdateTeacherProfile(teacher *user.TeacherProfile) error {
	return r.db.Save(teacher).Error
}

func (r *AuthRepository) DeleteTeacher(userID, profileID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&user.TeacherProfile{}, "id = ?", profileID).Error; err != nil {
			return err
		}
		return tx.Delete(&user.User{}, "id = ?", userID).Error
	})
}

func (r *AuthRepository) GetCompanies(offset, limit int, search string) ([]user.CompanyProfile, int64, error) {
	var companies []user.CompanyProfile
	var total int64

	query := r.db.Preload("User").Model(&user.CompanyProfile{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Joins("JOIN users ON users.id = company_profiles.user_id").
			Where("LOWER(company_profiles.company_name) LIKE ? OR LOWER(users.email) LIKE ?",
				searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("company_profiles.created_at DESC").Find(&companies).Error
	return companies, total, err
}

func (r *AuthRepository) GetCompanyByID(id uuid.UUID) (*user.CompanyProfile, error) {
	var company user.CompanyProfile
	err := r.db.Preload("User").Where("id = ?", id).First(&company).Error
	return &company, err
}

func (r *AuthRepository) UpdateCompanyProfile(company *user.CompanyProfile) error {
	return r.db.Save(company).Error
}

func (r *AuthRepository) DeleteCompany(userID, profileID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&user.CompanyProfile{}, "id = ?", profileID).Error; err != nil {
			return err
		}
		return tx.Delete(&user.User{}, "id = ?", userID).Error
	})
}

func (r *AuthRepository) GetAlumni(offset, limit int, search string) ([]user.AlumniProfile, int64, error) {
	var alumni []user.AlumniProfile
	var total int64

	query := r.db.Preload("User").Model(&user.AlumniProfile{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Joins("JOIN users ON users.id = alumni_profiles.user_id").
			Where("LOWER(alumni_profiles.fullname) LIKE ? OR LOWER(users.email) LIKE ?",
				searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("alumni_profiles.created_at DESC").Find(&alumni).Error
	return alumni, total, err
}

func (r *AuthRepository) GetAlumniByID(id uuid.UUID) (*user.AlumniProfile, error) {
	var alumni user.AlumniProfile
	err := r.db.Preload("User").Where("id = ?", id).First(&alumni).Error
	return &alumni, err
}

func (r *AuthRepository) UpdateAlumniProfile(alumni *user.AlumniProfile) error {
	return r.db.Save(alumni).Error
}

func (r *AuthRepository) DeleteAlumni(userID, profileID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&user.AlumniProfile{}, "id = ?", profileID).Error; err != nil {
			return err
		}
		return tx.Delete(&user.User{}, "id = ?", userID).Error
	})
}

func (r *AuthRepository) CheckEmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&user.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *AuthRepository) GetLatestQuestionnaireResponseByStudentProfile(studentProfileID uuid.UUID) (*uuid.UUID, *string, error) {
	var response struct {
		RecommendedProfilingRoleID *string `gorm:"column:recommended_profiling_role_id"`
		RoleName                   *string `gorm:"column:name"`
	}

	err := r.db.Table("questionnaire_responses qr").
		Select("qr.recommended_profiling_role_id, tr.name").
		Joins("LEFT JOIN target_roles tr ON qr.recommended_profiling_role_id::uuid = tr.id").
		Where("qr.student_profile_id = ? AND qr.recommended_profiling_role_id IS NOT NULL", studentProfileID.String()).
		Order("qr.submitted_at DESC").
		First(&response).Error

	if err != nil {
		return nil, nil, err
	}

	var roleID *uuid.UUID
	if response.RecommendedProfilingRoleID != nil {
		if parsedID, err := uuid.Parse(*response.RecommendedProfilingRoleID); err == nil {
			roleID = &parsedID
		}
	}

	return roleID, response.RoleName, nil
}
