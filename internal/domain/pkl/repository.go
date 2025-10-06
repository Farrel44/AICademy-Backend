package pkl

import (
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PklRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewPklRepository(db *gorm.DB, rdb *redis.Client) *PklRepository {
	return &PklRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

func (r *PklRepository) CreateInternshipPosition(internship *Internship) error {
	return r.db.Create(internship).Error
}

func (r *PklRepository) GetInternshipByID(id uuid.UUID) (*Internship, error) {
	var internship Internship
	err := r.db.Preload("CompanyProfile").
		Preload("CompanyProfile.User").
		First(&internship, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &internship, nil
}

func (r *PklRepository) GetAllInternships(offset, limit int, search string) ([]Internship, int64, error) {
	var internships []Internship
	var total int64

	query := r.db.Preload("CompanyProfile").
		Preload("CompanyProfile.User").
		Model(&Internship{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("posted_at DESC").Find(&internships).Error
	return internships, total, err
}

func (r *PklRepository) UpdateInternshipPosition(internship *Internship) error {
	return r.db.Save(internship).Error
}

func (r *PklRepository) DeleteInternshipPosition(id uuid.UUID) error {
	return r.db.Delete(&Internship{}, "id = ?", id).Error
}

func (r *PklRepository) GetCompanyByID(id uuid.UUID) (*user.CompanyProfile, error) {
	var company user.CompanyProfile
	err := r.db.Preload("User").Where("id = ?", id).First(&company).Error
	return &company, err
}
