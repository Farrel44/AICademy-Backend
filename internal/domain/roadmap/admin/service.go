package admin

import (
	"errors"

	"github.com/Farrel44/AICademy-Backend/internal/domain/roadmap"
	"github.com/google/uuid"
)

type AdminRoadmapService struct {
	repo *roadmap.RoadmapRepository
}

func NewAdminRoadmapService(repo *roadmap.RoadmapRepository) *AdminRoadmapService {
	return &AdminRoadmapService{repo: repo}
}

func (s *AdminRoadmapService) CreateRoadmap(req CreateRoadmapRequest, adminID uuid.UUID) (*RoadmapResponse, error) {
	newRoadmap := &roadmap.FeatureRoadmap{
		ProfilingRoleID: req.ProfilingRoleID,
		RoadmapName:     req.RoadmapName,
		Description:     req.Description,
		Status:          roadmap.RoadmapStatusDraft,
		CreatedBy:       adminID,
	}

	err := s.repo.CreateRoadmap(newRoadmap)
	if err != nil {
		return nil, errors.New("failed to create roadmap")
	}

	return s.mapRoadmapToResponse(newRoadmap), nil
}

func (s *AdminRoadmapService) GetRoadmaps(page, limit int, profilingRoleID *uuid.UUID, status *string) (*PaginatedRoadmapsResponse, error) {
	offset := (page - 1) * limit

	var roadmapStatus *roadmap.RoadmapStatus
	if status != nil {
		rs := roadmap.RoadmapStatus(*status)
		roadmapStatus = &rs
	}

	roadmaps, total, err := s.repo.GetRoadmaps(offset, limit, profilingRoleID, roadmapStatus)
	if err != nil {
		return nil, errors.New("failed to get roadmaps")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	roadmapResponses := make([]RoadmapResponse, len(roadmaps))
	for i, rm := range roadmaps {
		roadmapResponses[i] = *s.mapRoadmapToResponse(&rm)
	}

	return &PaginatedRoadmapsResponse{
		Data:       roadmapResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminRoadmapService) GetRoadmapByID(id uuid.UUID) (*RoadmapResponse, error) {
	roadmapData, err := s.repo.GetRoadmapByID(id)
	if err != nil {
		return nil, errors.New("roadmap not found")
	}

	response := s.mapRoadmapToResponse(roadmapData)

	if len(roadmapData.Steps) > 0 {
		steps := make([]StepResponse, len(roadmapData.Steps))
		for i, step := range roadmapData.Steps {
			steps[i] = s.mapStepToResponse(&step)
		}
		response.Steps = steps
		response.TotalSteps = len(steps)
	}

	return response, nil
}

func (s *AdminRoadmapService) UpdateRoadmap(id uuid.UUID, req UpdateRoadmapRequest) (*RoadmapResponse, error) {
	roadmapData, err := s.repo.GetRoadmapByID(id)
	if err != nil {
		return nil, errors.New("roadmap not found")
	}

	if req.RoadmapName != "" {
		roadmapData.RoadmapName = req.RoadmapName
	}
	if req.Description != nil {
		roadmapData.Description = req.Description
	}
	if req.Status != "" {
		roadmapData.Status = roadmap.RoadmapStatus(req.Status)
	}

	err = s.repo.UpdateRoadmap(roadmapData)
	if err != nil {
		return nil, errors.New("failed to update roadmap")
	}

	return s.mapRoadmapToResponse(roadmapData), nil
}

func (s *AdminRoadmapService) DeleteRoadmap(id uuid.UUID) error {
	_, err := s.repo.GetRoadmapByID(id)
	if err != nil {
		return errors.New("roadmap not found")
	}

	return s.repo.DeleteRoadmap(id)
}

func (s *AdminRoadmapService) CreateRoadmapStep(roadmapID uuid.UUID, req CreateStepRequest) (*StepResponse, error) {
	_, err := s.repo.GetRoadmapByID(roadmapID)
	if err != nil {
		return nil, errors.New("roadmap not found")
	}

	newStep := &roadmap.RoadmapStep{
		RoadmapID:            roadmapID,
		StepOrder:            req.StepOrder,
		Title:                req.Title,
		Description:          req.Description,
		LearningObjectives:   req.LearningObjectives,
		SubmissionGuidelines: req.SubmissionGuidelines,
		ResourceLinks:        req.ResourceLinks,
		EstimatedDuration:    req.EstimatedDuration,
		DifficultyLevel:      req.DifficultyLevel,
	}

	err = s.repo.CreateRoadmapStep(newStep)
	if err != nil {
		return nil, errors.New("failed to create step")
	}

	return &StepResponse{
		ID:                   newStep.ID,
		RoadmapID:            newStep.RoadmapID,
		StepOrder:            newStep.StepOrder,
		Title:                newStep.Title,
		Description:          newStep.Description,
		LearningObjectives:   newStep.LearningObjectives,
		SubmissionGuidelines: newStep.SubmissionGuidelines,
		ResourceLinks:        newStep.ResourceLinks,
		EstimatedDuration:    newStep.EstimatedDuration,
		DifficultyLevel:      newStep.DifficultyLevel,
		CreatedAt:            newStep.CreatedAt,
		UpdatedAt:            newStep.UpdatedAt,
	}, nil
}

func (s *AdminRoadmapService) UpdateRoadmapStep(stepID uuid.UUID, req UpdateStepRequest) (*StepResponse, error) {
	step, err := s.repo.GetRoadmapStepByID(stepID)
	if err != nil {
		return nil, errors.New("step not found")
	}

	if req.Title != "" {
		step.Title = req.Title
	}
	if req.Description != "" {
		step.Description = req.Description
	}
	if req.LearningObjectives != "" {
		step.LearningObjectives = req.LearningObjectives
	}
	if req.SubmissionGuidelines != "" {
		step.SubmissionGuidelines = req.SubmissionGuidelines
	}
	if req.ResourceLinks != nil {
		step.ResourceLinks = req.ResourceLinks
	}
	if req.EstimatedDuration > 0 {
		step.EstimatedDuration = req.EstimatedDuration
	}
	if req.DifficultyLevel != "" {
		step.DifficultyLevel = req.DifficultyLevel
	}
	if req.StepOrder > 0 {
		step.StepOrder = req.StepOrder
	}

	err = s.repo.UpdateRoadmapStep(step)
	if err != nil {
		return nil, errors.New("failed to update step")
	}

	stepResponse := s.mapStepToResponse(step)
	return &stepResponse, nil
}

func (s *AdminRoadmapService) DeleteRoadmapStep(stepID uuid.UUID) error {
	_, err := s.repo.GetRoadmapStepByID(stepID)
	if err != nil {
		return errors.New("step not found")
	}

	return s.repo.DeleteRoadmapStep(stepID)
}

func (s *AdminRoadmapService) UpdateStepOrders(req BulkStepOrderRequest) error {
	steps := make([]roadmap.RoadmapStep, len(req.Steps))
	for i, stepUpdate := range req.Steps {
		steps[i] = roadmap.RoadmapStep{
			ID:        stepUpdate.StepID,
			StepOrder: stepUpdate.Order,
		}
	}

	return s.repo.UpdateStepOrders(steps)
}

func (s *AdminRoadmapService) GetAllStudentProgress(roadmapID uuid.UUID, page, limit int) (*PaginatedSubmissionsResponse, error) {
	offset := (page - 1) * limit
	progressList, total, err := s.repo.GetAllStudentProgress(roadmapID, offset, limit)
	if err != nil {
		return nil, errors.New("failed to get student progress")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	summaries := make([]ProgressSummary, len(progressList))
	for i, progress := range progressList {
		student, _ := s.repo.GetStudentProfile(progress.StudentProfileID)

		summaries[i] = ProgressSummary{
			StudentID:       progress.StudentProfileID,
			StudentName:     student.Fullname,
			StudentEmail:    student.User.Email,
			TotalSteps:      progress.TotalSteps,
			CompletedSteps:  progress.CompletedSteps,
			ProgressPercent: progress.ProgressPercent,
			StartedAt:       progress.StartedAt,
			LastActivityAt:  progress.LastActivityAt,
			CompletedAt:     progress.CompletedAt,
		}
	}

	return &PaginatedSubmissionsResponse{
		Data:       make([]PendingSubmissionResponse, 0),
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminRoadmapService) GetPendingSubmissions(page, limit int, teacherID *uuid.UUID) (*PaginatedSubmissionsResponse, error) {
	offset := (page - 1) * limit
	submissions, total, err := s.repo.GetPendingSubmissions(teacherID, offset, limit)
	if err != nil {
		return nil, errors.New("failed to get pending submissions")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	submissionResponses := make([]PendingSubmissionResponse, len(submissions))
	for i, submission := range submissions {
		student, _ := s.repo.GetStudentProfile(submission.StudentRoadmapProgress.StudentProfileID)

		submissionResponses[i] = PendingSubmissionResponse{
			ID: submission.ID,
			StudentInfo: StudentInfo{
				ID:    student.ID,
				Name:  student.Fullname,
				Email: student.User.Email,
				NIS:   student.NIS,
				Class: student.Class,
			},
			RoadmapInfo: RoadmapSummary{
				ID:          submission.StudentRoadmapProgress.RoadmapID,
				Name:        submission.StudentRoadmapProgress.Roadmap.RoadmapName,
				Description: submission.StudentRoadmapProgress.Roadmap.Description,
				TotalSteps:  submission.StudentRoadmapProgress.TotalSteps,
			},
			StepInfo: StepSummary{
				ID:                submission.RoadmapStep.ID,
				Order:             submission.RoadmapStep.StepOrder,
				Title:             submission.RoadmapStep.Title,
				EstimatedDuration: submission.RoadmapStep.EstimatedDuration,
				DifficultyLevel:   submission.RoadmapStep.DifficultyLevel,
			},
			EvidenceLink:    submission.EvidenceLink,
			EvidenceType:    submission.EvidenceType,
			SubmissionNotes: submission.SubmissionNotes,
			SubmittedAt:     submission.SubmittedAt,
		}
	}

	return &PaginatedSubmissionsResponse{
		Data:       submissionResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminRoadmapService) GetStatistics() (map[string]interface{}, error) {
	return s.repo.GetRoadmapStatistics()
}

func (s *AdminRoadmapService) mapRoadmapToResponse(rm *roadmap.FeatureRoadmap) *RoadmapResponse {
	return &RoadmapResponse{
		ID:              rm.ID,
		ProfilingRoleID: rm.ProfilingRoleID,
		RoadmapName:     rm.RoadmapName,
		Description:     rm.Description,
		Status:          string(rm.Status),
		CreatedBy:       rm.CreatedBy,
		TotalSteps:      len(rm.Steps),
		CreatedAt:       rm.CreatedAt,
		UpdatedAt:       rm.UpdatedAt,
	}
}

func (s *AdminRoadmapService) mapStepToResponse(step *roadmap.RoadmapStep) StepResponse {
	return StepResponse{
		ID:                   step.ID,
		RoadmapID:            step.RoadmapID,
		StepOrder:            step.StepOrder,
		Title:                step.Title,
		Description:          step.Description,
		LearningObjectives:   step.LearningObjectives,
		SubmissionGuidelines: step.SubmissionGuidelines,
		ResourceLinks:        step.ResourceLinks,
		EstimatedDuration:    step.EstimatedDuration,
		DifficultyLevel:      step.DifficultyLevel,
		CreatedAt:            step.CreatedAt,
		UpdatedAt:            step.UpdatedAt,
	}
}
