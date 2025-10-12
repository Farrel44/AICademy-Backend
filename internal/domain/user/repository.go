package user

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewUserRepository(db *gorm.DB, rdb *redis.Client) *UserRepository {
	return &UserRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*User, error) {
	var u User
	err := r.db.
		Preload("StudentProfile").
		Preload("AlumniProfile").
		First(&u, "id = ?", id).Error
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}
	return &u, nil
}

// Get user by NIS
func (r *UserRepository) GetUserByNIS(nis string) (*User, error) {
	var user User
	err := r.db.
		Preload("StudentProfile").
		Preload("AlumniProfile").
		Joins("JOIN student_profiles ON users.id = student_profiles.user_id").
		Where("student_profiles.nis = ?", nis).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Get student profile by NIS
func (r *UserRepository) GetStudentProfileByNIS(nis string) (*StudentProfile, error) {
	var studentProfile StudentProfile
	err := r.db.
		Preload("User").
		Where("nis = ?", nis).
		First(&studentProfile).Error

	if err != nil {
		return nil, err
	}
	return &studentProfile, nil
}

// Get projects by student profile ID for public view
func (r *UserRepository) GetUserProjectsByStudentProfileID(studentProfileID uuid.UUID) ([]UserProjectInfo, error) {
	var projects []UserProjectInfo

	err := r.db.Table("projects").
		Select(`
            projects.id,
            projects.project_name,
            projects.description,
            projects.link_url,
            projects.start_date,
            projects.end_date,
            projects.created_at,
            COALESCE(photo_counts.photo_count, 0) as photo_count,
            CASE 
                WHEN projects.end_date < NOW() THEN true
                ELSE false
            END as is_completed
        `).
		Joins(`LEFT JOIN (
            SELECT project_id, COUNT(*) as photo_count 
            FROM project_photos 
            GROUP BY project_id
        ) as photo_counts ON photo_counts.project_id = projects.id`).
		Where("projects.owner_student_profile_id = ?", studentProfileID).
		Order("projects.created_at DESC").
		Limit(10). // Show latest 10 projects in public profile
		Scan(&projects).Error

	return projects, err
}

// Get certifications by student profile ID for public view
func (r *UserRepository) GetUserCertificationsByStudentProfileID(studentProfileID uuid.UUID) ([]UserCertificationInfo, error) {
	var certifications []UserCertificationInfo

	err := r.db.Table("certifications").
		Select(`
            id,
            name,
            issuing_organization,
            issue_date,
            expiration_date,
            credential_id,
            credential_url,
            created_at,
            CASE 
                WHEN expiration_date IS NULL THEN false
                WHEN expiration_date < NOW() THEN true
                ELSE false
            END as is_expired,
            CASE 
                WHEN expiration_date IS NULL THEN false
                WHEN expiration_date < NOW() + INTERVAL '30 days' AND expiration_date >= NOW() THEN true
                ELSE false
            END as is_expiring_soon
        `).
		Where("student_profile_id = ?", studentProfileID).
		Order("created_at DESC").
		Limit(10). // Show latest 10 certifications in public profile
		Scan(&certifications).Error

	return certifications, err
}

// Calculate profile completeness
func (r *UserRepository) CalculateProfileCompleteness(profile *StudentProfile) int {
	completeness := 0
	total := 7 // Total fields to check

	// Basic info (always filled)
	completeness += 2 // fullname, nis

	// Optional fields
	if profile.ProfilePicture != "" {
		completeness++
	}
	if profile.Headline != "" {
		completeness++
	}
	if profile.Bio != "" {
		completeness++
	}
	if profile.CVFile != nil && *profile.CVFile != "" {
		completeness++
	}

	// Check if has projects/certifications
	var hasProjects, hasCertifications bool
	r.db.Table("projects").Where("owner_student_profile_id = ?", profile.ID).Limit(1).Scan(&hasProjects)
	r.db.Table("certifications").Where("student_profile_id = ?", profile.ID).Limit(1).Scan(&hasCertifications)

	if hasProjects {
		completeness++
	}
	if hasCertifications {
		total++ // Add extra point for certifications
		completeness++
	}

	return (completeness * 100) / total
}

func (r *UserRepository) GetStudentRecommendedRole(userID uuid.UUID) (*RecommendedRoleInfo, error) {
	var result struct {
		RoleID          *uuid.UUID `json:"role_id"`
		RoleName        *string    `json:"role_name"`
		RoleDescription *string    `json:"role_description"`
		RoleCategory    *string    `json:"role_category"`
		Score           *float64   `json:"score"`
		Justification   *string    `json:"justification"`
	}

	err := r.db.Table("questionnaire_responses").
		Select(`
            target_roles.id as role_id,
            target_roles.name as role_name,
            target_roles.description as role_description,
            target_roles.category as role_category,
            questionnaire_responses.total_score::float as score,
            questionnaire_responses.ai_analysis as justification
        `).
		Joins("LEFT JOIN target_roles ON target_roles.id::text = questionnaire_responses.recommended_profiling_role_id").
		Joins("LEFT JOIN student_profiles ON student_profiles.id::text = questionnaire_responses.student_profile_id").
		Where("student_profiles.user_id = ? AND questionnaire_responses.recommended_profiling_role_id IS NOT NULL AND questionnaire_responses.recommended_profiling_role_id != ''", userID).
		Order("questionnaire_responses.created_at DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if result.RoleID == nil {
		return nil, nil
	}

	return &RecommendedRoleInfo{
		RoleID:          *result.RoleID,
		RoleName:        *result.RoleName,
		RoleDescription: *result.RoleDescription,
		RoleCategory:    *result.RoleCategory,
		Score:           result.Score,
		Justification:   result.Justification,
	}, nil
}

func (r *UserRepository) UpdateStudentProfile(student *StudentProfile) error {
	return r.db.Save(student).Error
}
