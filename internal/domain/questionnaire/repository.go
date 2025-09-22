package questionnaire

import (
	"aicademy-backend/internal/domain/user"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuestionnaireRepository struct {
	db *gorm.DB
}

func NewQuestionnaireRepository(db *gorm.DB) *QuestionnaireRepository {
	return &QuestionnaireRepository{db: db}
}

func (r *QuestionnaireRepository) GetActiveQuestionnaire() (*ProfilingQuestionnaire, error) {
	var questionnaire ProfilingQuestionnaire
	err := r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("question_order ASC")
	}).Where("active = ?", true).First(&questionnaire).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tidak ada kuesioner aktif")
		}
		return nil, err
	}

	return &questionnaire, nil
}

func (r *QuestionnaireRepository) GetQuestionnaireByID(id uuid.UUID) (*ProfilingQuestionnaire, error) {
	var questionnaire ProfilingQuestionnaire
	err := r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("question_order ASC")
	}).Where("id = ?", id).First(&questionnaire).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("kuesioner tidak ditemukan")
		}
		return nil, err
	}

	return &questionnaire, nil
}

func (r *QuestionnaireRepository) GetAllQuestionnaires(page, limit int) ([]ProfilingQuestionnaire, int64, error) {
	var questionnaires []ProfilingQuestionnaire
	var total int64

	err := r.db.Model(&ProfilingQuestionnaire{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("question_order ASC")
	}).Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&questionnaires).Error

	return questionnaires, total, err
}

func (r *QuestionnaireRepository) CreateQuestionnaire(questionnaire *ProfilingQuestionnaire) error {
	return r.db.Create(questionnaire).Error
}

func (r *QuestionnaireRepository) UpdateQuestionnaire(questionnaire *ProfilingQuestionnaire) error {
	return r.db.Save(questionnaire).Error
}

func (r *QuestionnaireRepository) DeleteQuestionnaire(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("questionnaire_id = ?", id).Delete(&QuestionnaireQuestion{}).Error; err != nil {
			return err
		}
		return tx.Delete(&ProfilingQuestionnaire{}, id).Error
	})
}

func (r *QuestionnaireRepository) ActivateQuestionnaire(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ProfilingQuestionnaire{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}
		return tx.Model(&ProfilingQuestionnaire{}).Where("id = ?", id).Update("active", true).Error
	})
}

func (r *QuestionnaireRepository) DeactivateAllQuestionnaires() error {
	return r.db.Model(&ProfilingQuestionnaire{}).Where("active = ?", true).Update("active", false).Error
}

func (r *QuestionnaireRepository) CreateQuestionnaireResponse(response *QuestionnaireResponse) error {
	return r.db.Create(response).Error
}

func (r *QuestionnaireRepository) UpdateResponse(response *QuestionnaireResponse) error {
	return r.db.Save(response).Error
}

func (r *QuestionnaireRepository) GetResponseByID(id uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Preload("Questionnaire").Where("id = ?", id).First(&response).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("respons tidak ditemukan")
		}
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) GetStudentProfileByUserID(userID uuid.UUID) (*user.StudentProfile, error) {
	var profile user.StudentProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("profil siswa tidak ditemukan")
		}
		return nil, err
	}
	return &profile, nil
}

func (r *QuestionnaireRepository) GetLatestResponseByStudentProfile(studentProfileID uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Preload("Questionnaire").
		Where("student_profile_id = ?", studentProfileID).
		Order("submitted_at DESC").
		First(&response).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("belum ada respons")
		}
		return nil, err
	}

	return &response, nil
}

func (r *QuestionnaireRepository) GetResponseByStudentAndQuestionnaire(studentProfileID, questionnaireID uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Where("student_profile_id = ? AND questionnaire_id = ?", studentProfileID, questionnaireID).
		First(&response).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &response, nil
}

func (r *QuestionnaireRepository) AddQuestionsToQuestionnaire(questionnaireID uuid.UUID, questions []QuestionnaireQuestion) error {
	for i := range questions {
		questions[i].QuestionnaireID = questionnaireID
	}
	return r.db.Create(&questions).Error
}

func (r *QuestionnaireRepository) GetQuestionsByQuestionnaireID(questionnaireID uuid.UUID) ([]QuestionnaireQuestion, error) {
	var questions []QuestionnaireQuestion
	err := r.db.Where("questionnaire_id = ?", questionnaireID).Order("question_order ASC").Find(&questions).Error
	return questions, err
}

