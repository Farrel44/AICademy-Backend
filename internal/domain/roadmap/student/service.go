package student

import (
	"errors"
	"fmt"

	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	"github.com/google/uuid"
)

type StudentRoadmapService struct {
	repo *roadmap.RoadmapRepository
}

func NewStudentRoadmapService(repo *roadmap.RoadmapRepository) *StudentRoadmapService {
	return &StudentRoadmapService{repo: repo}
}

func (s *StudentRoadmapService) GetMyRoadmap(studentProfileID uuid.UUID) (*MyRoadmapResponse, error) {
	// Get the assigned roadmap for this student
	roadmap, err := s.repo.GetStudentAssignedRoadmap(studentProfileID)
	if err != nil {
		if err.Error() == "no recommended role found for student" {
			return &MyRoadmapResponse{
				Message: "Silakan lengkapi questionnaire profiling karir terlebih dahulu untuk melihat roadmap yang sesuai dengan minat dan kemampuan Anda",
			}, nil
		}
		if err.Error() == "no active roadmap found for student's recommended role" {
			return &MyRoadmapResponse{
				Message: "Belum ada roadmap yang tersedia untuk role yang direkomendasikan",
			}, nil
		}
		return nil, errors.New("gagal mendapatkan roadmap")
	}

	// Get role information
	roleInfo, err := s.repo.GetProfilingRole(roadmap.ProfilingRoleID)
	if err != nil {
		return nil, errors.New("gagal mendapatkan informasi role")
	}

	roleName := "Role Tidak Diketahui"
	if roleInfo != nil {
		if name, exists := roleInfo["name"].(string); exists {
			roleName = name
		}
	}

	// Get student's progress for this roadmap
	roadmapProgress, err := s.repo.GetStudentRoadmapProgress(roadmap.ID, studentProfileID)

	var progressInfo *RoadmapProgressInfo
	var steps []StudentStepResponse

	if roadmapProgress != nil {
		// Student has started this roadmap - get progress details using existing method approach
		progressInfo = &RoadmapProgressInfo{
			ID:              roadmapProgress.ID,
			TotalSteps:      roadmapProgress.TotalSteps,
			CompletedSteps:  roadmapProgress.CompletedSteps,
			ProgressPercent: roadmapProgress.ProgressPercent,
			IsFinished:      roadmapProgress.CompletedAt != nil,
			StartedAt:       roadmapProgress.StartedAt,
			CompletedAt:     roadmapProgress.CompletedAt,
			LastActivityAt:  roadmapProgress.LastActivityAt,
		}

		// Build steps with progress - use roadmap steps
		steps = make([]StudentStepResponse, len(roadmap.Steps))
		for i, step := range roadmap.Steps {
			stepResponse := StudentStepResponse{
				ID:                   step.ID,
				StepOrder:            step.StepOrder,
				Title:                step.Title,
				Description:          step.Description,
				LearningObjectives:   step.LearningObjectives,
				SubmissionGuidelines: step.SubmissionGuidelines,
				ResourceLinks:        step.ResourceLinks,
				EstimatedDuration:    step.EstimatedDuration,
				DifficultyLevel:      step.DifficultyLevel,
			}

			// Get step progress if exists
			stepProgress, err := s.repo.GetStudentStepProgress(roadmapProgress.ID, step.ID)
			if err == nil && stepProgress != nil {
				// Step has progress
				stepResponse.Status = string(stepProgress.Status)
				stepResponse.EvidenceLink = stepProgress.EvidenceLink
				stepResponse.EvidenceType = stepProgress.EvidenceType
				stepResponse.SubmissionNotes = stepProgress.SubmissionNotes
				stepResponse.ValidationNotes = stepProgress.ValidationNotes
				stepResponse.ValidationScore = stepProgress.ValidationScore
				stepResponse.SubmittedAt = stepProgress.SubmittedAt
				stepResponse.CompletedAt = stepProgress.CompletedAt
				stepResponse.StartedAt = stepProgress.StartedAt

				// Set helper flags based on status
				stepResponse.CanStart = stepProgress.Status == "unlocked"
				stepResponse.CanSubmit = stepProgress.Status == "in_progress"
				stepResponse.IsLocked = stepProgress.Status == "locked"
			} else {
				// No progress yet, determine status based on position
				if step.StepOrder == 1 {
					stepResponse.Status = "unlocked"
					stepResponse.CanStart = true
					stepResponse.IsLocked = false
				} else {
					stepResponse.Status = "locked"
					stepResponse.CanStart = false
					stepResponse.IsLocked = true
				}
				stepResponse.CanSubmit = false
			}

			steps[i] = stepResponse
		}
	} else {
		// Student hasn't started this roadmap yet - show basic step info
		steps = make([]StudentStepResponse, len(roadmap.Steps))
		for i, step := range roadmap.Steps {
			stepResponse := StudentStepResponse{
				ID:                   step.ID,
				StepOrder:            step.StepOrder,
				Title:                step.Title,
				Description:          step.Description,
				LearningObjectives:   step.LearningObjectives,
				SubmissionGuidelines: step.SubmissionGuidelines,
				ResourceLinks:        step.ResourceLinks,
				EstimatedDuration:    step.EstimatedDuration,
				DifficultyLevel:      step.DifficultyLevel,
			}

			// Set status and helper flags for non-started roadmap
			if step.StepOrder == 1 {
				stepResponse.Status = "unlocked"
				stepResponse.CanStart = true
				stepResponse.IsLocked = false
			} else {
				stepResponse.Status = "locked"
				stepResponse.CanStart = false
				stepResponse.IsLocked = true
			}
			stepResponse.CanSubmit = false

			steps[i] = stepResponse
		}
	}

	// Calculate total duration and difficulty
	totalDuration := 0
	difficultyLevel := "beginner"
	for _, step := range roadmap.Steps {
		totalDuration += step.EstimatedDuration
		if step.DifficultyLevel == "advanced" {
			difficultyLevel = "advanced"
		} else if step.DifficultyLevel == "intermediate" && difficultyLevel != "advanced" {
			difficultyLevel = "intermediate"
		}
	}

	response := &MyRoadmapResponse{
		Message: "Roadmap berhasil diambil",
		Roadmap: &RoadmapDetailResponse{
			ID:                roadmap.ID,
			RoadmapName:       roadmap.RoadmapName,
			Description:       roadmap.Description,
			RoleName:          roleName,
			TotalSteps:        len(roadmap.Steps),
			EstimatedDuration: totalDuration,
			DifficultyLevel:   difficultyLevel,
			Progress:          progressInfo,
			Steps:             steps,
		},
	}

	return response, nil
}
func (s *StudentRoadmapService) StartRoadmap(roadmapID, studentProfileID uuid.UUID) (*StartRoadmapResponse, error) {
	existingProgress, _ := s.repo.GetStudentRoadmapProgress(roadmapID, studentProfileID)
	if existingProgress != nil {
		return nil, errors.New("roadmap sudah pernah dimulai")
	}

	progress, err := s.repo.StartRoadmap(roadmapID, studentProfileID)
	if err != nil {
		return nil, errors.New("gagal memulai roadmap")
	}

	roadmapData, err := s.repo.GetRoadmapByID(roadmapID)
	if err != nil {
		return nil, errors.New("gagal mengambil detail roadmap")
	}

	return &StartRoadmapResponse{
		Message: "Roadmap berhasil dimulai",
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
	// Get the assigned roadmap for this student
	assignedRoadmap, err := s.repo.GetStudentAssignedRoadmap(studentProfileID)
	if err != nil {
		return nil, errors.New("belum ada roadmap yang tersedia untuk siswa")
	}

	// Check if the requested roadmap matches the assigned roadmap
	if assignedRoadmap.ID != roadmapID {
		return nil, errors.New("roadmap tidak dapat diakses oleh siswa")
	}

	// Get progress for the assigned roadmap - if not started yet, return basic roadmap info
	progress, err := s.repo.GetStudentRoadmapProgress(roadmapID, studentProfileID)
	if err != nil {
		// Student hasn't started the roadmap yet - return roadmap with all steps locked
		return s.buildUnstartedRoadmapResponse(assignedRoadmap, studentProfileID)
	}

	// Get all step progress
	stepProgress, err := s.repo.GetStudentStepProgressList(progress.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil progress step")
	}

	roleInfo, _ := s.repo.GetProfilingRole(assignedRoadmap.ProfilingRoleID)
	roleName := "Role Tidak Diketahui"
	if roleInfo != nil {
		if name, exists := roleInfo["name"].(string); exists {
			roleName = name
		}
	}

	// If no step progress found, build from roadmap steps
	if len(stepProgress) == 0 {
		return s.buildProgressFromRoadmapSteps(assignedRoadmap, progress, roleName)
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
		RoadmapName:        assignedRoadmap.RoadmapName,
		RoadmapDescription: assignedRoadmap.Description,
		RoleName:           roleName,
		TotalSteps:         progress.TotalSteps,
		CompletedSteps:     progress.CompletedSteps,
		ProgressPercent:    progress.ProgressPercent,
		IsFinished:         progress.CompletedAt != nil,
		StartedAt:          progress.StartedAt,
		LastActivityAt:     progress.LastActivityAt,
		CompletedAt:        progress.CompletedAt,
		Steps:              steps,
	}, nil
}

func (s *StudentRoadmapService) StartStep(stepID, studentProfileID uuid.UUID) (*StartStepResponse, error) {
	progress, step, err := s.getStepProgress(stepID, studentProfileID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data step: %w", err)
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
		if err != nil {
			return nil, fmt.Errorf("gagal memeriksa step sebelumnya: %w", err)
		}
		if !prevApproved {
			return nil, errors.New("step sebelumnya harus disetujui oleh pengajar terlebih dahulu")
		}
	}

	switch progress.Status {
	case roadmap.RoadmapProgressStatusUnlocked, roadmap.RoadmapProgressStatusRejected:
		// boleh mulai
	case roadmap.RoadmapProgressStatusInProgress:
		return nil, errors.New("step sudah sedang dikerjakan")
	case roadmap.RoadmapProgressStatusCompleted:
		return nil, errors.New("step ini sudah selesai dikerjakan")
	case roadmap.RoadmapProgressStatusApproved:
		return nil, errors.New("step ini sudah disetujui, tidak bisa dimulai ulang")
	default:
		return nil, fmt.Errorf("status step tidak dikenali: %s", progress.Status)
	}

	if err := s.repo.StartStep(progress.ID); err != nil {
		return nil, fmt.Errorf("gagal memulai step: %w", err)
	}

	return &StartStepResponse{
		Message:    "Step berhasil dimulai",
		StepStatus: string(roadmap.RoadmapProgressStatusInProgress),
		StartedAt:  step.UpdatedAt,
	}, nil
}

func (s *StudentRoadmapService) SubmitEvidence(req SubmitEvidenceRequest, studentProfileID uuid.UUID) (*SubmitEvidenceResponse, error) {
	progress, _, err := s.getStepProgress(req.StepID, studentProfileID)
	if err != nil {
		return nil, err
	}

	if req.EvidenceType != "url" {
		return nil, errors.New("only URL evidence type is supported currently")
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

func (s *StudentRoadmapService) GetStepProgress(stepID, studentProfileID uuid.UUID) (*StudentStepResponse, error) {
	// Get the assigned roadmap for this student
	assignedRoadmap, err := s.repo.GetStudentAssignedRoadmap(studentProfileID)
	if err != nil {
		return nil, errors.New("no roadmap assigned to student")
	}

	// Get step details
	step, err := s.repo.GetRoadmapStepByID(stepID)
	if err != nil {
		return nil, errors.New("step not found")
	}

	// Check if step belongs to assigned roadmap
	if step.RoadmapID != assignedRoadmap.ID {
		return nil, errors.New("step not accessible to student")
	}

	// Get roadmap progress
	roadmapProgress, err := s.repo.GetStudentRoadmapProgress(assignedRoadmap.ID, studentProfileID)
	if err != nil {
		return nil, errors.New("roadmap not started")
	}

	// Get step progress
	stepProgress, err := s.repo.GetStudentStepProgress(roadmapProgress.ID, stepID)
	if err != nil {
		return nil, errors.New("step progress not found")
	}

	canStart := stepProgress.Status == roadmap.RoadmapProgressStatusUnlocked
	canSubmit := stepProgress.Status == roadmap.RoadmapProgressStatusInProgress && stepProgress.SubmittedAt == nil
	isLocked := stepProgress.Status == roadmap.RoadmapProgressStatusLocked

	return &StudentStepResponse{
		ID:                   step.ID,
		StepOrder:            step.StepOrder,
		Title:                step.Title,
		Description:          step.Description,
		LearningObjectives:   step.LearningObjectives,
		SubmissionGuidelines: step.SubmissionGuidelines,
		ResourceLinks:        step.ResourceLinks,
		EstimatedDuration:    step.EstimatedDuration,
		DifficultyLevel:      step.DifficultyLevel,
		Status:               string(stepProgress.Status),
		EvidenceLink:         stepProgress.EvidenceLink,
		EvidenceType:         stepProgress.EvidenceType,
		SubmissionNotes:      stepProgress.SubmissionNotes,
		ValidationNotes:      stepProgress.ValidationNotes,
		ValidationScore:      stepProgress.ValidationScore,
		StartedAt:            stepProgress.StartedAt,
		SubmittedAt:          stepProgress.SubmittedAt,
		CompletedAt:          stepProgress.CompletedAt,
		CanStart:             canStart,
		CanSubmit:            canSubmit,
		IsLocked:             isLocked,
	}, nil
}

func (s *StudentRoadmapService) buildProgressFromRoadmapSteps(assignedRoadmap *roadmap.FeatureRoadmap, progress *roadmap.StudentRoadmapProgress, roleName string) (*RoadmapProgressResponse, error) {
	// Get roadmap steps
	steps, err := s.repo.GetRoadmapSteps(assignedRoadmap.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data step roadmap")
	}

	// Build step responses based on roadmap steps
	stepResponses := make([]StudentStepResponse, len(steps))
	for i, step := range steps {
		status := roadmap.RoadmapProgressStatusLocked
		canStart := false
		canSubmit := false
		isLocked := true

		// First step should be unlocked for starting
		if i == 0 {
			status = roadmap.RoadmapProgressStatusUnlocked
			canStart = true
			isLocked = false
		}

		stepResponses[i] = StudentStepResponse{
			ID:                   step.ID,
			StepOrder:            step.StepOrder,
			Title:                step.Title,
			Description:          step.Description,
			LearningObjectives:   step.LearningObjectives,
			SubmissionGuidelines: step.SubmissionGuidelines,
			ResourceLinks:        step.ResourceLinks,
			EstimatedDuration:    step.EstimatedDuration,
			DifficultyLevel:      step.DifficultyLevel,
			Status:               string(status),
			CanStart:             canStart,
			CanSubmit:            canSubmit,
			IsLocked:             isLocked,
		}
	}

	return &RoadmapProgressResponse{
		ID:                 progress.ID,
		RoadmapID:          progress.RoadmapID,
		RoadmapName:        assignedRoadmap.RoadmapName,
		RoadmapDescription: assignedRoadmap.Description,
		RoleName:           roleName,
		TotalSteps:         progress.TotalSteps,
		CompletedSteps:     progress.CompletedSteps,
		ProgressPercent:    progress.ProgressPercent,
		IsFinished:         progress.CompletedAt != nil,
		StartedAt:          progress.StartedAt,
		LastActivityAt:     progress.LastActivityAt,
		CompletedAt:        progress.CompletedAt,
		Steps:              stepResponses,
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

func (s *StudentRoadmapService) buildUnstartedRoadmapResponse(assignedRoadmap *roadmap.FeatureRoadmap, studentProfileID uuid.UUID) (*RoadmapProgressResponse, error) {
	roleInfo, _ := s.repo.GetProfilingRole(assignedRoadmap.ProfilingRoleID)
	roleName := "Role Tidak Diketahui"
	if roleInfo != nil {
		if name, exists := roleInfo["name"].(string); exists {
			roleName = name
		}
	}

	// Get roadmap steps
	steps, err := s.repo.GetRoadmapSteps(assignedRoadmap.ID)
	if err != nil {
		return nil, errors.New("gagal mengambil data step roadmap")
	}

	// Build step responses with all steps locked except the first one
	stepResponses := make([]StudentStepResponse, len(steps))
	for i, step := range steps {
		status := roadmap.RoadmapProgressStatusLocked
		canStart := false
		canSubmit := false
		isLocked := true

		// First step should be unlocked for starting
		if i == 0 {
			status = roadmap.RoadmapProgressStatusUnlocked
			canStart = true
			isLocked = false
		}

		stepResponses[i] = StudentStepResponse{
			ID:                   step.ID,
			StepOrder:            step.StepOrder,
			Title:                step.Title,
			Description:          step.Description,
			LearningObjectives:   step.LearningObjectives,
			SubmissionGuidelines: step.SubmissionGuidelines,
			ResourceLinks:        step.ResourceLinks,
			EstimatedDuration:    step.EstimatedDuration,
			DifficultyLevel:      step.DifficultyLevel,
			Status:               string(status),
			CanStart:             canStart,
			CanSubmit:            canSubmit,
			IsLocked:             isLocked,
		}
	}

	return &RoadmapProgressResponse{
		ID:                 uuid.Nil,
		RoadmapID:          assignedRoadmap.ID,
		RoadmapName:        assignedRoadmap.RoadmapName,
		RoadmapDescription: assignedRoadmap.Description,
		RoleName:           roleName,
		TotalSteps:         len(steps),
		CompletedSteps:     0,
		ProgressPercent:    0,
		IsFinished:         false,
		StartedAt:          nil,
		LastActivityAt:     nil,
		CompletedAt:        nil,
		Steps:              stepResponses,
	}, nil
}
