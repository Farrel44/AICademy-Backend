package roadmap

import (
	"errors"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoadmapRepository struct {
	db *gorm.DB
}

func NewRoadmapRepository(db *gorm.DB) *RoadmapRepository {
	return &RoadmapRepository{db: db}
}

// Admin Repository Methods

func (r *RoadmapRepository) CreateRoadmap(roadmap *FeatureRoadmap) error {
	return r.db.Create(roadmap).Error
}

func (r *RoadmapRepository) GetRoadmapByID(id uuid.UUID) (*FeatureRoadmap, error) {
	var roadmap FeatureRoadmap
	err := r.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Where("id = ?", id).First(&roadmap).Error
	if err != nil {
		return nil, err
	}
	return &roadmap, nil
}

// ini dipakai admin
func (r *RoadmapRepository) GetRoadmaps(offset, limit int, profilingRoleID *uuid.UUID, status *RoadmapStatus) ([]FeatureRoadmap, int64, error) {
	var roadmaps []FeatureRoadmap
	var total int64

	query := r.db.Model(&FeatureRoadmap{})

	if profilingRoleID != nil {
		query = query.Where("profiling_role_id = ?", *profilingRoleID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Steps").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&roadmaps).Error

	return roadmaps, total, err
}

// Optimized methods for search performance
func (r *RoadmapRepository) CountRoadmaps(search string, profilingRoleID *uuid.UUID, status *RoadmapStatus) (int64, error) {
	var total int64
	query := r.db.Model(&FeatureRoadmap{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(roadmap_name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if profilingRoleID != nil {
		query = query.Where("profiling_role_id = ?", *profilingRoleID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	return total, err
}

func (r *RoadmapRepository) GetRoadmapsOptimized(offset, limit int, search string, profilingRoleID *uuid.UUID, status *RoadmapStatus) ([]FeatureRoadmap, error) {
	var roadmaps []FeatureRoadmap
	query := r.db.Select("feature_roadmaps.*").Model(&FeatureRoadmap{})

	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(feature_roadmaps.roadmap_name) LIKE ? OR LOWER(feature_roadmaps.description) LIKE ?", searchTerm, searchTerm)
	}

	if profilingRoleID != nil {
		query = query.Where("profiling_role_id = ?", *profilingRoleID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Order("feature_roadmaps.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&roadmaps).Error

	return roadmaps, err
}

func (r *RoadmapRepository) UpdateRoadmap(roadmap *FeatureRoadmap) error {
	return r.db.Save(roadmap).Error
}

func (r *RoadmapRepository) DeleteRoadmap(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("roadmap_id = ?", id).Delete(&RoadmapStep{}).Error; err != nil {
			return err
		}
		return tx.Delete(&FeatureRoadmap{}, id).Error
	})
}

func (r *RoadmapRepository) GetRoadmapByProfilingRoleID(profilingRoleID uuid.UUID) (*FeatureRoadmap, error) {
	var roadmap FeatureRoadmap
	err := r.db.Where("profiling_role_id = ?", profilingRoleID).First(&roadmap).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error if not found
		}
		return nil, err
	}
	return &roadmap, nil
}

func (r *RoadmapRepository) CreateRoadmapStep(step *RoadmapStep) error {
	return r.db.Create(step).Error
}

func (r *RoadmapRepository) GetRoadmapSteps(roadmapID uuid.UUID) ([]RoadmapStep, error) {
	var steps []RoadmapStep
	err := r.db.Where("roadmap_id = ?", roadmapID).
		Order("step_order ASC").
		Find(&steps).Error
	return steps, err
}

func (r *RoadmapRepository) GetRoadmapStepByID(stepID uuid.UUID) (*RoadmapStep, error) {
	var step RoadmapStep
	err := r.db.Where("id = ?", stepID).First(&step).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *RoadmapRepository) UpdateRoadmapStep(step *RoadmapStep) error {
	return r.db.Save(step).Error
}

func (r *RoadmapRepository) DeleteRoadmapStep(stepID uuid.UUID) error {
	return r.db.Delete(&RoadmapStep{}, stepID).Error
}

func (r *RoadmapRepository) GetStudentRoadmapProgress(roadmapID, studentProfileID uuid.UUID) (*StudentRoadmapProgress, error) {
	var progress StudentRoadmapProgress
	err := r.db.Where("roadmap_id = ? AND student_profile_id = ?", roadmapID, studentProfileID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *RoadmapRepository) GetStudentStepProgress(roadmapProgressID, stepID uuid.UUID) (*StudentStepProgress, error) {
	var progress StudentStepProgress
	err := r.db.Where("student_roadmap_progress_id = ? AND roadmap_step_id = ?", roadmapProgressID, stepID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *RoadmapRepository) GetStudentStepProgressList(roadmapProgressID uuid.UUID) ([]StudentStepProgress, error) {
	var stepProgress []StudentStepProgress
	err := r.db.Where("student_roadmap_progress_id = ?", roadmapProgressID).
		Preload("RoadmapStep").
		Joins("JOIN roadmap_steps ON student_step_progress.roadmap_step_id = roadmap_steps.id").
		Order("roadmap_steps.step_order ASC").
		Find(&stepProgress).Error
	return stepProgress, err
}

func (r *RoadmapRepository) GetAllStudentProgress(roadmapID uuid.UUID, offset, limit int) ([]StudentRoadmapProgress, int64, error) {
	var progressList []StudentRoadmapProgress
	var total int64

	query := r.db.Model(&StudentRoadmapProgress{}).Where("roadmap_id = ?", roadmapID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("StepProgress").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&progressList).Error

	return progressList, total, err
}

func (r *RoadmapRepository) GetPendingSubmissions(teacherID *uuid.UUID, offset, limit int) ([]StudentStepProgress, int64, error) {
	var submissions []StudentStepProgress
	var total int64

	query := r.db.Model(&StudentStepProgress{}).
		Where("status = ? AND submitted_at IS NOT NULL", RoadmapProgressStatusSubmitted)

	if teacherID != nil {
		query = query.Where("validated_by_teacher_id = ? OR validated_by_teacher_id IS NULL", *teacherID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("StudentRoadmapProgress.StudentProfile.User").
		Preload("StudentRoadmapProgress.Roadmap").
		Preload("RoadmapStep").
		Order("submitted_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error

	return submissions, total, err
}

func (r *RoadmapRepository) GetPendingSubmissionsOptimized(teacherID *uuid.UUID, offset, limit int, search string) ([]StudentStepProgress, int64, error) {
	var submissions []StudentStepProgress
	var total int64

	// Base condition for both count and data queries
	baseCondition := "student_step_progress.status = ? AND student_step_progress.submitted_at IS NOT NULL"
	args := []interface{}{string(RoadmapProgressStatusSubmitted)}

	// Separate count query
	countQuery := r.db.Model(&StudentStepProgress{}).
		Joins("LEFT JOIN student_roadmap_progress srp ON student_step_progress.student_roadmap_progress_id = srp.id").
		Joins("LEFT JOIN student_profiles sp ON srp.student_profile_id = sp.id").
		Joins("LEFT JOIN users u ON sp.user_id = u.id").
		Joins("LEFT JOIN feature_roadmaps fr ON srp.roadmap_id = fr.id").
		Where(baseCondition, args...)

	// Data query
	dataQuery := r.db.Model(&StudentStepProgress{}).
		Where(baseCondition, args...)

	// Apply teacher filter
	if teacherID != nil {
		teacherCondition := "(student_step_progress.validated_by_teacher_id = ? OR student_step_progress.validated_by_teacher_id IS NULL)"
		countQuery = countQuery.Where(teacherCondition, *teacherID)
		dataQuery = dataQuery.Where(teacherCondition, *teacherID)
	}

	// Apply search filter
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		searchCondition := "LOWER(sp.fullname) LIKE ? OR LOWER(u.email) LIKE ? OR LOWER(fr.roadmap_name) LIKE ?"
		countQuery = countQuery.Where(searchCondition, searchTerm, searchTerm, searchTerm)

		// For data query, add the joins needed for search
		dataQuery = dataQuery.
			Joins("LEFT JOIN student_roadmap_progress srp ON student_step_progress.student_roadmap_progress_id = srp.id").
			Joins("LEFT JOIN student_profiles sp ON srp.student_profile_id = sp.id").
			Joins("LEFT JOIN users u ON sp.user_id = u.id").
			Joins("LEFT JOIN feature_roadmaps fr ON srp.roadmap_id = fr.id").
			Where(searchCondition, searchTerm, searchTerm, searchTerm)
	}

	// Get count
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get data
	err := dataQuery.
		Preload("StudentRoadmapProgress.StudentProfile.User").
		Preload("StudentRoadmapProgress.Roadmap").
		Preload("RoadmapStep").
		Order("student_step_progress.submitted_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error

	return submissions, total, err
}

func (r *RoadmapRepository) UpdateStepProgress(progress *StudentStepProgress) error {
	return r.db.Save(progress).Error
}

// Student Repository Methods

func (r *RoadmapRepository) StartRoadmap(roadmapID, studentProfileID uuid.UUID) (*StudentRoadmapProgress, error) {
	steps, err := r.GetRoadmapSteps(roadmapID)
	if err != nil {
		return nil, err
	}

	var roadmapProgress *StudentRoadmapProgress

	err = r.db.Transaction(func(tx *gorm.DB) error {
		roadmapProgress = &StudentRoadmapProgress{
			RoadmapID:        roadmapID,
			StudentProfileID: studentProfileID,
			TotalSteps:       len(steps),
			CompletedSteps:   0,
			ProgressPercent:  0,
		}

		if err := tx.Create(roadmapProgress).Error; err != nil {
			return err
		}

		for i, step := range steps {
			status := RoadmapProgressStatusLocked
			if i == 0 {
				status = RoadmapProgressStatusUnlocked
			}

			stepProgress := &StudentStepProgress{
				StudentRoadmapProgressID: roadmapProgress.ID,
				RoadmapStepID:            step.ID,
				Status:                   status,
			}

			if err := tx.Create(stepProgress).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return roadmapProgress, nil
}

func (r *RoadmapRepository) GetStudentRoadmapWithProgress(roadmapID, studentProfileID uuid.UUID) (*FeatureRoadmap, *StudentRoadmapProgress, []StudentStepProgress, error) {
	roadmap, err := r.GetRoadmapByID(roadmapID)
	if err != nil {
		return nil, nil, nil, err
	}

	progress, err := r.GetStudentRoadmapProgress(roadmapID, studentProfileID)
	if err != nil {
		return nil, nil, nil, err
	}

	var stepProgress []StudentStepProgress
	err = r.db.Where("student_roadmap_progress_id = ?", progress.ID).
		Preload("RoadmapStep").
		Order("RoadmapStep.step_order ASC").
		Find(&stepProgress).Error
	if err != nil {
		return nil, nil, nil, err
	}

	return roadmap, progress, stepProgress, nil
}

func (r *RoadmapRepository) StartStep(stepProgressID uuid.UUID) error {
	now := gorm.Expr("NOW()")
	return r.db.Model(&StudentStepProgress{}).
		Where("id = ? AND status = ?", stepProgressID, RoadmapProgressStatusUnlocked).
		Updates(map[string]interface{}{
			"status":     RoadmapProgressStatusInProgress,
			"started_at": now,
		}).Error
}

func (r *RoadmapRepository) SubmitEvidence(stepProgressID uuid.UUID, evidenceLink, evidenceType string, notes *string) error {
	now := gorm.Expr("NOW()")
	updates := map[string]interface{}{
		"evidence_link":    evidenceLink,
		"evidence_type":    evidenceType,
		"submission_notes": notes,
		"submitted_at":     now,
	}

	return r.db.Model(&StudentStepProgress{}).
		Where("id = ? AND status = ?", stepProgressID, RoadmapProgressStatusInProgress).
		Updates(updates).Error
}

func (r *RoadmapRepository) ApproveStep(stepProgressID, teacherID uuid.UUID, score *int, notes *string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get teacher profile ID from user ID
		var teacherProfile struct {
			ID uuid.UUID `json:"id"`
		}
		err := tx.Table("teacher_profiles").
			Select("id").
			Where("user_id = ?", teacherID).
			First(&teacherProfile).Error
		if err != nil {
			return err
		}

		now := gorm.Expr("NOW()")
		updates := map[string]interface{}{
			"status":                  RoadmapProgressStatusCompleted,
			"validated_by_teacher_id": teacherProfile.ID,
			"validation_score":        score,
			"validation_notes":        notes,
			"completed_at":            now,
		}

		err = tx.Model(&StudentStepProgress{}).
			Where("id = ?", stepProgressID).
			Updates(updates).Error
		if err != nil {
			return err
		}

		var stepProgress StudentStepProgress
		err = tx.Preload("StudentRoadmapProgress").
			Where("id = ?", stepProgressID).
			First(&stepProgress).Error
		if err != nil {
			return err
		}

		err = r.updateRoadmapProgress(tx, stepProgress.StudentRoadmapProgressID)
		if err != nil {
			return err
		}

		return r.unlockNextStep(tx, stepProgress.StudentRoadmapProgressID, stepProgress.RoadmapStepID)
	})
}

func (r *RoadmapRepository) RejectStep(stepProgressID, teacherID uuid.UUID, notes *string) error {
	// Get teacher profile ID from user ID
	var teacherProfile struct {
		ID uuid.UUID `json:"id"`
	}
	err := r.db.Table("teacher_profiles").
		Select("id").
		Where("user_id = ?", teacherID).
		First(&teacherProfile).Error
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"validated_by_teacher_id": teacherProfile.ID,
		"validation_notes":        notes,
		"status":                  RoadmapProgressStatusRejected,
	}

	return r.db.Model(&StudentStepProgress{}).
		Where("id = ?", stepProgressID).
		Updates(updates).Error
}

// GetStudentAssignedRoadmap retrieves the single roadmap assigned to a student based on their questionnaire profile
func (r *RoadmapRepository) GetStudentAssignedRoadmap(studentProfileID uuid.UUID) (*FeatureRoadmap, error) {
	// Get student's recommended role from questionnaire
	recommendedRoleID, err := r.GetStudentRecommendedRole(studentProfileID)
	if err != nil {
		return nil, err
	}

	if recommendedRoleID == nil {
		return nil, errors.New("no recommended role found for student")
	}

	// Get the roadmap for this role
	var roadmap FeatureRoadmap
	err = r.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Where("profiling_role_id = ? AND status = ?", *recommendedRoleID, RoadmapStatusActive).
		First(&roadmap).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("no active roadmap found for student's recommended role")
		}
		return nil, err
	}

	return &roadmap, nil
}

// Helper Methods

func (r *RoadmapRepository) GetStudentProfile(userID uuid.UUID) (*user.StudentProfile, error) {
	var profile user.StudentProfile
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *RoadmapRepository) GetTeacherProfile(teacherProfileID uuid.UUID) (*user.TeacherProfile, error) {
	var profile user.TeacherProfile
	err := r.db.Preload("User").Where("id = ?", teacherProfileID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *RoadmapRepository) GetStudentRecommendedRole(studentProfileID uuid.UUID) (*uuid.UUID, error) {
	var result struct {
		RecommendedProfilingRoleID *string `gorm:"column:recommended_profiling_role_id"`
	}

	err := r.db.Table("questionnaire_responses").
		Select("recommended_profiling_role_id").
		Where("student_profile_id = ? AND recommended_profiling_role_id IS NOT NULL", studentProfileID).
		Order("submitted_at DESC").
		Limit(1).
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if err == gorm.ErrRecordNotFound || result.RecommendedProfilingRoleID == nil {
		return nil, nil
	}
	roleID, err := uuid.Parse(*result.RecommendedProfilingRoleID)
	if err != nil {
		return nil, err
	}

	return &roleID, nil
}

func (r *RoadmapRepository) GetNextStepOrder(roadmapID uuid.UUID) (int, error) {
	var maxOrder int
	err := r.db.Model(&RoadmapStep{}).
		Where("roadmap_id = ?", roadmapID).
		Select("COALESCE(MAX(step_order), 0)").
		Scan(&maxOrder).Error

	if err != nil {
		return 0, err
	}

	return maxOrder + 1, nil
}

func (r *RoadmapRepository) GetProfilingRole(roleID uuid.UUID) (map[string]interface{}, error) {
	var role struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
	}

	err := r.db.Table("target_roles").
		Select("id, name, description").
		Where("id = ?", roleID).
		First(&role).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
	}, nil
}

func (r *RoadmapRepository) updateRoadmapProgress(tx *gorm.DB, roadmapProgressID uuid.UUID) error {
	var completedCount int64
	var totalCount int64

	err := tx.Model(&StudentStepProgress{}).
		Where("student_roadmap_progress_id = ?", roadmapProgressID).
		Count(&totalCount).Error
	if err != nil {
		return err
	}

	err = tx.Model(&StudentStepProgress{}).
		Where("student_roadmap_progress_id = ? AND status = ?", roadmapProgressID, RoadmapProgressStatusCompleted).
		Count(&completedCount).Error
	if err != nil {
		return err
	}

	progressPercent := float64(0)
	if totalCount > 0 {
		progressPercent = (float64(completedCount) / float64(totalCount)) * 100
	}

	updates := map[string]interface{}{
		"completed_steps":  completedCount,
		"progress_percent": progressPercent,
		"last_activity_at": gorm.Expr("NOW()"),
	}

	if completedCount == totalCount {
		updates["completed_at"] = gorm.Expr("NOW()")
	}

	return tx.Model(&StudentRoadmapProgress{}).
		Where("id = ?", roadmapProgressID).
		Updates(updates).Error
}

func (r *RoadmapRepository) unlockNextStep(tx *gorm.DB, roadmapProgressID, currentStepID uuid.UUID) error {
	var currentStep RoadmapStep
	err := tx.Where("id = ?", currentStepID).First(&currentStep).Error
	if err != nil {
		return err
	}

	var nextStep RoadmapStep
	err = tx.Where("roadmap_id = ? AND step_order = ?", currentStep.RoadmapID, currentStep.StepOrder+1).
		First(&nextStep).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	return tx.Model(&StudentStepProgress{}).
		Where("student_roadmap_progress_id = ? AND roadmap_step_id = ?", roadmapProgressID, nextStep.ID).
		Update("status", RoadmapProgressStatusUnlocked).Error
}

func (r *RoadmapRepository) GetRoadmapStatistics() (map[string]interface{}, error) {
	var totalRoadmaps int64
	var activeRoadmaps int64
	var totalStudentProgress int64
	var completedProgress int64

	r.db.Model(&FeatureRoadmap{}).Count(&totalRoadmaps)
	r.db.Model(&FeatureRoadmap{}).Where("status = ?", RoadmapStatusActive).Count(&activeRoadmaps)
	r.db.Model(&StudentRoadmapProgress{}).Count(&totalStudentProgress)
	r.db.Model(&StudentRoadmapProgress{}).Where("completed_at IS NOT NULL").Count(&completedProgress)

	return map[string]interface{}{
		"total_roadmaps":         totalRoadmaps,
		"active_roadmaps":        activeRoadmaps,
		"total_student_progress": totalStudentProgress,
		"completed_progress":     completedProgress,
	}, nil
}

func (r *RoadmapRepository) IsPreviousStepApproved(stepID, studentProfileID uuid.UUID) (bool, error) {
	var currentStep RoadmapStep
	err := r.db.First(&currentStep, stepID).Error
	if err != nil {
		return false, err
	}

	if currentStep.StepOrder <= 1 {
		return true, nil
	}

	var prevStep RoadmapStep
	err = r.db.Where("roadmap_id = ? AND step_order = ?", currentStep.RoadmapID, currentStep.StepOrder-1).
		First(&prevStep).Error
	if err != nil {
		return false, err
	}

	var roadmapProgress StudentRoadmapProgress
	err = r.db.Where("roadmap_id = ? AND student_profile_id = ?", currentStep.RoadmapID, studentProfileID).
		First(&roadmapProgress).Error
	if err != nil {
		return false, err
	}

	var prevStepProgress StudentStepProgress
	err = r.db.Where("student_roadmap_progress_id = ? AND roadmap_step_id = ?", roadmapProgress.ID, prevStep.ID).
		First(&prevStepProgress).Error
	if err != nil {
		return false, err
	}

	return prevStepProgress.Status == RoadmapProgressStatusApproved, nil
}

func (r *RoadmapRepository) UpdateStepStatusToSubmitted(progressID uuid.UUID, evidenceLink, evidenceType string, notes *string) error {
	updates := map[string]interface{}{
		"evidence_link":    evidenceLink,
		"evidence_type":    evidenceType,
		"submission_notes": notes,
		"status":           RoadmapProgressStatusSubmitted,
		"submitted_at":     time.Now(),
	}

	return r.db.Model(&StudentStepProgress{}).
		Where("id = ?", progressID).
		Updates(updates).Error
}
