package pkl

import (
	"fmt"
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

// Optimized methods for search performance
func (r *PklRepository) CountInternships(search string) (int64, error) {
	var total int64
	query := r.db.Model(&Internship{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(position) LIKE ? OR LOWER(location) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	err := query.Count(&total).Error
	return total, err
}

func (r *PklRepository) GetInternshipsOptimized(offset, limit int, search string) ([]Internship, error) {
	var internships []Internship
	query := r.db.Select("internships.*, company_profiles.company_name, users.name as company_user_name").
		Joins("LEFT JOIN company_profiles ON internships.company_profile_id = company_profiles.id").
		Joins("LEFT JOIN users ON company_profiles.user_id = users.id").
		Model(&Internship{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(internships.position) LIKE ? OR LOWER(internships.location) LIKE ? OR LOWER(internships.description) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	err := query.Offset(offset).Limit(limit).Order("internships.posted_at DESC").Find(&internships).Error
	return internships, err
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

func (r *PklRepository) ApplyInternshipByID(internshipPosition *InternshipApplication) error {
	return r.db.Save(internshipPosition).Error
}

func (r *PklRepository) GetUserByID(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.
		Preload("StudentProfile").
		First(&u, "id = ?", id).Error
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}
	return &u, nil
}

func (r *PklRepository) HasExistingApplication(internshipID, studentProfileID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&InternshipApplication{}).
		Where("internship_id = ? AND student_profile_id = ?", internshipID, studentProfileID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PklRepository) GetSubmissionsByInternshipID(internshipID uuid.UUID) ([]InternshipApplication, error) {
	var submissions []InternshipApplication
	err := r.db.Preload("StudentProfile").
		Preload("StudentProfile.User").
		Preload("AlumniProfile").
		Preload("AlumniProfile.User").
		Preload("Internship").
		Preload("Internship.CompanyProfile").
		Preload("Internship.CompanyProfile.User").
		Preload("ApprovedByUser").
		Where("internship_id = ?", internshipID).
		Order("applied_at DESC").
		Find(&submissions).Error
	return submissions, err
}

func (r *PklRepository) GetInternshipsWithSubmissionsByCompanyID(companyID uuid.UUID) ([]Internship, error) {
	var internships []Internship
	err := r.db.Preload("CompanyProfile").
		Preload("CompanyProfile.User").
		Preload("InternshipApplications").
		Preload("InternshipApplications.StudentProfile").
		Preload("InternshipApplications.StudentProfile.User").
		Preload("InternshipApplications.AlumniProfile").
		Preload("InternshipApplications.AlumniProfile.User").
		Preload("InternshipApplications.ApprovedByUser").
		Where("company_profile_id = ?", companyID).
		Order("posted_at DESC").
		Find(&internships).Error
	return internships, err
}

func (r *PklRepository) GetSubmissionByID(submissionID uuid.UUID) (*InternshipApplication, error) {
	var submission InternshipApplication
	err := r.db.Preload("StudentProfile").
		Preload("StudentProfile.User").
		Preload("AlumniProfile").
		Preload("AlumniProfile.User").
		Preload("Internship").
		Preload("Internship.CompanyProfile").
		Preload("Internship.CompanyProfile.User").
		Preload("ApprovedByUser").
		First(&submission, "id = ?", submissionID).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *PklRepository) UpdateSubmissionStatus(submissionID uuid.UUID, status ApplicationStatus, approvedBy *uuid.UUID, role string) error {
	updates := map[string]interface{}{
		"status":      status,
		"reviewed_at": time.Now(),
		"updated_at":  time.Now(),
	}

	if approvedBy != nil {
		updates["approved_by_user_id"] = *approvedBy
		updates["approved_by_role"] = strings.ToLower(role)
	}

	return r.db.Model(&InternshipApplication{}).
		Where("id = ?", submissionID).
		Updates(updates).Error
}

func (r *PklRepository) GetUserSubmissionByInternshipID(internshipID uuid.UUID) ([]InternshipApplication, error) {
	var submissions []InternshipApplication
	err := r.db.Preload("StudentProfile").
		Preload("StudentProfile.User").
		Preload("AlumniProfile").
		Preload("AlumniProfile.User").
		Where("internship_id = ?", internshipID).
		Find(&submissions).Error
	return submissions, err
}

// Alumni-specific methods
func (r *PklRepository) GetAlumniUserByID(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.
		Preload("AlumniProfile").
		First(&u, "id = ?", id).Error
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}
	return &u, nil
}

func (r *PklRepository) HasExistingAlumniApplication(internshipID, alumniProfileID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&InternshipApplication{}).
		Where("internship_id = ? AND alumni_profile_id = ?", internshipID, alumniProfileID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PklRepository) GetAlumniInternships(offset, limit int, search string) ([]Internship, int64, error) {
	var internships []Internship
	var total int64

	query := r.db.Preload("CompanyProfile").
		Preload("CompanyProfile.User").
		Model(&Internship{}).
		Where("type IN (?)", []InternshipType{InternshipTypeJob, InternshipTypeFreelance})

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

func (r *PklRepository) GetAlumniApplicationsByProfileID(alumniProfileID uuid.UUID, offset, limit int) ([]InternshipApplication, int64, error) {
	var applications []InternshipApplication
	var total int64

	query := r.db.Preload("Internship").
		Preload("Internship.CompanyProfile").
		Preload("Internship.CompanyProfile.User").
		Preload("AlumniProfile").
		Preload("AlumniProfile.User").
		Preload("ApprovedByUser").
		Model(&InternshipApplication{}).
		Where("alumni_profile_id = ?", alumniProfileID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Order("applied_at DESC").Find(&applications).Error
	return applications, total, err
}

// Company-specific methods
func (r *PklRepository) GetCompanyUserByID(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.
		Preload("CompanyProfile").
		First(&u, "id = ?", id).Error
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}
	return &u, nil
}

func (r *PklRepository) GetInternshipsByCompanyID(companyProfileID uuid.UUID, offset, limit int, search string) ([]Internship, int64, error) {
	var internships []Internship
	var total int64

	query := r.db.Preload("CompanyProfile").
		Preload("CompanyProfile.User").
		Model(&Internship{}).
		Where("company_profile_id = ?", companyProfileID)

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
