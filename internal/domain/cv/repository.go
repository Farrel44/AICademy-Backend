package cv

import (
	"encoding/json"
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
		cacheTTL:     10 * time.Minute,
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

func (r *CVRepository) getCVByIDTx(tx *gorm.DB, id uuid.UUID) (*CV, error) {
	var cv CV
	if err := tx.First(&cv, "id = ?", id).Error; err != nil {
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
	return r.db.Transaction(func(tx *gorm.DB) error {
		cv, err := r.getCVByIDTx(tx, id)
		if err != nil {
			return err
		}

		if err := tx.Model(&CV{}).
			Where("student_profile_id = ? AND status = ? AND id != ?",
				cv.StudentProfileID, CVStatusPublished, id).
			Updates(map[string]interface{}{
				"status":    CVStatusDraft,
				"is_public": false,
			}).Error; err != nil {
			return err
		}

		return tx.Model(&CV{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":       CVStatusPublished,
			"is_public":    true,
			"published_at": "NOW()",
		}).Error
	})
}

func (r *CVRepository) UnpublishCV(id uuid.UUID) error {
	return r.db.Model(&CV{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    CVStatusDraft,
		"is_public": false,
	}).Error
}

func (r *CVRepository) GetPublicCVsByNIS(nis string) ([]CV, error) {
	var cvs []CV

	query := `
		SELECT c.*
		FROM cvs c
		JOIN student_profiles sp ON c.student_profile_id = sp.id
		WHERE sp.nis = ? AND c.is_public = true AND c.status = 'published'
		ORDER BY c.published_at DESC
	`

	err := r.db.Raw(query, nis).Scan(&cvs).Error
	return cvs, err
}

func (r *CVRepository) GetStudentProjects(userID uuid.UUID, limit int) ([]CVProject, error) {
	var projects []CVProject

	query := `
		SELECT DISTINCT ON (p.project_name)
			p.id,
			p.project_name as name,
			p.description,
			COALESCE(pc.project_role, 'Project Owner') as role,
			p.start_date,
			p.end_date,
			p.link_url as url,
			p.tech_stack
		FROM projects p
		LEFT JOIN project_contributors pc ON p.id = pc.project_id 
		JOIN student_profiles sp ON (p.owner_student_profile_id = sp.id OR pc.student_profile_id = sp.id)
		WHERE sp.user_id = ?
		ORDER BY p.project_name, p.start_date DESC
		LIMIT ?
	`

	err := r.db.Raw(query, userID, limit).Scan(&projects).Error
	return projects, err
}

func (r *CVRepository) GetStudentCertifications(userID uuid.UUID, limit int) ([]CVCertification, error) {
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
		LIMIT ?
	`

	err := r.db.Raw(query, userID, limit).Scan(&certifications).Error
	return certifications, err
}

func (r *CVRepository) GetStudentProfile(userID uuid.UUID) (*PersonalInfo, error) {
	var info PersonalInfo

	query := `
		SELECT 
			u.email,
			sp.fullname as full_name,
			COALESCE(sp.phone, '') as phone,
			COALESCE(sp.personal_email, '') as personal_email,
			COALESCE(sp.location, '') as location,
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

func (r *CVRepository) GetStudentExperiences(userID uuid.UUID) ([]CVExperience, error) {
	var experiences []CVExperience

	query := `
		SELECT 
			e.id,
			e.company_name,
			e.position,
			COALESCE(e.department, '') as department,
			e.employment_type,
			e.location,
			e.location_type,
			e.description,
			e.responsibilities,
			e.achievements,
			e.skills,
			e.start_date,
			e.end_date,
			e.is_current
		FROM experiences e
		JOIN student_profiles sp ON e.student_profile_id = sp.id
		WHERE sp.user_id = ?
		ORDER BY e.start_date DESC
	`

	err := r.db.Raw(query, userID).Scan(&experiences).Error
	return experiences, err
}

func (r *CVRepository) GetStudentLanguages(userID uuid.UUID) ([]CVLanguage, error) {
	var languagesJSON string

	query := `
		SELECT 
			COALESCE(sp.languages::text, '[]') as languages
		FROM student_profiles sp
		WHERE sp.user_id = ?
	`

	err := r.db.Raw(query, userID).Scan(&languagesJSON).Error
	if err != nil {
		return []CVLanguage{}, err
	}

	// Parse JSON string to Languages slice
	var userLanguages []struct {
		Name      string `json:"name"`
		Level     string `json:"level"`
		Certified bool   `json:"certified"`
	}

	if languagesJSON == "" || languagesJSON == "[]" {
		return []CVLanguage{}, nil
	}

	err = json.Unmarshal([]byte(languagesJSON), &userLanguages)
	if err != nil {
		return []CVLanguage{}, err
	}

	// Convert to CVLanguage format
	cvLanguages := make([]CVLanguage, len(userLanguages))
	for i, lang := range userLanguages {
		cvLanguages[i] = CVLanguage{
			Name:      lang.Name,
			Level:     lang.Level,
			Certified: lang.Certified,
		}
	}

	return cvLanguages, nil
}

func (r *CVRepository) GetStudentSkills(userID uuid.UUID, category, level string, limit int) ([]CVSkill, error) {
	var skills []CVSkill

	query := `
		SELECT DISTINCT 
			UNNEST(p.tech_stack) as name,
			? as category,
			? as level
		FROM projects p
		JOIN student_profiles sp ON p.owner_student_profile_id = sp.id
		WHERE sp.user_id = ?
		AND p.tech_stack IS NOT NULL
		LIMIT ?
	`

	err := r.db.Raw(query, category, level, userID, limit).Scan(&skills).Error
	return skills, err
}

func (r *CVRepository) GetStudentEducation(userID uuid.UUID) (*CVEducation, error) {
	var education CVEducation

	query := `
		SELECT 
			'SMK Telkom Purwokerto' as school,
			'Diploma' as degree,
			CASE 
				WHEN sp.class LIKE '%RPL%' THEN 'Rekayasa Perangkat Lunak'
				WHEN sp.class LIKE '%TKJ%' THEN 'Teknik Komputer dan Jaringan'
				ELSE 'Teknologi Informasi'
			END as major,
			EXTRACT(YEAR FROM CURRENT_DATE) - 3 as start_year,
			EXTRACT(YEAR FROM CURRENT_DATE) as end_year
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

func (r *CVRepository) UpdateStudentProfileCVFile(studentProfileID uuid.UUID, pdfPath string) error {
	return r.db.Table("student_profiles").
		Where("id = ?", studentProfileID).
		Update("cv_file", pdfPath).Error
}
