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
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, dashboardData, "short")
	return dashboardData, nil
}
