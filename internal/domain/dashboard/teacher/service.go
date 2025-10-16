package teacher

import (
	"fmt"

	"github.com/Farrel44/AICademy-Backend/internal/domain/dashboard"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type Service interface {
	GetTeacherDashboard(userID uuid.UUID) (*TeacherDashboardData, error)
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

func (s *service) GetTeacherDashboard(userID uuid.UUID) (*TeacherDashboardData, error) {
	cacheKey := fmt.Sprintf("teacher_dashboard:%s", userID.String())

	var cachedData TeacherDashboardData
	if err := s.redis.GetJSON(cacheKey, &cachedData); err == nil {
		return &cachedData, nil
	}

	teacherProfileID, err := s.repo.GetTeacherProfileIDByUserID(userID)
	if err != nil {
		return nil, err
	}

	challengeStats, err := s.repo.GetTeacherChallengeStats(teacherProfileID)
	if err != nil {
		return nil, err
	}

	submissionStats, err := s.repo.GetTeacherSubmissionStats(teacherProfileID)
	if err != nil {
		return nil, err
	}

	roadmapStats, err := s.repo.GetTeacherRoadmapStats(teacherProfileID)
	if err != nil {
		return nil, err
	}

	challenges, err := s.repo.GetTeacherChallenges(teacherProfileID)
	if err != nil {
		return nil, err
	}

	challengeSubmissions, err := s.repo.GetTeacherChallengeSubmissions(teacherProfileID)
	if err != nil {
		return nil, err
	}

	roadmapSubmissions, err := s.repo.GetTeacherRoadmapSubmissions(teacherProfileID)
	if err != nil {
		return nil, err
	}

	// Convert repository types to service types
	teacherChallenges := make([]TeacherChallenge, len(challenges))
	for i, c := range challenges {
		teacherChallenges[i] = TeacherChallenge{
			ID:                  c.ID,
			Title:               c.Title,
			Description:         c.Description,
			Deadline:            c.Deadline,
			Prize:               c.Prize,
			MaxParticipants:     c.MaxParticipants,
			CurrentParticipants: c.CurrentParticipants,
			IsActive:            c.IsActive,
			CreatedAt:           c.CreatedAt,
		}
	}

	teacherSubmissions := make([]ChallengeSubmissionItem, len(challengeSubmissions))
	for i, s := range challengeSubmissions {
		teacherSubmissions[i] = ChallengeSubmissionItem{
			ID:               s.ID,
			ChallengeID:      s.ChallengeID,
			ChallengeTitle:   s.ChallengeTitle,
			Title:            s.Title,
			TeamID:           s.TeamID,
			TeamName:         s.TeamName,
			StudentProfileID: s.StudentProfileID,
			StudentName:      s.StudentName,
			ImageURL:         s.ImageURL,
			RepoURL:          s.RepoURL,
			DocsURL:          s.DocsURL,
			SubmittedAt:      s.SubmittedAt,
			Points:           s.Points,
			IsScored:         s.IsScored,
		}
	}

	teacherRoadmapSubmissions := make([]RoadmapStepSubmission, len(roadmapSubmissions))
	for i, r := range roadmapSubmissions {
		teacherRoadmapSubmissions[i] = RoadmapStepSubmission{
			ID:               r.ID,
			RoadmapStepID:    r.RoadmapStepID,
			StepTitle:        r.StepTitle,
			StudentProfileID: r.StudentProfileID,
			StudentName:      r.StudentName,
			EvidenceLink:     r.EvidenceLink,
			EvidenceType:     r.EvidenceType,
			SubmissionNotes:  r.SubmissionNotes,
			ValidationNotes:  r.ValidationNotes,
			ValidationScore:  r.ValidationScore,
			Status:           r.Status,
			SubmittedAt:      r.SubmittedAt,
			IsValidated:      r.IsValidated,
		}
	}

	dashboardData := &TeacherDashboardData{
		ChallengeStats: TeacherChallengeStats{
			TotalChallenges:     challengeStats.TotalChallenges,
			ActiveChallenges:    challengeStats.ActiveChallenges,
			CompletedChallenges: challengeStats.CompletedChallenges,
		},
		SubmissionStats: TeacherSubmissionStats{
			TotalSubmissions:   submissionStats.TotalSubmissions,
			ScoredSubmissions:  submissionStats.ScoredSubmissions,
			PendingSubmissions: submissionStats.PendingSubmissions,
		},
		RoadmapStats: TeacherRoadmapStats{
			TotalStepSubmissions: roadmapStats.TotalStepSubmissions,
			ValidatedSubmissions: roadmapStats.ValidatedSubmissions,
			PendingValidations:   roadmapStats.PendingValidations,
		},
		Challenges:           teacherChallenges,
		ChallengeSubmissions: teacherSubmissions,
		RoadmapSubmissions:   teacherRoadmapSubmissions,
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, dashboardData, "short")
	return dashboardData, nil
}
