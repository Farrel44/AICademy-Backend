package cv

import (
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CVRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewCVRepository(db *gorm.DB, rdb *redis.Client) *CVRepository {
	return &CVRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

func (r *CVRepository) CreateCV(cv *CV) error {
	return r.db.Create(cv).Error
}

func (r *CVRepository) GetCVByID(id uuid.UUID) (*CV, error) {
	var cv CV
	if err := r.db.First(&cv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cv, nil
}

func (r *CVRepository) GetCVsByUserID(userID uuid.UUID) ([]CV, error) {
	var cvs []CV

	query := `
		SELECT c.*
		FROM cvs c
		JOIN student_profiles sp ON c.student_profile_id = sp.id
		WHERE sp.user_id = ?
		ORDER BY c.created_at DESC
	`

	err := r.db.Raw(query, userID).Scan(&cvs).Error
	return cvs, err
}

func (r *CVRepository) UpdateCV(cv *CV) error {
	return r.db.Save(cv).Error
}

func (r *CVRepository) DeleteCV(id uuid.UUID) error {
	return r.db.Delete(&CV{}, "id = ?", id).Error
}

func (r *CVRepository) PublishCV(id uuid.UUID) error {
	return r.db.Model(&CV{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       CVStatusPublished,
		"is_public":    true,
		"published_at": "NOW()",
	}).Error
}

func (r *CVRepository) UnpublishCV(id uuid.UUID) error {
	return r.db.Model(&CV{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    CVStatusDraft,
		"is_public": false,
	}).Error
}

func (r *CVRepository) GetPublicCVsByUserID(userID uuid.UUID) ([]CV, error) {
	var cvs []CV

	query := `
		SELECT c.*
		FROM cvs c
		JOIN student_profiles sp ON c.student_profile_id = sp.id
		WHERE sp.user_id = ? AND c.is_public = true
		ORDER BY c.published_at DESC
	`

	err := r.db.Raw(query, userID).Scan(&cvs).Error
	return cvs, err
}

func (r *CVRepository) GetStudentProjects(userID uuid.UUID) ([]CVProject, error) {
	var projects []CVProject

	query := `
		SELECT DISTINCT ON (p.project_name)
			p.id,
			p.project_name as name,
			p.description,
			COALESCE(pc.project_role, 'Project Owner') as role,
			p.start_date,
			p.end_date,
			p.link_url as url
		FROM projects p
		LEFT JOIN project_contributors pc ON p.id = pc.project_id 
		JOIN student_profiles sp ON (p.owner_student_profile_id = sp.id OR pc.student_profile_id = sp.id)
		WHERE sp.user_id = ?
		ORDER BY p.project_name, p.start_date DESC
		LIMIT 5
	`

	err := r.db.Raw(query, userID).Scan(&projects).Error
	return projects, err
}

func (r *CVRepository) GetStudentCertifications(userID uuid.UUID) ([]CVCertification, error) {
	var certifications []CVCertification

	query := `
		SELECT 
			c.id,
			c.name,
			c.issuing_organization,
			c.issue_date,
			c.expiration_date,
			c.credential_id,
			c.credential_url
		FROM certifications c
		JOIN student_profiles sp ON c.student_profile_id = sp.id
		WHERE sp.user_id = ?
		ORDER BY c.issue_date DESC
		LIMIT 10
	`

	err := r.db.Raw(query, userID).Scan(&certifications).Error
	return certifications, err
}

func (r *CVRepository) GetStudentProfile(userID uuid.UUID) (*PersonalInfo, error) {
	var info PersonalInfo

	query := `
		SELECT 
			u.email,
			sp.fullname as full_name,
			'' as phone,
			'' as location,
			'' as linkedin,
			'' as github,
			'' as portfolio
		FROM users u
		JOIN student_profiles sp ON u.id = sp.user_id
		WHERE u.id = ?
	`

	err := r.db.Raw(query, userID).Scan(&info).Error
	return &info, err
}

func (r *CVRepository) GetStudentSkills(userID uuid.UUID) ([]CVSkill, error) {
	var skills []CVSkill

	query := `
		SELECT DISTINCT 
			UNNEST(p.tech_stack) as name,
			'Technical' as category,
			'Intermediate' as level
		FROM projects p
		JOIN student_profiles sp ON p.owner_student_profile_id = sp.id
		WHERE sp.user_id = ?
		AND p.tech_stack IS NOT NULL
		LIMIT 20
	`

	err := r.db.Raw(query, userID).Scan(&skills).Error
	return skills, err
}

func (r *CVRepository) GetStudentEducation(userID uuid.UUID) (*CVEducation, error) {
	var education CVEducation

	query := `
		SELECT 
			'SMK Telkom Malang' as school,
			'Diploma' as degree,
			CASE 
				WHEN sp.class LIKE '%RPL%' THEN 'Rekayasa Perangkat Lunak'
				WHEN sp.class LIKE '%TKJ%' THEN 'Teknik Komputer dan Jaringan'
				ELSE 'Teknologi Informasi'
			END as major,
			2022 as start_year,
			2025 as end_year
		FROM student_profiles sp
		WHERE sp.user_id = ?
	`

	err := r.db.Raw(query, userID).Scan(&education).Error
	return &education, err
}

func (r *CVRepository) GetStudentProfileID(userID uuid.UUID) (uuid.UUID, error) {
	var studentProfile struct {
		ID uuid.UUID `json:"id"`
	}

	err := r.db.Table("student_profiles").
		Select("id").
		Where("user_id = ?", userID).
		First(&studentProfile).Error

	return studentProfile.ID, err
}
