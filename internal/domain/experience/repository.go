package experience

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateExperience(experience *Experience) error
	GetExperienceByID(id uuid.UUID) (*Experience, error)
	GetExperiencesByStudentID(studentProfileID uuid.UUID, page, limit int) ([]Experience, int64, error)
	UpdateExperience(experience *Experience) error
	DeleteExperience(id uuid.UUID) error
	ExistsForStudent(id, studentProfileID uuid.UUID) (bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateExperience(experience *Experience) error {
	return r.db.Create(experience).Error
}

func (r *repository) GetExperienceByID(id uuid.UUID) (*Experience, error) {
	var experience Experience
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&experience).Error
	if err != nil {
		return nil, err
	}
	return &experience, nil
}

func (r *repository) GetExperiencesByStudentID(studentProfileID uuid.UUID, page, limit int) ([]Experience, int64, error) {
	var experiences []Experience
	var total int64

	// Count total records
	err := r.db.Model(&Experience{}).
		Where("student_profile_id = ? AND deleted_at IS NULL", studentProfileID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated records ordered by start_date descending (most recent first)
	offset := (page - 1) * limit
	err = r.db.Where("student_profile_id = ? AND deleted_at IS NULL", studentProfileID).
		Order("start_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&experiences).Error

	if err != nil {
		return nil, 0, err
	}

	return experiences, total, nil
}

func (r *repository) UpdateExperience(experience *Experience) error {
	return r.db.Save(experience).Error
}

func (r *repository) DeleteExperience(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&Experience{}).Error
}

func (r *repository) ExistsForStudent(id, studentProfileID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&Experience{}).
		Where("id = ? AND student_profile_id = ? AND deleted_at IS NULL", id, studentProfileID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check experience existence: %w", err)
	}

	return count > 0, nil
}
