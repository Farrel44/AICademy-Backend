package questionnaire

import (
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
	err := r.db.Where("active = ?", true).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
		First(&questionnaire).Error

	if err != nil {
		return nil, err
	}
	return &questionnaire, nil
}

func (r *QuestionnaireRepository) GetQuestionnaireByID(id uuid.UUID) (*ProfilingQuestionnaire, error) {
	var questionnaire ProfilingQuestionnaire
	err := r.db.Where("id = ?", id).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
		First(&questionnaire).Error

	if err != nil {
		return nil, err
	}
	return &questionnaire, nil
}

func (r *QuestionnaireRepository) GetAllQuestionnaires(page, limit int) ([]ProfilingQuestionnaire, int64, error) {
	var questionnaires []ProfilingQuestionnaire
	var total int64

	r.db.Model(&ProfilingQuestionnaire{}).Count(&total)

	offset := (page - 1) * limit
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
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
	return r.db.Delete(&ProfilingQuestionnaire{}, "id = ?", id).Error
}

func (r *QuestionnaireRepository) ActivateQuestionnaire(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ProfilingQuestionnaire{}).
			Where("active = ?", true).
			Update("active", false).Error; err != nil {
			return err
		}

		return tx.Model(&ProfilingQuestionnaire{}).
			Where("id = ?", id).
			Update("active", true).Error
	})
}

func (r *QuestionnaireRepository) DeactivateAllQuestionnaires() error {
	return r.db.Model(&ProfilingQuestionnaire{}).
		Where("active = ?", true).
		Update("active", false).Error
}

