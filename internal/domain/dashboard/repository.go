package dashboard

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	GetStudentAssignedRoadmapWithProgress(studentProfileID uuid.UUID) (*DashboardRoadmapData, error)
	GetStudentChallenges(studentProfileID uuid.UUID) ([]StudentChallengeItem, error)
	GetStudentSummary(studentProfileID uuid.UUID) (*StudentSummary, error)
	GetAdminTotals() (*AdminTotals, error)
	GetStudentStatistics() (*StudentStatistics, error)
	GetTeacherChallengeStats(teacherProfileID uuid.UUID) (*TeacherChallengeStats, error)
	GetTeacherSubmissionStats(teacherProfileID uuid.UUID) (*TeacherSubmissionStats, error)
	GetStudentProfileIDByUserID(userID uuid.UUID) (uuid.UUID, error)
	GetTeacherProfileIDByUserID(userID uuid.UUID) (uuid.UUID, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetStudentAssignedRoadmapWithProgress(studentProfileID uuid.UUID) (*DashboardRoadmapData, error) {
	// Get student's recommended role from questionnaire - same logic as roadmap repo
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

	if result.RecommendedProfilingRoleID == nil {
		return nil, nil
	}

	roleID, err := uuid.Parse(*result.RecommendedProfilingRoleID)
	if err != nil {
		return nil, err
	}

	// Get the roadmap for this role using GORM model
	var roadmap struct {
		ID          uuid.UUID `gorm:"column:id"`
		RoadmapName string    `gorm:"column:roadmap_name"`
		Description *string   `gorm:"column:description"`
	}

	err = r.db.Table("feature_roadmaps").
		Select("id, roadmap_name, description").
		Where("profiling_role_id = ? AND status = ?", roleID, "active").
		Scan(&roadmap).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Set basic roadmap data
	var roadmapData DashboardRoadmapData
	roadmapData.ID = roadmap.ID
	roadmapData.RoadmapName = roadmap.RoadmapName
	roadmapData.Description = roadmap.Description

	// Get role name
	var roleName string
	err = r.db.Table("profiling_roles").
		Select("name").
		Where("id = ?", roleID).
		Pluck("name", &roleName).Error

	if err == nil {
		roadmapData.RoleName = roleName
	}

	// Get student's progress for this roadmap
	var progressResult struct {
		ID              *string `gorm:"column:id"`
		CompletedSteps  int     `gorm:"column:completed_steps"`
		ProgressPercent float64 `gorm:"column:progress_percent"`
		StartedAt       *string `gorm:"column:started_at"`
		UpdatedAt       *string `gorm:"column:updated_at"`
		CompletedAt     *string `gorm:"column:completed_at"`
	}

	err = r.db.Table("student_roadmap_progress").
		Select("id, completed_steps, progress_percent, started_at, updated_at, completed_at").
		Where("roadmap_id = ? AND student_profile_id = ?", roadmapData.ID, studentProfileID).
		First(&progressResult).Error

	if err == nil && progressResult.ID != nil {
		progressID, parseErr := uuid.Parse(*progressResult.ID)
		if parseErr == nil {
			roadmapData.ProgressID = &progressID
			roadmapData.CompletedSteps = progressResult.CompletedSteps
			roadmapData.ProgressPercent = progressResult.ProgressPercent
			roadmapData.ProgressStartedAt = progressResult.StartedAt
			roadmapData.ProgressUpdatedAt = progressResult.UpdatedAt
			roadmapData.ProgressCompletedAt = progressResult.CompletedAt
		}
	}

	// Get steps with progress
	var steps []DashboardStepData
	query := `
		SELECT 
			rs.id,
			rs.step_order,
			rs.title,
			rs.description,
			rs.learning_objectives,
			rs.submission_guidelines,
			rs.resource_links,
			rs.estimated_duration,
			rs.difficulty_level,
			COALESCE(ssp.status, 'locked') as status,
			ssp.evidence_link,
			ssp.evidence_type,
			ssp.submission_notes,
			ssp.validation_notes,
			ssp.validation_score,
			ssp.started_at,
			ssp.submitted_at,
			ssp.completed_at
		FROM roadmap_steps rs
		LEFT JOIN student_step_progress ssp ON rs.id = ssp.roadmap_step_id AND ssp.student_roadmap_progress_id = ?
		WHERE rs.roadmap_id = ?
		ORDER BY rs.step_order
	`

	var progressIDParam interface{}
	if roadmapData.ProgressID != nil {
		progressIDParam = *roadmapData.ProgressID
	}

	err = r.db.Raw(query, progressIDParam, roadmapData.ID).Scan(&steps).Error
	if err != nil {
		return nil, err
	}

	roadmapData.Steps = steps
	return &roadmapData, nil
}

func (r *repository) GetStudentProfileIDByUserID(userID uuid.UUID) (uuid.UUID, error) {
	var studentProfile struct {
		ID uuid.UUID
	}

	err := r.db.Table("student_profiles").
		Select("id").
		Where("user_id = ?", userID).
		First(&studentProfile).Error

	return studentProfile.ID, err
}

func (r *repository) GetTeacherProfileIDByUserID(userID uuid.UUID) (uuid.UUID, error) {
	var teacherProfile struct {
		ID uuid.UUID
	}

	err := r.db.Table("teacher_profiles").
		Select("id").
		Where("user_id = ?", userID).
		First(&teacherProfile).Error

	return teacherProfile.ID, err
}

func (r *repository) GetStudentProgress(studentProfileID uuid.UUID) ([]StudentProgressItem, error) {
	var progress []StudentProgressItem

	// Get student's assigned roadmap based on questionnaire profiling - use same pattern as my-roadmap
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
			return progress, nil // Return empty progress if no questionnaire found
		}
		return progress, err
	}

	if result.RecommendedProfilingRoleID == nil {
		return progress, nil // Return empty progress if no recommended role
	}

	// Parse the UUID string
	recommendedRoleID, err := uuid.Parse(*result.RecommendedProfilingRoleID)
	if err != nil {
		return progress, err
	}

	// Get the specific roadmap for this recommended role
	var assignedRoadmap struct {
		ID   uuid.UUID
		Name string
	}
	err = r.db.Table("feature_roadmaps").
		Select("id, roadmap_name as name").
		Where("profiling_role_id = ? AND status = ?", recommendedRoleID, "active").
		Scan(&assignedRoadmap).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return progress, nil // Return empty progress if no active roadmap found
		}
		return progress, err
	}

	// Check if student has progress for this roadmap
	var progressResult struct {
		ID *string `gorm:"column:id"`
	}
	err = r.db.Table("student_roadmap_progress").
		Select("id").
		Where("roadmap_id = ? AND student_profile_id = ?", assignedRoadmap.ID, studentProfileID).
		First(&progressResult).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Student hasn't started this roadmap yet, show all steps as not completed
			err = r.db.Table("roadmap_steps").
				Select(`
					? as roadmap_id,
					? as roadmap_title,
					id as step_id,
					title as step_title,
					step_order,
					false as is_completed,
					NULL as completed_at
				`, assignedRoadmap.ID, assignedRoadmap.Name).
				Where("roadmap_id = ?", assignedRoadmap.ID).
				Order("step_order").
				Find(&progress).Error
			return progress, err
		}
		return progress, err
	}

	if progressResult.ID == nil {
		// Student hasn't started this roadmap yet, show all steps as not completed
		err = r.db.Table("roadmap_steps").
			Select(`
				? as roadmap_id,
				? as roadmap_title,
				id as step_id,
				title as step_title,
				step_order,
				false as is_completed,
				NULL as completed_at
			`, assignedRoadmap.ID, assignedRoadmap.Name).
			Where("roadmap_id = ?", assignedRoadmap.ID).
			Order("step_order").
			Find(&progress).Error
		return progress, err
	}

	// Parse the UUID string
	studentRoadmapProgressID, err := uuid.Parse(*progressResult.ID)
	if err != nil {
		return progress, err
	}

	// Student has started this roadmap, get actual progress
	err = r.db.Table("roadmap_steps rs").
		Select(`
			? as roadmap_id,
			? as roadmap_title,
			rs.id as step_id,
			rs.title as step_title,
			rs.step_order,
			COALESCE(ssp.status = 'completed', false) as is_completed,
			ssp.completed_at
		`, assignedRoadmap.ID, assignedRoadmap.Name).
		Joins("LEFT JOIN student_step_progress ssp ON rs.id = ssp.roadmap_step_id AND ssp.student_roadmap_progress_id = ?", studentRoadmapProgressID).
		Where("rs.roadmap_id = ?", assignedRoadmap.ID).
		Order("rs.step_order").
		Find(&progress).Error

	return progress, err
}

