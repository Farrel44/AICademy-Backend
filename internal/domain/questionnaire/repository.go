package questionnaire

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"

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

func (r *QuestionnaireRepository) UpdateQuestionnaireGenerationStatus(id uuid.UUID, status string, progress int, message string) error {
	updates := map[string]interface{}{
		"generation_status":     status,
		"generation_progress":   progress,
		"generation_message":    message,
		"generation_updated_at": time.Now(),
	}
	return r.db.Model(&ProfilingQuestionnaire{}).Where("id = ?", id).Updates(updates).Error
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

// New methods for restructured questionnaire system

// Target Role methods
func (r *QuestionnaireRepository) CreateTargetRole(role *TargetRole) error {
	return r.db.Create(role).Error
}

func (r *QuestionnaireRepository) GetTargetRoles(offset, limit int) ([]TargetRole, int64, error) {
	log.Printf("*** REPOSITORY: GetTargetRoles called with offset=%d, limit=%d ***", offset, limit)

	var roles []TargetRole
	var total int64

	err := r.db.Model(&TargetRole{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	log.Printf("DEBUG: TargetRoles count: %d", total)

	err = r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&roles).Error

	log.Printf("DEBUG: Found %d target roles", len(roles))

	return roles, total, err
}

func (r *QuestionnaireRepository) GetTargetRoleByID(id uuid.UUID) (*TargetRole, error) {
	var role TargetRole
	err := r.db.Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target role tidak ditemukan")
		}
		return nil, err
	}
	return &role, nil
}

func (r *QuestionnaireRepository) GetTargetRoleByName(name string) (*TargetRole, error) {
	var role TargetRole
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("target role tidak ditemukan")
		}
		return nil, err
	}
	return &role, nil
}

func (r *QuestionnaireRepository) UpdateTargetRole(role *TargetRole) error {
	return r.db.Save(role).Error
}

func (r *QuestionnaireRepository) DeleteTargetRole(id uuid.UUID) error {
	return r.db.Delete(&TargetRole{}, id).Error
}

// Questionnaire methods for new structure
func (r *QuestionnaireRepository) CreateQuestionnaireNew(questionnaire *ProfilingQuestionnaire) error {
	return r.db.Create(questionnaire).Error
}

func (r *QuestionnaireRepository) GetQuestionnairesNew(offset, limit int) ([]ProfilingQuestionnaire, int64, error) {
	var questionnaires []ProfilingQuestionnaire
	var total int64

	err := r.db.Model(&ProfilingQuestionnaire{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&questionnaires).Error
	return questionnaires, total, err
}

func (r *QuestionnaireRepository) LinkQuestionnaireTargetRole(questionnaireID, targetRoleID uuid.UUID) error {
	// Create junction table entry
	type QuestionnaireTargetRole struct {
		QuestionnaireID uuid.UUID `gorm:"primaryKey"`
		TargetRoleID    uuid.UUID `gorm:"primaryKey"`
	}

	link := QuestionnaireTargetRole{
		QuestionnaireID: questionnaireID,
		TargetRoleID:    targetRoleID,
	}

	return r.db.Create(&link).Error
}

func (r *QuestionnaireRepository) GetTargetRolesByQuestionnaireID(questionnaireID uuid.UUID) ([]TargetRole, error) {
	var roles []TargetRole

	err := r.db.Table("target_roles").
		Joins("JOIN questionnaire_target_roles ON target_roles.id = questionnaire_target_roles.target_role_id").
		Where("questionnaire_target_roles.questionnaire_id = ?", questionnaireID).
		Find(&roles).Error

	return roles, err
}

// Question methods for new structure
func (r *QuestionnaireRepository) CreateQuestionnaireQuestionNew(question *QuestionnaireQuestion) error {
	return r.db.Create(question).Error
}

func (r *QuestionnaireRepository) GetQuestionsByQuestionnaireIDNew(questionnaireID uuid.UUID) ([]QuestionnaireQuestion, error) {
	var questions []QuestionnaireQuestion
	err := r.db.Where("questionnaire_id = ?", questionnaireID).
		Order("question_order ASC").
		Find(&questions).Error
	return questions, err
}

// Questionnaire activation for new structure
func (r *QuestionnaireRepository) DeactivateAllQuestionnairesNew() error {
	return r.db.Model(&ProfilingQuestionnaire{}).Where("active = ?", true).Update("active", false).Error
}

func (r *QuestionnaireRepository) SetQuestionnaireActiveNew(id uuid.UUID, active bool) error {
	return r.db.Model(&ProfilingQuestionnaire{}).Where("id = ?", id).Update("active", active).Error
}

func (r *QuestionnaireRepository) GetQuestionnaireSubmissionCountNew(questionnaireID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&QuestionnaireResponse{}).Where("questionnaire_id = ?", questionnaireID).Count(&count).Error
	return int(count), err
}

