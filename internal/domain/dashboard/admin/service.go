package admin

import (
	"github.com/Farrel44/AICademy-Backend/internal/domain/dashboard"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
)

type Service interface {
	GetAdminDashboard() (*AdminDashboardData, error)
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

func (s *service) GetAdminDashboard() (*AdminDashboardData, error) {
	cacheKey := "admin_dashboard"

	var cachedData AdminDashboardData
	if err := s.redis.GetJSON(cacheKey, &cachedData); err == nil {
		return &cachedData, nil
	}

	totals, err := s.repo.GetAdminTotals()
	if err != nil {
		return nil, err
	}

	studentStats, err := s.repo.GetStudentStatistics()
	if err != nil {
		return nil, err
	}

	dashboardData := &AdminDashboardData{
		Totals: AdminTotals{
			TotalUsers:     totals.TotalUsers,
			TotalStudents:  totals.TotalStudents,
			TotalTeachers:  totals.TotalTeachers,
			TotalCompanies: totals.TotalCompanies,
		},
		StudentStats: StudentStatistics{
			TotalTKJ:  studentStats.TotalTKJ,
			TotalTJA:  studentStats.TotalTJA,
			TotalPPLG: studentStats.TotalPPLG,
			TotalRPL:  studentStats.TotalRPL,
		},
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, dashboardData, "medium")
	return dashboardData, nil
}