func (r *repository) GetStudentChallenges(studentProfileID uuid.UUID) ([]StudentChallengeItem, error) {
	var challenges []StudentChallengeItem

	err := r.db.Table("team_members tm").
		Select(`
			c.id as challenge_id,
			c.title as challenge_title,
			t.id as team_id,
			t.team_name,
			s.id as submission_id,
			COALESCE(s.points, 0) as points,
			c.deadline,
			(c.deadline > NOW()) as is_active,
			tm.joined_at as registered_at
		`).
		Joins("JOIN teams t ON tm.team_id = t.id").
		Joins("LEFT JOIN submissions s ON t.id = s.team_id").
		Joins("JOIN challenges c ON s.challenge_id = c.id").
		Where("tm.student_profile_id = ?", studentProfileID).
		Order("c.deadline DESC").
		Limit(10).
		Find(&challenges).Error

	return challenges, err
}

func (r *repository) GetStudentSummary(studentProfileID uuid.UUID) (*StudentSummary, error) {
	summary := &StudentSummary{}

	// Get student's assigned roadmap based on questionnaire profiling - use same pattern as my-roadmap
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
			// No questionnaire found, return zero summary
			summary.TotalRoadmaps = 0
			summary.CompletedSteps = 0
			summary.TotalSteps = 0
			summary.ActiveChallenges = 0
			summary.CompletedChallenges = 0
			summary.TotalPoints = 0
			return summary, nil
		}
		return nil, err
	}

	if result.RecommendedProfilingRoleID == nil {
		// No recommended role, return zero summary
		summary.TotalRoadmaps = 0
		summary.CompletedSteps = 0
		summary.TotalSteps = 0
		summary.ActiveChallenges = 0
		summary.CompletedChallenges = 0
		summary.TotalPoints = 0
		return summary, nil
	}

	// Parse the UUID string
	recommendedRoleID, err := uuid.Parse(*result.RecommendedProfilingRoleID)
	if err != nil {
		return nil, err
	}

	// Get the specific roadmap for this recommended role
	var roadmapResult struct {
		ID *string `gorm:"column:id"`
	}
	err = r.db.Table("feature_roadmaps").
		Select("id").
		Where("profiling_role_id = ? AND status = ?", recommendedRoleID, "active").
		First(&roadmapResult).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No active roadmap found, return zero summary
			summary.TotalRoadmaps = 0
			summary.CompletedSteps = 0
			summary.TotalSteps = 0
			summary.ActiveChallenges = 0
			summary.CompletedChallenges = 0
			summary.TotalPoints = 0
			return summary, nil
		}
		return nil, err
	}

	if roadmapResult.ID == nil {
		// No active roadmap found, return zero summary
		summary.TotalRoadmaps = 0
		summary.CompletedSteps = 0
		summary.TotalSteps = 0
		summary.ActiveChallenges = 0
		summary.CompletedChallenges = 0
		summary.TotalPoints = 0
		return summary, nil
	}

	// Parse the UUID string
	assignedRoadmapID, err := uuid.Parse(*roadmapResult.ID)
	if err != nil {
		return nil, err
	}

	// Check if student has progress for this roadmap
	var progressResult struct {
		ID *string `gorm:"column:id"`
	}
	err = r.db.Table("student_roadmap_progress").
		Select("id").
		Where("roadmap_id = ? AND student_profile_id = ?", assignedRoadmapID, studentProfileID).
		First(&progressResult).Error

	var studentRoadmapProgressID *uuid.UUID
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err != gorm.ErrRecordNotFound && progressResult.ID != nil {
		parsedID, parseErr := uuid.Parse(*progressResult.ID)
		if parseErr != nil {
			return nil, parseErr
		}
		studentRoadmapProgressID = &parsedID
	}

	// Count stats for the assigned roadmap
	type RoadmapStats struct {
		TotalRoadmaps  int
		CompletedSteps int
		TotalSteps     int
	}

	var roadmapStats RoadmapStats

	if studentRoadmapProgressID != nil {
		// Student has started this roadmap
		err = r.db.Table("roadmap_steps rs").
			Select(`
				1 as total_roadmaps,
				COUNT(CASE WHEN ssp.status = 'completed' THEN 1 END) as completed_steps,
				COUNT(rs.id) as total_steps
			`).
			Joins("LEFT JOIN student_step_progress ssp ON rs.id = ssp.roadmap_step_id AND ssp.student_roadmap_progress_id = ?", *studentRoadmapProgressID).
			Where("rs.roadmap_id = ?", assignedRoadmapID).
			Scan(&roadmapStats).Error
	} else {
		// Student hasn't started this roadmap yet
		err = r.db.Table("roadmap_steps").
			Select(`
				1 as total_roadmaps,
				0 as completed_steps,
				COUNT(id) as total_steps
			`).
			Where("roadmap_id = ?", assignedRoadmapID).
			Scan(&roadmapStats).Error
	}

	if err != nil {
		return nil, err
	}

	summary.TotalRoadmaps = roadmapStats.TotalRoadmaps
	summary.CompletedSteps = roadmapStats.CompletedSteps
	summary.TotalSteps = roadmapStats.TotalSteps

	type ChallengeStats struct {
		ActiveChallenges    int
		CompletedChallenges int
		TotalPoints         int
	}

	var challengeStats ChallengeStats
	err = r.db.Table("team_members tm").
		Select(`
			COUNT(CASE WHEN c.deadline > NOW() THEN 1 END) as active_challenges,
			COUNT(CASE WHEN c.deadline <= NOW() THEN 1 END) as completed_challenges,
			COALESCE(SUM(s.points), 0) as total_points
		`).
		Joins("JOIN teams t ON tm.team_id = t.id").
		Joins("LEFT JOIN submissions s ON t.id = s.team_id").
		Joins("JOIN challenges c ON s.challenge_id = c.id").
		Where("tm.student_profile_id = ?", studentProfileID).
		Scan(&challengeStats).Error

	if err != nil {
		return nil, err
	}

	summary.ActiveChallenges = challengeStats.ActiveChallenges
	summary.CompletedChallenges = challengeStats.CompletedChallenges
	summary.TotalPoints = challengeStats.TotalPoints

	return summary, nil
}

