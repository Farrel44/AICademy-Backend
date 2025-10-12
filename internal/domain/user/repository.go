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

// Get certifications by student profile ID for public view
func (r *UserRepository) GetUserCertificationsByStudentProfileID(studentProfileID uuid.UUID) ([]UserCertificationInfo, error) {
	var certifications []struct {
		ID                  uuid.UUID  `json:"id"`
		StudentProfileID    uuid.UUID  `json:"student_profile_id"`
		Name                string     `json:"name"`
		IssuingOrganization string     `json:"issuing_organization"`
		IssueDate           time.Time  `json:"issue_date"`
		ExpirationDate      *time.Time `json:"expiration_date"`
		CredentialID        *string    `json:"credential_id"`
		CredentialURL       *string    `json:"credential_url"`
		CreatedAt           time.Time  `json:"created_at"`
	}

	// Get the certifications without photos first to avoid N+1
	err := r.db.Table("certifications").
		Where("student_profile_id = ?", studentProfileID).
		Order("created_at DESC").
		Limit(10).
		Find(&certifications).Error

	if err != nil {
		return nil, err
	}

	// Get certification IDs for photo query
	var certificationIDs []uuid.UUID
	for _, cert := range certifications {
		certificationIDs = append(certificationIDs, cert.ID)
	}

	// Get all photos for these certifications in one query
	var photos []struct {
		ID              uuid.UUID `json:"id"`
		CertificationID uuid.UUID `json:"certification_id"`
		URL             string    `json:"url"`
		Caption         *string   `json:"caption"`
		IsPrimary       bool      `json:"is_primary"`
		CreatedAt       time.Time `json:"created_at"`
	}

	if len(certificationIDs) > 0 {
		err = r.db.Table("certification_photos").
			Where("certification_id IN ?", certificationIDs).
			Find(&photos).Error
		if err != nil {
			return nil, err
		}
	}

	// Group photos by certification ID
	photoMap := make(map[uuid.UUID][]CertificationPhoto)
	for _, photo := range photos {
		photoMap[photo.CertificationID] = append(photoMap[photo.CertificationID], CertificationPhoto{
			ID:              photo.ID,
			CertificationID: photo.CertificationID,
			URL:             photo.URL,
			Caption:         photo.Caption,
			IsPrimary:       photo.IsPrimary,
			CreatedAt:       photo.CreatedAt,
		})
	}

	// Convert to UserCertificationInfo
	var certificationInfos []UserCertificationInfo
	for _, cert := range certifications {
		// Check if expired
		isExpired := false
		isExpiringSoon := false
		if cert.ExpirationDate != nil {
			isExpired = time.Now().After(*cert.ExpirationDate)
			if !isExpired {
				thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
				isExpiringSoon = cert.ExpirationDate.Before(thirtyDaysFromNow)
			}
		}

		certInfo := UserCertificationInfo{
			ID:                  cert.ID,
			Name:                cert.Name,
			IssuingOrganization: cert.IssuingOrganization,
			IssueDate:           cert.IssueDate,
			ExpirationDate:      cert.ExpirationDate,
			CredentialID:        cert.CredentialID,
			CredentialURL:       cert.CredentialURL,
			IsExpired:           isExpired,
			IsExpiringSoon:      isExpiringSoon,
			CreatedAt:           cert.CreatedAt,
			Photos:              photoMap[cert.ID],
		}
		certificationInfos = append(certificationInfos, certInfo)
	}

	return certificationInfos, nil
}

// Get projects by student profile ID for public view
func (r *UserRepository) GetUserProjectsByStudentProfileID(studentProfileID uuid.UUID) ([]UserProjectInfo, error) {
	var projects []struct {
		ID          uuid.UUID `json:"id"`
		ProjectName string    `json:"project_name"`
		Description string    `json:"description"`
		LinkURL     *string   `json:"link_url"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
		CreatedAt   time.Time `json:"created_at"`
	}

	// Get the projects without photos first to avoid N+1
	err := r.db.Table("projects").
		Where("owner_student_profile_id = ?", studentProfileID).
		Order("created_at DESC").
		Limit(10).
		Find(&projects).Error

	if err != nil {
		return nil, err
	}

	// Get project IDs for photo query
	var projectIDs []uuid.UUID
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	// Get all photos for these projects in one query
	var photos []struct {
		ID        uuid.UUID `json:"id"`
		ProjectID uuid.UUID `json:"project_id"`
		URL       string    `json:"url"`
		Caption   *string   `json:"caption"`
		IsPrimary bool      `json:"is_primary"`
		CreatedAt time.Time `json:"created_at"`
	}

	if len(projectIDs) > 0 {
		err = r.db.Table("project_photos").
			Where("project_id IN ?", projectIDs).
			Find(&photos).Error
		if err != nil {
			return nil, err
		}
	}

	// Group photos by project ID
	photoMap := make(map[uuid.UUID][]ProjectPhoto)
	for _, photo := range photos {
		photoMap[photo.ProjectID] = append(photoMap[photo.ProjectID], ProjectPhoto{
			ID:        photo.ID,
			ProjectID: photo.ProjectID,
			URL:       photo.URL,
			Caption:   photo.Caption,
			IsPrimary: photo.IsPrimary,
			CreatedAt: photo.CreatedAt,
		})
	}

	// Convert to UserProjectInfo
	var projectInfos []UserProjectInfo
	for _, project := range projects {
		projectInfo := UserProjectInfo{
			ID:          project.ID,
			ProjectName: project.ProjectName,
			Description: project.Description,
			LinkURL:     project.LinkURL,
			StartDate:   project.StartDate,
			EndDate:     project.EndDate,
			CreatedAt:   project.CreatedAt,
			PhotoCount:  len(photoMap[project.ID]),
			IsCompleted: project.EndDate.Before(time.Now()),
			Photos:      photoMap[project.ID],
		}
		projectInfos = append(projectInfos, projectInfo)
	}

	return projectInfos, nil
}

func (r *UserRepository) GetStudentProfileByNIS(nis string) (*StudentProfile, error) {
	var studentProfile StudentProfile
	err := r.db.Where("nis = ?", nis).First(&studentProfile).Error
	if err != nil {
		return nil, err
	}
	return &studentProfile, nil
}

// Also add this function that might be needed by project service
func (r *UserRepository) GetStudentProfileByEmail(email string) (*StudentProfile, error) {
	var studentProfile StudentProfile
	err := r.db.
		Joins("JOIN users ON users.id = student_profiles.user_id").
		Where("users.email = ?", email).
		First(&studentProfile).Error
	if err != nil {
		return nil, err
	}
	return &studentProfile, nil
}
