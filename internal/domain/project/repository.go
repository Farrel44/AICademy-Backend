package project

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewProjectRepository(db *gorm.DB, rdb *redis.Client) *ProjectRepository {
	return &ProjectRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

// Project methods
func (r *ProjectRepository) CreateProject(project *Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) GetProjectByID(id uuid.UUID) (*Project, error) {
	var project Project
	err := r.db.Preload("Contributors").Preload("Photos").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetProjectsByOwnerID(ownerID uuid.UUID) ([]Project, error) {
	var projects []Project
	err := r.db.Preload("Contributors").Preload("Photos").Where("owner_student_profile_id = ?", ownerID).Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) UpdateProject(project *Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) DeleteProject(id uuid.UUID) error {
	return r.db.Delete(&Project{}, "id = ?", id).Error
}

func (r *ProjectRepository) AddProjectPhoto(photo *ProjectPhoto) error {
	return r.db.Create(photo).Error
}

func (r *ProjectRepository) DeleteProjectPhoto(id uuid.UUID) error {
	return r.db.Delete(&ProjectPhoto{}, "id = ?", id).Error
}

func (r *ProjectRepository) AddProjectContributor(contributor *ProjectContributor) error {
	return r.db.Create(contributor).Error
}

func (r *ProjectRepository) RemoveProjectContributor(projectID, studentProfileID uuid.UUID) error {
	return r.db.Delete(&ProjectContributor{}, "project_id = ? AND student_profile_id = ?", projectID, studentProfileID).Error
}

// Certification methods
func (r *ProjectRepository) CreateCertification(certification *Certification) error {
	return r.db.Create(certification).Error
}

func (r *ProjectRepository) GetCertificationByID(id uuid.UUID) (*Certification, error) {
	var certification Certification
	err := r.db.Preload("Photos").First(&certification, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &certification, nil
}

func (r *ProjectRepository) GetCertificationsByStudentID(studentID uuid.UUID) ([]Certification, error) {
	var certifications []Certification
	err := r.db.Preload("Photos").Where("student_profile_id = ?", studentID).Find(&certifications).Error
	return certifications, err
}

func (r *ProjectRepository) UpdateCertification(certification *Certification) error {
	return r.db.Save(certification).Error
}

func (r *ProjectRepository) DeleteCertification(id uuid.UUID) error {
	return r.db.Delete(&Certification{}, "id = ?", id).Error
}

func (r *ProjectRepository) AddCertificationPhoto(photo *CertificationPhoto) error {
	return r.db.Create(photo).Error
}

func (r *ProjectRepository) DeleteCertificationPhoto(id uuid.UUID) error {
	return r.db.Delete(&CertificationPhoto{}, "id = ?", id).Error
}

func (r *ProjectRepository) GetStudentProfileIDByUserID(userID uuid.UUID) (uuid.UUID, error) {
	var studentProfileID uuid.UUID
	err := r.db.Table("users").
		Select("student_profiles.id").
		Joins("JOIN student_profiles ON users.id = student_profiles.user_id").
		Where("users.id = ?", userID).
		Scan(&studentProfileID).Error
	return studentProfileID, err
}