func (r *repository) GetAdminTotals() (*AdminTotals, error) {
	totals := &AdminTotals{}

	err := r.db.Model(&struct{}{}).Table("users").Count(&totals.TotalUsers).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&struct{}{}).Table("users").Where("role = ?", "student").Count(&totals.TotalStudents).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&struct{}{}).Table("users").Where("role = ?", "teacher").Count(&totals.TotalTeachers).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&struct{}{}).Table("users").Where("role = ?", "company").Count(&totals.TotalCompanies).Error
	if err != nil {
		return nil, err
	}

	return totals, nil
}

func (r *repository) GetStudentStatistics() (*StudentStatistics, error) {
	stats := &StudentStatistics{}

	err := r.db.Model(&struct{}{}).Table("student_profiles").
		Select(`
			COUNT(CASE WHEN UPPER(class) LIKE '%TKJ%' THEN 1 END) as total_tkj,
			COUNT(CASE WHEN UPPER(class) LIKE '%TJA%' THEN 1 END) as total_tja,
			COUNT(CASE WHEN UPPER(class) LIKE '%PPLG%' THEN 1 END) as total_pplg,
			COUNT(CASE WHEN UPPER(class) LIKE '%RPL%' THEN 1 END) as total_rpl
		`).
		Scan(stats).Error

	return stats, err
}