// Response methods for new structure
func (r *QuestionnaireRepository) CreateQuestionnaireResponseNew(response *QuestionnaireResponse) error {
	return r.db.Create(response).Error
}

func (r *QuestionnaireRepository) GetQuestionnaireResponseByIDNew(id uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Where("id = ?", id).First(&response).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("response tidak ditemukan")
		}
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) GetResponseByStudentAndQuestionnaireNew(studentProfileID, questionnaireID uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Where("student_profile_id = ? AND questionnaire_id = ?", studentProfileID, questionnaireID).First(&response).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No existing response found
		}
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) GetLatestResponseByStudentIDNew(studentProfileID uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Where("student_profile_id = ?", studentProfileID).
		Order("submitted_at DESC").
		First(&response).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tidak ada hasil kuesioner")
		}
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) UpdateResponseStatusNew(responseID uuid.UUID, status, errorMessage string) error {
	updates := map[string]interface{}{
		"processing_status": status,
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	return r.db.Model(&QuestionnaireResponse{}).Where("id = ?", responseID).Updates(updates).Error
}

func (r *QuestionnaireRepository) UpdateResponseWithAIResultsNew(responseID uuid.UUID, analysisJSON, status string) error {
	return r.db.Model(&QuestionnaireResponse{}).Where("id = ?", responseID).Updates(map[string]interface{}{
		"analysis_json":     analysisJSON,
		"processing_status": status,
	}).Error
}

func (r *QuestionnaireRepository) GetQuestionnaireResponsesNew(offset, limit int, questionnaireID *uuid.UUID) ([]QuestionnaireResponse, int64, error) {
	var responses []QuestionnaireResponse
	var total int64

	query := r.db.Model(&QuestionnaireResponse{})
	if questionnaireID != nil {
		query = query.Where("questionnaire_id = ?", *questionnaireID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("submitted_at DESC").Offset(offset).Limit(limit).Find(&responses).Error
	return responses, total, err
}

// Recommendation methods for new structure
func (r *QuestionnaireRepository) CreateRoleRecommendationNew(recommendation *RoleRecommendation) error {
	return r.db.Create(recommendation).Error
}

func (r *QuestionnaireRepository) GetRecommendationsByResponseIDNew(responseID uuid.UUID) ([]RoleRecommendation, error) {
	var recommendations []RoleRecommendation
	err := r.db.Where("response_id = ?", responseID).
		Order("score DESC").
		Find(&recommendations).Error
	return recommendations, err
}

// Student profile methods for new structure
func (r *QuestionnaireRepository) GetStudentByProfileIDNew(profileID uuid.UUID) (*user.StudentProfile, error) {
	var student user.StudentProfile
	err := r.db.Preload("User").Where("id = ?", profileID).First(&student).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("student tidak ditemukan")
		}
		return nil, err
	}
	return &student, nil
}

func (r *QuestionnaireRepository) GetStudentProfileIDByUserID(userID uuid.UUID) (uuid.UUID, error) {
	var profile user.StudentProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return uuid.Nil, err
	}
	return profile.ID, nil
}
func (r *QuestionnaireRepository) GetResponseByIDAndStudent(responseID, studentProfileID uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Where("id = ? AND student_profile_id = ?", responseID, studentProfileID).
		Preload("Questionnaire").
		First(&response).Error
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) UpdateResponseProcessingStatus(responseID uuid.UUID, status string, recommendations, analysis *string) error {
	updateData := map[string]interface{}{
		"processed_at": time.Now(),
	}

	if recommendations != nil {
		updateData["ai_recommendations"] = *recommendations
	}

	if analysis != nil {
		updateData["ai_analysis"] = *analysis
	}

	return r.db.Model(&QuestionnaireResponse{}).Where("id = ?", responseID).Updates(updateData).Error
}

func (r *QuestionnaireRepository) UpdateResponseProcessedAt(responseID uuid.UUID, processedAt *time.Time) error {
	return r.db.Model(&QuestionnaireResponse{}).Where("id = ?", responseID).Update("processed_at", processedAt).Error
}

func (r *QuestionnaireRepository) UpdateResponseRecommendedRole(responseID uuid.UUID, roleID *uuid.UUID) error {
	return r.db.Model(&QuestionnaireResponse{}).Where("id = ?", responseID).Update("recommended_profiling_role_id", roleID).Error
}
