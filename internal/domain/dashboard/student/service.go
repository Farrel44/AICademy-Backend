package student

import (
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/dashboard"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type Service interface {
	GetStudentDashboard(userID uuid.UUID) (*StudentDashboardData, error)
}

type service struct {
	repo         dashboard.Repository
	redis        *utils.RedisClient
	cacheManager *utils.CacheManager
}

func NewService(repo dashboard.Repository, redis *utils.RedisClient) Service {
	return &service{
		repo:         repo,
		redis:        redis,
		cacheManager: utils.NewCacheManager(redis),
	}
}

func (s *service) GetStudentDashboard(userID uuid.UUID) (*StudentDashboardData, error) {
	cacheKey := fmt.Sprintf("student_dashboard:%s", userID.String())

	var cachedData StudentDashboardData
	if err := s.redis.GetJSON(cacheKey, &cachedData); err == nil {
		return &cachedData, nil
	}

	studentProfileID, err := s.repo.GetStudentProfileIDByUserID(userID)
	if err != nil {
		return nil, err
	}

	roadmapData, err := s.repo.GetStudentAssignedRoadmapWithProgress(studentProfileID)
	if err != nil {
		return nil, err
	}

	challengeData, err := s.repo.GetStudentChallenges(studentProfileID)
	if err != nil {
		return nil, err
	}

	summaryData, err := s.repo.GetStudentSummary(studentProfileID)
	if err != nil {
		return nil, err
	}

	var dashboardRoadmap *DashboardRoadmapResponse
	if roadmapData != nil {
		dashboardRoadmap = &DashboardRoadmapResponse{
			ID:                roadmapData.ID,
			RoadmapName:       roadmapData.RoadmapName,
			Description:       roadmapData.Description,
			RoleName:          roadmapData.RoleName,
			TotalSteps:        roadmapData.TotalSteps,
			EstimatedDuration: roadmapData.EstimatedDuration,
			DifficultyLevel:   roadmapData.DifficultyLevel,
		}

		if roadmapData.ProgressID != nil {
			dashboardRoadmap.Progress = &DashboardRoadmapProgress{
				ID:              *roadmapData.ProgressID,
				TotalSteps:      roadmapData.TotalSteps,
				CompletedSteps:  roadmapData.CompletedSteps,
				ProgressPercent: roadmapData.ProgressPercent,
				IsFinished:      roadmapData.ProgressCompletedAt != nil,
			}

			if roadmapData.ProgressStartedAt != nil {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", *roadmapData.ProgressStartedAt); err == nil {
					dashboardRoadmap.Progress.StartedAt = &parsedTime
				}
			}
			if roadmapData.ProgressUpdatedAt != nil {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", *roadmapData.ProgressUpdatedAt); err == nil {
					dashboardRoadmap.Progress.LastActivityAt = &parsedTime
				}
			}
			if roadmapData.ProgressCompletedAt != nil {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", *roadmapData.ProgressCompletedAt); err == nil {
					dashboardRoadmap.Progress.CompletedAt = &parsedTime
				}
			}
		}

		dashboardRoadmap.Steps = make([]DashboardStepResponse, len(roadmapData.Steps))
		for i, step := range roadmapData.Steps {
			dashboardRoadmap.Steps[i] = DashboardStepResponse{
				ID:                   step.ID,
				StepOrder:            step.StepOrder,
				Title:                step.Title,
				Description:          step.Description,
				LearningObjectives:   step.LearningObjectives,
				SubmissionGuidelines: step.SubmissionGuidelines,
				ResourceLinks:        step.ResourceLinks,
				EstimatedDuration:    step.EstimatedDuration,
				DifficultyLevel:      step.DifficultyLevel,
				Status:               step.Status,
				EvidenceLink:         step.EvidenceLink,
				EvidenceType:         step.EvidenceType,
				SubmissionNotes:      step.SubmissionNotes,
				ValidationNotes:      step.ValidationNotes,
				ValidationScore:      step.ValidationScore,
			}

			if step.StartedAt != nil {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", *step.StartedAt); err == nil {
					dashboardRoadmap.Steps[i].StartedAt = &parsedTime
				}
			}
			if step.SubmittedAt != nil {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", *step.SubmittedAt); err == nil {
					dashboardRoadmap.Steps[i].SubmittedAt = &parsedTime
				}
			}
			if step.CompletedAt != nil {
				if parsedTime, err := time.Parse("2006-01-02 15:04:05", *step.CompletedAt); err == nil {
					dashboardRoadmap.Steps[i].CompletedAt = &parsedTime
				}
			}

			dashboardRoadmap.Steps[i].CanStart = step.Status == "unlocked"
			dashboardRoadmap.Steps[i].CanSubmit = step.Status == "in_progress"
			dashboardRoadmap.Steps[i].IsLocked = step.Status == "locked"
		}
	}

	challenges := make([]StudentChallengeItem, len(challengeData))
	for i, item := range challengeData {
		var deadline time.Time
		var registeredAt time.Time

		if parsedDeadline, err := time.Parse("2006-01-02 15:04:05", item.Deadline); err == nil {
			deadline = parsedDeadline
		}

		if parsedRegistered, err := time.Parse("2006-01-02 15:04:05", item.RegisteredAt); err == nil {
			registeredAt = parsedRegistered
		}

		challenges[i] = StudentChallengeItem{
			ChallengeID:    item.ChallengeID,
			ChallengeTitle: item.ChallengeTitle,
			TeamID:         item.TeamID,
			TeamName:       item.TeamName,
			SubmissionID:   item.SubmissionID,
			Points:         item.Points,
			Deadline:       deadline,
			IsActive:       item.IsActive,
			RegisteredAt:   registeredAt,
		}
	}

	dashboardData := &StudentDashboardData{
		Roadmap:    dashboardRoadmap,
		Challenges: challenges,
		Summary: StudentSummary{
			TotalRoadmaps:       summaryData.TotalRoadmaps,
			CompletedSteps:      summaryData.CompletedSteps,
			TotalSteps:          summaryData.TotalSteps,
			ActiveChallenges:    summaryData.ActiveChallenges,
			CompletedChallenges: summaryData.CompletedChallenges,
			TotalPoints:         summaryData.TotalPoints,
		},
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, dashboardData, "short")
	return dashboardData, nil
}