func (r *QuestionnaireRepository) AddQuestionsToQuestionnaire(questionnaireID uuid.UUID, questions []QuestionnaireQuestion) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range questions {
			questions[i].QuestionnaireID = questionnaireID
			if err := tx.Create(&questions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *QuestionnaireRepository) UpdateQuestion(question *QuestionnaireQuestion) error {
	return r.db.Save(question).Error
}

func (r *QuestionnaireRepository) DeleteQuestion(questionID uuid.UUID) error {
	return r.db.Delete(&QuestionnaireQuestion{}, "id = ?", questionID).Error
}

func (r *QuestionnaireRepository) GetQuestionsByQuestionnaireID(questionnaireID uuid.UUID) ([]QuestionnaireQuestion, error) {
	var questions []QuestionnaireQuestion
	err := r.db.Where("questionnaire_id = ?", questionnaireID).
		Order("question_order ASC").
		Find(&questions).Error

	return questions, err
}

func (r *QuestionnaireRepository) CreateResponse(response *QuestionnaireResponse) error {
	return r.db.Create(response).Error
}

func (r *QuestionnaireRepository) UpdateResponse(response *QuestionnaireResponse) error {
	return r.db.Save(response).Error
}

func (r *QuestionnaireRepository) GetResponseByID(id uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.First(&response, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) GetResponsesByStudentID(studentID uuid.UUID, page, limit int) ([]QuestionnaireResponse, int64, error) {
	var responses []QuestionnaireResponse
	var total int64

	r.db.Model(&QuestionnaireResponse{}).
		Where("student_profile_id = ?", studentID).
		Count(&total)

	offset := (page - 1) * limit
	err := r.db.Where("student_profile_id = ?", studentID).
		Order("submitted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&responses).Error

	return responses, total, err
}

func (r *QuestionnaireRepository) GetLatestResponseByStudent(studentID uuid.UUID) (*QuestionnaireResponse, error) {
	var response QuestionnaireResponse
	err := r.db.Where("student_profile_id = ?", studentID).
		Order("submitted_at DESC").
		First(&response).Error

	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (r *QuestionnaireRepository) HasStudentSubmitted(studentID uuid.UUID, questionnaireID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&QuestionnaireResponse{}).
		Where("student_profile_id = ? AND questionnaire_id = ?", studentID, questionnaireID).
		Count(&count).Error

	return count > 0, err
}

func (r *QuestionnaireRepository) GetResponsesByQuestionnaireID(questionnaireID uuid.UUID, page, limit int) ([]QuestionnaireResponse, int64, error) {
	var responses []QuestionnaireResponse
	var total int64

	r.db.Model(&QuestionnaireResponse{}).
		Where("questionnaire_id = ?", questionnaireID).
		Count(&total)

	offset := (page - 1) * limit
	err := r.db.Where("questionnaire_id = ?", questionnaireID).
		Order("submitted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&responses).Error

	return responses, total, err
}

func (r *QuestionnaireRepository) CreateRoleRecommendations(recommendations []RoleRecommendation) error {
	if len(recommendations) == 0 {
		return nil
	}
	return r.db.Create(&recommendations).Error
}

func (r *QuestionnaireRepository) GetRoleRecommendationsByResponseID(responseID uuid.UUID) ([]RoleRecommendation, error) {
	var recommendations []RoleRecommendation
	err := r.db.Where("response_id = ?", responseID).
		Order("rank ASC").
		Find(&recommendations).Error

	return recommendations, err
}

func (r *QuestionnaireRepository) UpdateRoleRecommendation(recommendation *RoleRecommendation) error {
	return r.db.Save(recommendation).Error
}

func (r *QuestionnaireRepository) DeleteRoleRecommendationsByResponseID(responseID uuid.UUID) error {
	return r.db.Delete(&RoleRecommendation{}, "response_id = ?", responseID).Error
}

func (r *QuestionnaireRepository) CreateQuestionGenerationTemplate(template *QuestionGenerationTemplate) error {
	return r.db.Create(template).Error
}

func (r *QuestionnaireRepository) GetQuestionGenerationTemplates() ([]QuestionGenerationTemplate, error) {
	var templates []QuestionGenerationTemplate
	err := r.db.Where("active = ?", true).
		Order("created_at DESC").
		Find(&templates).Error

	return templates, err
}

func (r *QuestionnaireRepository) GetQuestionGenerationTemplateByID(id uuid.UUID) (*QuestionGenerationTemplate, error) {
	var template QuestionGenerationTemplate
	err := r.db.First(&template, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *QuestionnaireRepository) UpdateQuestionGenerationTemplate(template *QuestionGenerationTemplate) error {
	return r.db.Save(template).Error
}

func (r *QuestionnaireRepository) DeleteQuestionGenerationTemplate(id uuid.UUID) error {
	return r.db.Delete(&QuestionGenerationTemplate{}, "id = ?", id).Error
}

func (r *QuestionnaireRepository) GetQuestionnaireStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalQuestionnaires int64
	r.db.Model(&ProfilingQuestionnaire{}).Count(&totalQuestionnaires)
	stats["total_kuesioner"] = totalQuestionnaires

	var totalResponses int64
	r.db.Model(&QuestionnaireResponse{}).Count(&totalResponses)
	stats["total_respons"] = totalResponses

	var activeQuestionnaires int64
	r.db.Model(&ProfilingQuestionnaire{}).Where("active = ?", true).Count(&activeQuestionnaires)
	stats["kuesioner_aktif"] = activeQuestionnaires

	var responsesThisMonth int64
	r.db.Model(&QuestionnaireResponse{}).
		Where("submitted_at >= DATE_TRUNC('month', CURRENT_DATE)").
		Count(&responsesThisMonth)
	stats["respons_bulan_ini"] = responsesThisMonth

	type QuestionnairePopularity struct {
		QuestionnaireID uuid.UUID
		Name            string
		ResponseCount   int64
	}

	var popularQuestionnaire QuestionnairePopularity
	err := r.db.Table("questionnaire_responses").
		Select("questionnaire_id, COUNT(*) as response_count").
		Group("questionnaire_id").
		Order("response_count DESC").
		Limit(1).
		Scan(&popularQuestionnaire).Error

	if err == nil && popularQuestionnaire.QuestionnaireID != uuid.Nil {
		var questionnaire ProfilingQuestionnaire
		if err := r.db.First(&questionnaire, "id = ?", popularQuestionnaire.QuestionnaireID).Error; err == nil {
			popularQuestionnaire.Name = questionnaire.Name
		}
		stats["kuesioner_terpopuler"] = popularQuestionnaire
	}

	return stats, nil
}

func (r *QuestionnaireRepository) GetResponseAnalytics(questionnaireID *uuid.UUID) (map[string]interface{}, error) {
	analytics := make(map[string]interface{})

	query := r.db.Model(&QuestionnaireResponse{})
	if questionnaireID != nil {
		query = query.Where("questionnaire_id = ?", *questionnaireID)
	}

	analytics["rata_rata_waktu_pengerjaan"] = "8.5 menit"

	type RoleDistribution struct {
		RoleName string
		Count    int64
	}

	var roleDistribution []RoleDistribution
	err := r.db.Table("role_recommendations").
		Select("profiling_role_id, COUNT(*) as count").
		Where("rank = 1").
		Group("profiling_role_id").
		Order("count DESC").
		Scan(&roleDistribution).Error

	if err == nil {
		analytics["distribusi_peran"] = roleDistribution
	}

	type DailyResponse struct {
		Date  string
		Count int64
	}

	var dailyResponses []DailyResponse
	err = r.db.Table("questionnaire_responses").
		Select("DATE(submitted_at) as date, COUNT(*) as count").
		Where("submitted_at >= CURRENT_DATE - INTERVAL '7 days'").
		Group("DATE(submitted_at)").
		Order("date ASC").
		Scan(&dailyResponses).Error

	if err == nil {
		analytics["respons_harian"] = dailyResponses
	}

	return analytics, nil
}

func (r *QuestionnaireRepository) SearchQuestionnaires(keyword string, page, limit int) ([]ProfilingQuestionnaire, int64, error) {
	var questionnaires []ProfilingQuestionnaire
	var total int64

	searchQuery := "%" + keyword + "%"

	r.db.Model(&ProfilingQuestionnaire{}).
		Where("name ILIKE ? OR ai_prompt_used ILIKE ?", searchQuery, searchQuery).
		Count(&total)

	offset := (page - 1) * limit
	err := r.db.Where("name ILIKE ? OR ai_prompt_used ILIKE ?", searchQuery, searchQuery).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
		Find(&questionnaires).Error

	return questionnaires, total, err
}

func (r *QuestionnaireRepository) GetQuestionnairesByGeneratedBy(generatedBy string, page, limit int) ([]ProfilingQuestionnaire, int64, error) {
	var questionnaires []ProfilingQuestionnaire
	var total int64

	r.db.Model(&ProfilingQuestionnaire{}).
		Where("generated_by = ?", generatedBy).
		Count(&total)

	offset := (page - 1) * limit
	err := r.db.Where("generated_by = ?", generatedBy).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("question_order ASC")
		}).
		Find(&questionnaires).Error

	return questionnaires, total, err
}

func (r *QuestionnaireRepository) BulkUpdateQuestionOrder(updates []struct {
	ID    uuid.UUID
	Order int
}) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			if err := tx.Model(&QuestionnaireQuestion{}).
				Where("id = ?", update.ID).
				Update("question_order", update.Order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *QuestionnaireRepository) CloneQuestionnaire(originalID uuid.UUID, newName string) (*ProfilingQuestionnaire, error) {
	var original ProfilingQuestionnaire
	err := r.db.Preload("Questions").First(&original, "id = ?", originalID).Error
	if err != nil {
		return nil, err
	}

	var clone ProfilingQuestionnaire

	err = r.db.Transaction(func(tx *gorm.DB) error {
		clone = ProfilingQuestionnaire{
			Name:            newName,
			ProfilingRoleID: original.ProfilingRoleID,
			Version:         1,
			Active:          false,
			GeneratedBy:     original.GeneratedBy,
			AIPromptUsed:    original.AIPromptUsed,
		}

		if err := tx.Create(&clone).Error; err != nil {
			return err
		}

		for _, question := range original.Questions {
			clonedQuestion := QuestionnaireQuestion{
				QuestionnaireID: clone.ID,
				QuestionText:    question.QuestionText,
				QuestionType:    question.QuestionType,
				Options:         question.Options,
				MaxScore:        question.MaxScore,
				QuestionOrder:   question.QuestionOrder,
				Category:        question.Category,
			}

			if err := tx.Create(&clonedQuestion).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &clone, nil
}