func (r *repository) GetTeacherChallengeStats(teacherProfileID uuid.UUID) (*TeacherChallengeStats, error) {
	stats := &TeacherChallengeStats{}

	err := r.db.Model(&struct{}{}).Table("challenges").
		Select(`
			COUNT(*) as total_challenges,
			COUNT(CASE WHEN deadline > NOW() THEN 1 END) as active_challenges,
			COUNT(CASE WHEN deadline <= NOW() THEN 1 END) as completed_challenges
		`).
		Where("created_by_teacher_id = ?", teacherProfileID).
		Scan(stats).Error

	return stats, err
}

func (r *repository) GetTeacherSubmissionStats(teacherProfileID uuid.UUID) (*TeacherSubmissionStats, error) {
	stats := &TeacherSubmissionStats{}

	err := r.db.Table("submissions s").
		Select(`
			COUNT(s.id) as total_submissions,
			COUNT(CASE WHEN s.points IS NOT NULL THEN 1 END) as scored_submissions,
			COUNT(CASE WHEN s.points IS NULL THEN 1 END) as pending_submissions
		`).
		Joins("JOIN challenges c ON s.challenge_id = c.id").
		Where("c.created_by_teacher_id = ?", teacherProfileID).
		Scan(stats).Error

	return stats, err
}

