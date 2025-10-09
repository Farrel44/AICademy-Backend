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
		First(&u, "id = ?", id).Error
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}
	return &u, nil
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