func (r *QuestionnaireRepository) GetResponsesByQuestionnaireID(questionnaireID uuid.UUID, page, limit int) ([]QuestionnaireResponse, int64, error) {
	var responses []QuestionnaireResponse
	var total int64

	err := r.db.Model(&QuestionnaireResponse{}).Where("questionnaire_id = ?", questionnaireID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Preload("Questionnaire").
		Where("questionnaire_id = ?", questionnaireID).
		Order("submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&responses).Error

	return responses, total, err
}

func (r *QuestionnaireRepository) GetQuestionnaireResponses(questionnaireID uuid.UUID, page, limit int) ([]QuestionnaireResponseSummary, int64, error) {
	var responses []QuestionnaireResponseSummary
	var total int64

	query := `
        SELECT 
            qr.id,
            COALESCE(u.name, 'Unknown') as student_name,
            COALESCE(u.email, 'unknown@email.com') as student_email,
            qr.submitted_at,
            qr.processed_at,
            qr.total_score,
            CASE 
                WHEN qr.processed_at IS NOT NULL THEN 'completed'
                ELSE 'processing'
            END as status
        FROM questionnaire_responses qr
        LEFT JOIN student_profiles sp ON qr.student_profile_id = sp.id
        LEFT JOIN users u ON sp.user_id = u.id
        WHERE qr.questionnaire_id = ?
    `

	countQuery := `
        SELECT COUNT(*)
        FROM questionnaire_responses qr
        WHERE qr.questionnaire_id = ?
    `

	err := r.db.Raw(countQuery, questionnaireID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	finalQuery := query + ` ORDER BY qr.submitted_at DESC LIMIT ? OFFSET ?`

	err = r.db.Raw(finalQuery, questionnaireID, limit, offset).Scan(&responses).Error
	if err != nil {
		return nil, 0, err
	}

	for i := range responses {
		if responses[i].ProcessedAt != nil {
			var recommendedRole string
			err := r.db.Raw(`
                SELECT ai_recommendations 
                FROM questionnaire_responses 
                WHERE id = ?
            `, responses[i].ID).Scan(&recommendedRole).Error

			if err == nil && recommendedRole != "" {
				var recommendations []AIRecommendation
				if json.Unmarshal([]byte(recommendedRole), &recommendations) == nil && len(recommendations) > 0 {
					responses[i].RecommendedRole = &recommendations[0].RoleName
				}
			}
		}
	}

	return responses, total, nil
}

func (r *QuestionnaireRepository) GetAllRoles() ([]RoleRecommendation, error) {
	var roles []RoleRecommendation
	err := r.db.Where("active = ?", true).Order("role_name ASC").Find(&roles).Error
	return roles, err
}

func (r *QuestionnaireRepository) GetRoleByName(roleName string) (*RoleRecommendation, error) {
	var role RoleRecommendation
	err := r.db.Where("role_name = ? AND active = ?", roleName, true).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role tidak ditemukan: %s", roleName)
		}
		return nil, err
	}
	return &role, nil
}

func (r *QuestionnaireRepository) GetRoleByID(id uuid.UUID) (*RoleRecommendation, error) {
	var role RoleRecommendation
	err := r.db.Where("id = ? AND active = ?", id, true).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role tidak ditemukan")
		}
		return nil, err
	}
	return &role, nil
}

func (r *QuestionnaireRepository) CreateRole(role *RoleRecommendation) error {
	var existingRole RoleRecommendation
	err := r.db.Where("role_name = ?", role.RoleName).First(&existingRole).Error
	if err == nil {
		return fmt.Errorf("role dengan nama '%s' sudah ada", role.RoleName)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.Create(role).Error
}

func (r *QuestionnaireRepository) UpdateRole(role *RoleRecommendation) error {
	return r.db.Save(role).Error
}

func (r *QuestionnaireRepository) DeleteRole(id uuid.UUID) error {
	return r.db.Model(&RoleRecommendation{}).Where("id = ?", id).Update("active", false).Error
}

func (r *QuestionnaireRepository) GetRolesByNames(roleNames []string) ([]RoleRecommendation, error) {
	var roles []RoleRecommendation
	err := r.db.Where("role_name IN ? AND active = ?", roleNames, true).Find(&roles).Error
	return roles, err
}
