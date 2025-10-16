package admin

import (
	"time"

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

	challenges, err := s.repo.GetAdminChallenges()
	if err != nil {
		return nil, err
	}

	roadmapSubmissions, err := s.repo.GetAdminRoadmapSubmissions()
	if err != nil {
		return nil, err
	}

	adminChallenges := make([]AdminChallenge, len(challenges))
	for i, challenge := range challenges {
		adminChallenges[i] = AdminChallenge{
			ID:                  challenge.ID,
			Title:               challenge.Title,
			Description:         challenge.Description,
			Deadline:            challenge.Deadline,
			Prize:               challenge.Prize,
			MaxParticipants:     challenge.MaxParticipants,
			CurrentParticipants: challenge.CurrentParticipants,
			IsActive:            challenge.IsActive,
			CreatedAt:           challenge.CreatedAt,
		}
	}

	adminRoadmapSubmissions := make([]RoadmapStepSubmission, len(roadmapSubmissions))
	for i, submission := range roadmapSubmissions {
		evidenceLink := ""
		if submission.EvidenceLink != nil {
			evidenceLink = *submission.EvidenceLink
		}

		evidenceType := ""
		if submission.EvidenceType != nil {
			evidenceType = *submission.EvidenceType
		}

		submissionNotes := ""
		if submission.SubmissionNotes != nil {
			submissionNotes = *submission.SubmissionNotes
		}

		submittedAt := time.Time{}
		if submission.SubmittedAt != nil {
			submittedAt = *submission.SubmittedAt
		}

		adminRoadmapSubmissions[i] = RoadmapStepSubmission{
			ID:               submission.ID,
			RoadmapStepID:    submission.RoadmapStepID,
			StepTitle:        submission.StepTitle,
			StudentProfileID: submission.StudentProfileID,
			StudentName:      submission.StudentName,
			EvidenceLink:     evidenceLink,
			EvidenceType:     evidenceType,
			SubmissionNotes:  submissionNotes,
			ValidationNotes:  submission.ValidationNotes,
			ValidationScore:  submission.ValidationScore,
			Status:           submission.Status,
			SubmittedAt:      submittedAt,
			IsValidated:      submission.IsValidated,
		}
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
		Challenges:         adminChallenges,
		RoadmapSubmissions: adminRoadmapSubmissions,
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, dashboardData, "medium")
	return dashboardData, nil
}
