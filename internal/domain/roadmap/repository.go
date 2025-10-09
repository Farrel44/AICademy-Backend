package roadmap

import (
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

func (r *RoadmapRepository) UpdateStepProgress(progress *StudentStepProgress) error {
	return r.db.Save(progress).Error
}

// Student Repository Methods

func (r *RoadmapRepository) GetAvailableRoadmaps(studentProfileID uuid.UUID, offset, limit int) ([]FeatureRoadmap, int64, error) {
	var roadmaps []FeatureRoadmap
	var total int64

	recommendedRoleID, _ := r.GetStudentRecommendedRole(studentProfileID)

	query := r.db.Model(&FeatureRoadmap{}).Where("status = ?", RoadmapStatusActive)

	if recommendedRoleID != nil {
		query = query.Where("profiling_role_id = ?", *recommendedRoleID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&roadmaps).Error

	for i := range roadmaps {
		progress, _ := r.GetStudentRoadmapProgress(roadmaps[i].ID, studentProfileID)
		if progress != nil {
			roadmaps[i].StudentProgress = []StudentRoadmapProgress{*progress}
		}
	}

	return roadmaps, total, err
}

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

func (r *RoadmapRepository) GetMyProgress(studentProfileID uuid.UUID) ([]StudentRoadmapProgress, error) {
	var progressList []StudentRoadmapProgress
	err := r.db.Where("student_profile_id = ?", studentProfileID).
		Preload("Roadmap").
		Order("created_at DESC").
		Find(&progressList).Error
	return progressList, err
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

	if err == gorm.ErrRecordNotFound || result.RecommendedProfilingRoleID == nil {
		return nil, nil
	}

	if err != nil {
		return nil, err
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

func (r *RoadmapRepository) GetProfilingRole(roleID uuid.UUID) (interface{}, error) {
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
	return role, nil
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