type StudentProgressItem struct {
	RoadmapID    uuid.UUID `json:"roadmap_id"`
	RoadmapTitle string    `json:"roadmap_title"`
	StepID       uuid.UUID `json:"step_id"`
	StepTitle    string    `json:"step_title"`
	StepOrder    int       `json:"step_order"`
	IsCompleted  bool      `json:"is_completed"`
	CompletedAt  *string   `json:"completed_at"`
}

type StudentChallengeItem struct {
	ChallengeID    uuid.UUID  `json:"challenge_id"`
	ChallengeTitle string     `json:"challenge_title"`
	TeamID         uuid.UUID  `json:"team_id"`
	TeamName       string     `json:"team_name"`
	SubmissionID   *uuid.UUID `json:"submission_id"`
	Points         int        `json:"points"`
	Deadline       string     `json:"deadline"`
	IsActive       bool       `json:"is_active"`
	RegisteredAt   string     `json:"registered_at"`
}

type StudentSummary struct {
	TotalRoadmaps       int `json:"total_roadmaps"`
	CompletedSteps      int `json:"completed_steps"`
	TotalSteps          int `json:"total_steps"`
	ActiveChallenges    int `json:"active_challenges"`
	CompletedChallenges int `json:"completed_challenges"`
	TotalPoints         int `json:"total_points"`
}

type AdminTotals struct {
	TotalUsers     int64 `json:"total_users"`
	TotalStudents  int64 `json:"total_students"`
	TotalTeachers  int64 `json:"total_teachers"`
	TotalCompanies int64 `json:"total_companies"`
}

type StudentStatistics struct {
	TotalTKJ  int64 `json:"total_tkj"`
	TotalTJA  int64 `json:"total_tja"`
	TotalPPLG int64 `json:"total_pplg"`
	TotalRPL  int64 `json:"total_rpl"`
}

type TeacherChallengeStats struct {
	TotalChallenges     int `json:"total_challenges"`
	ActiveChallenges    int `json:"active_challenges"`
	CompletedChallenges int `json:"completed_challenges"`
}

type TeacherSubmissionStats struct {
	TotalSubmissions   int `json:"total_submissions"`
	ScoredSubmissions  int `json:"scored_submissions"`
	PendingSubmissions int `json:"pending_submissions"`
}

// Dashboard roadmap data - internal struct for repository
type DashboardRoadmapData struct {
	ID                uuid.UUID `json:"id"`
	RoadmapName       string    `json:"roadmap_name"`
	Description       *string   `json:"description"`
	RoleName          string    `json:"role_name"`
	TotalSteps        int       `json:"total_steps"`
	EstimatedDuration int       `json:"estimated_duration"`
	DifficultyLevel   string    `json:"difficulty_level"`

	// Progress info (if exists)
	ProgressID          *uuid.UUID `json:"progress_id"`
	CompletedSteps      int        `json:"completed_steps"`
	ProgressPercent     float64    `json:"progress_percent"`
	ProgressStartedAt   *string    `json:"progress_started_at"`
	ProgressUpdatedAt   *string    `json:"progress_updated_at"`
	ProgressCompletedAt *string    `json:"progress_completed_at"`

	// Steps with progress
	Steps []DashboardStepData `json:"steps"`
}

type DashboardStepData struct {
	ID                   uuid.UUID `json:"id"`
	StepOrder            int       `json:"step_order"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	LearningObjectives   string    `json:"learning_objectives"`
	SubmissionGuidelines string    `json:"submission_guidelines"`
	ResourceLinks        *string   `json:"resource_links"`
	EstimatedDuration    int       `json:"estimated_duration"`
	DifficultyLevel      string    `json:"difficulty_level"`

	// Progress data (if exists)
	Status          string  `json:"status"`
	EvidenceLink    *string `json:"evidence_link"`
	EvidenceType    *string `json:"evidence_type"`
	SubmissionNotes *string `json:"submission_notes"`
	ValidationNotes *string `json:"validation_notes"`
	ValidationScore *int    `json:"validation_score"`
	StartedAt       *string `json:"started_at"`
	SubmittedAt     *string `json:"submitted_at"`
	CompletedAt     *string `json:"completed_at"`
}
