package admin_challenge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AdminChallengeService struct {
	repo        *challenge.ChallengeRepository
	redisClient *redis.Client
}

func NewAdminChallengeService(repo *challenge.ChallengeRepository, redisClient *redis.Client) *AdminChallengeService {
	return &AdminChallengeService{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *AdminChallengeService) CreateChallenge(c *fiber.Ctx, req *CreateChallengeRequest) (*challenge.Challenge, error) {
	// Verify admin access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return nil, errors.New("access denied: admin role required")
	}

	adminID := claims.UserID

	newChallenge := &challenge.Challenge{
		Title:            req.Title,
		Description:      req.Description,
		Deadline:         req.Deadline,
		Prize:            req.Prize,
		MaxParticipants:  req.MaxParticipants,
		CreatedByAdminID: &adminID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = s.repo.CreateChallenge(newChallenge)
	if err != nil {
		return nil, errors.New("failed to create challenge")
	}

	return newChallenge, nil
}

func (s *AdminChallengeService) UpdateChallenge(c *fiber.Ctx, challengeID uuid.UUID, req *UpdateChallengeRequest) (*challenge.Challenge, error) {
	// Verify admin access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return nil, errors.New("access denied: admin role required")
	}

	// Get existing challenge and verify ownership
	existingChallenge, err := s.repo.GetAdminChallengeByID(challengeID, claims.UserID)
	if err != nil {
		return nil, errors.New("challenge not found or access denied")
	}

	if req.Title != nil {
		existingChallenge.Title = *req.Title
	}
	if req.Description != nil {
		existingChallenge.Description = *req.Description
	}
	if req.Deadline != nil {
		existingChallenge.Deadline = *req.Deadline
	}
	if req.Prize != nil {
		existingChallenge.Prize = req.Prize
	}
	if req.MaxParticipants != nil {
		existingChallenge.MaxParticipants = *req.MaxParticipants
	}

	existingChallenge.UpdatedAt = time.Now()

	err = s.repo.UpdateChallenge(existingChallenge)
	if err != nil {
		return nil, errors.New("failed to update challenge")
	}

	return existingChallenge, nil
}

func (s *AdminChallengeService) DeleteChallenge(c *fiber.Ctx, challengeID uuid.UUID) error {
	// Verify admin access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return errors.New("access denied: admin role required")
	}

	// Verify ownership
	_, err = s.repo.GetAdminChallengeByID(challengeID, claims.UserID)
	if err != nil {
		return errors.New("challenge not found or access denied")
	}

	err = s.repo.DeleteChallenge(challengeID)
	if err != nil {
		return errors.New("failed to delete challenge")
	}

	return nil
}

func (s *AdminChallengeService) GetAllChallenges(c *fiber.Ctx, page, limit int, search string) (*utils.PaginationResponse, error) {
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return nil, errors.New("akses ditolak: diperlukan role admin")
	}

	// Check and auto announce winners before getting challenges
	s.repo.CheckAndAutoAnnounceWinners()

	// Validate search parameters
	validation, err := utils.ValidateSearchParams(search, page, limit)
	if err != nil {
		return nil, err
	}

	page = validation.Page
	limit = validation.Limit
	search = validation.Query

	cacheKey := fmt.Sprintf("admin_challenges:%d:%d:%s", page, limit, search)

	if cached, err := s.redisClient.Get(context.Background(), cacheKey).Result(); err == nil {
		var result utils.PaginationResponse
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	offset := (page - 1) * limit
	total, err := s.repo.CountChallenges(search)
	if err != nil {
		return nil, errors.New("gagal mengambil jumlah data tantangan")
	}

	// Get challenges data
	challenges, err := s.repo.GetChallengesOptimized(offset, limit, search)
	if err != nil {
		return nil, errors.New("gagal mengambil data tantangan")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := &utils.PaginationResponse{
		Data:       challenges,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, string(resultJSON), time.Minute*5)
	}

	return result, nil
}

func (s *AdminChallengeService) GetChallengeByID(c *fiber.Ctx, challengeID uuid.UUID) (*challenge.Challenge, error) {
	// Verify admin access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return nil, errors.New("access denied: admin role required")
	}

	// Use the new method with auto winner check
	challengeData, err := s.repo.GetChallengeByIDWithWinnerCheck(challengeID)
	if err != nil {
		return nil, errors.New("challenge not found")
	}

	return challengeData, nil
}

func (s *AdminChallengeService) GetSubmissionsByChallengeID(c *fiber.Ctx, challengeID uuid.UUID, page, limit int, search string) (*PaginatedSubmissionsResponse, error) {
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return nil, errors.New("access denied: admin role required")
	}

	// Validate search parameters
	validation, err := utils.ValidateSearchParams(search, page, limit)
	if err != nil {
		return nil, err
	}

	page = validation.Page
	limit = validation.Limit
	search = validation.Query

	cacheKey := fmt.Sprintf("admin_challenge_submissions:%s:%d:%d:%s", challengeID.String(), page, limit, search)

	if cached, err := s.redisClient.Get(context.Background(), cacheKey).Result(); err == nil {
		var result PaginatedSubmissionsResponse
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	offset := (page - 1) * limit
	submissions, total, err := s.repo.GetSubmissionsByChallengeOptimized(challengeID, offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to fetch submissions")
	}

	submissionResponses := make([]SubmissionResponse, len(submissions))
	for i, sub := range submissions {
		teamName := ""
		if sub.Team != nil {
			teamName = sub.Team.TeamName
		}

		repoURL := ""
		if sub.RepoURL != nil {
			repoURL = *sub.RepoURL
		}

		submissionResponses[i] = SubmissionResponse{
			ID:            sub.ID,
			ChallengeID:   sub.ChallengeID,
			ChallengeName: sub.Challenge.Title,
			TeamName:      teamName,
			GitHubURL:     repoURL,
			LiveURL:       sub.DocsURL,
			Description:   sub.Title,
			Points:        sub.Points,
			SubmittedAt:   sub.SubmittedAt,
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := &PaginatedSubmissionsResponse{
		Data:       submissionResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		s.redisClient.Set(context.Background(), cacheKey, string(resultJSON), time.Minute*5)
	}

	return result, nil
}

func (s *AdminChallengeService) ScoreSubmission(c *fiber.Ctx, req *ScoreSubmissionRequest) error {
	// Verify admin access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return errors.New("access denied: admin role required")
	}

	// Validate score range
	if req.Points < 1 || req.Points > 100 {
		return errors.New("points must be between 1 and 100")
	}

	err = s.repo.ScoreSubmission(req.SubmissionID, req.Points, claims.UserID, true)
	if err != nil {
		return errors.New("failed to score submission")
	}

	return nil
}

func (s *AdminChallengeService) GetLeaderboard(c *fiber.Ctx, challengeID *uuid.UUID) ([]challenge.LeaderboardEntry, error) {
	// Verify admin access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "admin" {
		return nil, errors.New("access denied: admin role required")
	}

	var leaderboard []challenge.LeaderboardEntry

	if challengeID != nil {
		// Get leaderboard for specific challenge
		leaderboard, err = s.repo.GetLeaderboard(*challengeID)
	} else {
		// Get global leaderboard
		leaderboard, err = s.repo.GetGlobalLeaderboard()
	}

	if err != nil {
		return nil, errors.New("failed to fetch leaderboard")
	}

	return leaderboard, nil
}
