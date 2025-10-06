package student

import (
	"errors"

	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	"github.com/google/uuid"
)

type StudentRoadmapService struct {
	repo *roadmap.RoadmapRepository
}

func NewStudentRoadmapService(repo *roadmap.RoadmapRepository) *StudentRoadmapService {
	return &StudentRoadmapService{repo: repo}
}

func (s *StudentRoadmapService) GetAvailableRoadmaps(studentProfileID uuid.UUID, page, limit int) (*PaginatedRoadmapsResponse, error) {
	offset := (page - 1) * limit
	roadmaps, total, err := s.repo.GetAvailableRoadmaps(studentProfileID, offset, limit)
	if err != nil {
		return nil, errors.New("failed to get available roadmaps")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	roadmapResponses := make([]AvailableRoadmapResponse, len(roadmaps))
	for i, rm := range roadmaps {
		isStarted := false
		progressPercent := float64(0)

		if len(rm.StudentProgress) > 0 {
			isStarted = true
			progressPercent = rm.StudentProgress[0].ProgressPercent
		}

		totalDuration := 0
		difficultyLevel := "beginner"
		for _, step := range rm.Steps {
			totalDuration += step.EstimatedDuration
			if step.DifficultyLevel == "advanced" {
				difficultyLevel = "advanced"
			} else if step.DifficultyLevel == "intermediate" && difficultyLevel != "advanced" {
				difficultyLevel = "intermediate"
			}
		}

		roleInfo, _ := s.repo.GetProfilingRole(rm.ProfilingRoleID)
		roleName := "Unknown Role"
		if roleInfo != nil {
			if role, ok := roleInfo.(map[string]interface{}); ok {
				if name, exists := role["name"].(string); exists {
					roleName = name
				}
			}
		}

		roadmapResponses[i] = AvailableRoadmapResponse{
			ID:                rm.ID,
			RoadmapName:       rm.RoadmapName,
			Description:       rm.Description,
			RoleName:          roleName,
			TotalSteps:        len(rm.Steps),
			EstimatedDuration: totalDuration,
			DifficultyLevel:   difficultyLevel,
			IsStarted:         isStarted,
			ProgressPercent:   progressPercent,
		}
	}

	return &PaginatedRoadmapsResponse{
		Data:       roadmapResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *StudentRoadmapService) StartRoadmap(roadmapID, studentProfileID uuid.UUID) (*StartRoadmapResponse, error) {
	existingProgress, _ := s.repo.GetStudentRoadmapProgress(roadmapID, studentProfileID)
	if existingProgress != nil {
		return nil, errors.New("roadmap already started")
	}

	progress, err := s.repo.StartRoadmap(roadmapID, studentProfileID)
	if err != nil {
		return nil, errors.New("failed to start roadmap")
	}

	roadmapData, err := s.repo.GetRoadmapByID(roadmapID)
	if err != nil {
		return nil, errors.New("failed to get roadmap details")
	}

	return &StartRoadmapResponse{
		Message: "Roadmap started successfully",
		RoadmapProgress: RoadmapProgressResponse{
			ID:              progress.ID,
			RoadmapID:       progress.RoadmapID,
			RoadmapName:     roadmapData.RoadmapName,
			TotalSteps:      progress.TotalSteps,
			CompletedSteps:  progress.CompletedSteps,
			ProgressPercent: progress.ProgressPercent,
			StartedAt:       progress.StartedAt,
		},
	}, nil
}

func (s *StudentRoadmapService) GetRoadmapProgress(roadmapID, studentProfileID uuid.UUID) (*RoadmapProgressResponse, error) {
	roadmapData, progress, stepProgress, err := s.repo.GetStudentRoadmapWithProgress(roadmapID, studentProfileID)
	if err != nil {
		return nil, errors.New("roadmap progress not found")
	}

	roleInfo, _ := s.repo.GetProfilingRole(roadmapData.ProfilingRoleID)
	roleName := "Unknown Role"
	if roleInfo != nil {
		if role, ok := roleInfo.(map[string]interface{}); ok {
			if name, exists := role["name"].(string); exists {
				roleName = name
			}
		}
	}

	steps := make([]StudentStepResponse, len(stepProgress))
	for i, sp := range stepProgress {
		canStart := sp.Status == roadmap.RoadmapProgressStatusUnlocked
		canSubmit := sp.Status == roadmap.RoadmapProgressStatusInProgress && sp.SubmittedAt == nil
		isLocked := sp.Status == roadmap.RoadmapProgressStatusLocked

		steps[i] = StudentStepResponse{
			ID:                   sp.RoadmapStep.ID,
			StepOrder:            sp.RoadmapStep.StepOrder,
			Title:                sp.RoadmapStep.Title,
			Description:          sp.RoadmapStep.Description,
			LearningObjectives:   sp.RoadmapStep.LearningObjectives,
			SubmissionGuidelines: sp.RoadmapStep.SubmissionGuidelines,
			ResourceLinks:        sp.RoadmapStep.ResourceLinks,
			EstimatedDuration:    sp.RoadmapStep.EstimatedDuration,
			DifficultyLevel:      sp.RoadmapStep.DifficultyLevel,
			Status:               string(sp.Status),
			EvidenceLink:         sp.EvidenceLink,
			EvidenceType:         sp.EvidenceType,
			SubmissionNotes:      sp.SubmissionNotes,
			ValidationNotes:      sp.ValidationNotes,
			ValidationScore:      sp.ValidationScore,
			StartedAt:            sp.StartedAt,
			SubmittedAt:          sp.SubmittedAt,
			CompletedAt:          sp.CompletedAt,
			CanStart:             canStart,
			CanSubmit:            canSubmit,
			IsLocked:             isLocked,
		}
	}

	return &RoadmapProgressResponse{
		ID:                 progress.ID,
		RoadmapID:          progress.RoadmapID,
		RoadmapName:        roadmapData.RoadmapName,
		RoadmapDescription: roadmapData.Description,
		RoleName:           roleName,
		TotalSteps:         progress.TotalSteps,
		CompletedSteps:     progress.CompletedSteps,
		ProgressPercent:    progress.ProgressPercent,
		StartedAt:          progress.StartedAt,
		LastActivityAt:     progress.LastActivityAt,
		CompletedAt:        progress.CompletedAt,
		Steps:              steps,
	}, nil
}

func (s *StudentRoadmapService) StartStep(stepID, studentProfileID uuid.UUID) (*StartStepResponse, error) {
	progress, step, err := s.getStepProgress(stepID, studentProfileID)
	if err != nil {
		return nil, err
	}

	stepOrder := 1
	for _, roadmapStep := range step.Roadmap.Steps {
		if roadmapStep.ID == stepID {
			stepOrder = roadmapStep.StepOrder
			break
		}
	}

	if stepOrder > 1 {
		prevApproved, err := s.repo.IsPreviousStepApproved(stepID, studentProfileID)
		if err != nil || !prevApproved {
			return nil, errors.New("previous step must be approved by teacher first")
		}
	}

	if progress.Status != roadmap.RoadmapProgressStatusUnlocked && progress.Status != roadmap.RoadmapProgressStatusRejected {
		return nil, errors.New("step is not available to start")
	}

	err = s.repo.StartStep(progress.ID)
	if err != nil {
		return nil, errors.New("failed to start step")
	}

	return &StartStepResponse{
		Message:    "Step started successfully",
		StepStatus: string(roadmap.RoadmapProgressStatusInProgress),
		StartedAt:  step.UpdatedAt,
	}, nil
}

func (s *StudentRoadmapService) SubmitEvidence(req SubmitEvidenceRequest, studentProfileID uuid.UUID) (*SubmitEvidenceResponse, error) {
	progress, _, err := s.getStepProgress(req.StepID, studentProfileID)
	if err != nil {
		return nil, err
	}

	if progress.Status != roadmap.RoadmapProgressStatusInProgress && progress.Status != roadmap.RoadmapProgressStatusRejected {
		return nil, errors.New("step is not available for submission")
	}

	if progress.Status == roadmap.RoadmapProgressStatusSubmitted {
		return nil, errors.New("evidence already submitted and waiting for review")
	}

	err = s.repo.UpdateStepStatusToSubmitted(progress.ID, req.EvidenceLink, req.EvidenceType, req.SubmissionNotes)
	if err != nil {
		return nil, errors.New("failed to submit evidence")
	}

	return &SubmitEvidenceResponse{
		Message:     "Evidence submitted successfully. Waiting for teacher review.",
		StepStatus:  string(roadmap.RoadmapProgressStatusSubmitted),
		SubmittedAt: progress.UpdatedAt,
	}, nil
}

func (s *StudentRoadmapService) GetMyProgress(studentProfileID uuid.UUID) (*MyProgressResponse, error) {
	progressList, err := s.repo.GetMyProgress(studentProfileID)
	if err != nil {
		return nil, errors.New("failed to get progress")
	}

	var activeRoadmaps []ActiveRoadmapSummary
	var completedRoadmaps []CompletedRoadmapSummary
	totalStarted := 0
	totalCompleted := 0
	totalStepsCompleted := 0

	for _, progress := range progressList {
		totalStarted++
		totalStepsCompleted += progress.CompletedSteps

		roleInfo, _ := s.repo.GetProfilingRole(progress.Roadmap.ProfilingRoleID)
		roleName := "Unknown Role"
		if roleInfo != nil {
			if role, ok := roleInfo.(map[string]interface{}); ok {
				if name, exists := role["name"].(string); exists {
					roleName = name
				}
			}
		}

		if progress.CompletedAt != nil {
			totalCompleted++
			completedRoadmaps = append(completedRoadmaps, CompletedRoadmapSummary{
				ID:            progress.ID,
				RoadmapName:   progress.Roadmap.RoadmapName,
				RoleName:      roleName,
				TotalSteps:    progress.TotalSteps,
				CompletedAt:   *progress.CompletedAt,
				TotalDuration: 0,
			})
		} else {
			activeRoadmaps = append(activeRoadmaps, ActiveRoadmapSummary{
				ID:              progress.ID,
				RoadmapName:     progress.Roadmap.RoadmapName,
				RoleName:        roleName,
				TotalSteps:      progress.TotalSteps,
				CompletedSteps:  progress.CompletedSteps,
				ProgressPercent: progress.ProgressPercent,
				StartedAt:       progress.StartedAt,
				LastActivityAt:  progress.LastActivityAt,
			})
		}
	}

	avgCompletion := float64(0)
	if totalStarted > 0 {
		avgCompletion = (float64(totalCompleted) / float64(totalStarted)) * 100
	}

	return &MyProgressResponse{
		ActiveRoadmaps:    activeRoadmaps,
		CompletedRoadmaps: completedRoadmaps,
		OverallStats: StudentStatistics{
			TotalRoadmapsStarted:   totalStarted,
			TotalRoadmapsCompleted: totalCompleted,
			TotalStepsCompleted:    totalStepsCompleted,
			AverageCompletionRate:  avgCompletion,
		},
	}, nil
}

func (s *StudentRoadmapService) getStepProgress(stepID, studentProfileID uuid.UUID) (*roadmap.StudentStepProgress, *roadmap.RoadmapStep, error) {
	step, err := s.repo.GetRoadmapStepByID(stepID)
	if err != nil {
		return nil, nil, errors.New("step not found")
	}

	roadmapProgress, err := s.repo.GetStudentRoadmapProgress(step.RoadmapID, studentProfileID)
	if err != nil {
		return nil, nil, errors.New("roadmap not started")
	}

	stepProgress, err := s.repo.GetStudentStepProgress(roadmapProgress.ID, stepID)
	if err != nil {
		return nil, nil, errors.New("step progress not found")
	}

	return stepProgress, step, nil
}
