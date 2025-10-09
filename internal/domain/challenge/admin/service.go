package admin_challenge

import (
    "errors"
    "time"

    "github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
    "github.com/Farrel44/AICademy-Backend/internal/utils"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

type AdminChallengeService struct {
    repo *challenge.ChallengeRepository
}

func NewAdminChallengeService(repo *challenge.ChallengeRepository) *AdminChallengeService {
    return &AdminChallengeService{
        repo: repo,
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
        ThumbnailImage:     req.ThumbnailImage,
        Title:              req.Title,
        Description:        req.Description,
        Deadline:           req.Deadline,
        Prize:              req.Prize,
        MaxParticipants:    req.MaxParticipants,
        CreatedByAdminID:   &adminID,
        CreatedAt:          time.Now(),
        UpdatedAt:          time.Now(),
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

    // Update fields
    if req.ThumbnailImage != nil {
        existingChallenge.ThumbnailImage = *req.ThumbnailImage
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

func (s *AdminChallengeService) GetAllChallenges(c *fiber.Ctx) ([]challenge.Challenge, error) {
    // Verify admin access
    claims, err := utils.GetClaimsFromHeader(c)
    if err != nil {
        return nil, errors.New("unauthorized")
    }
    
    if claims.Role != "admin" {
        return nil, errors.New("access denied: admin role required")
    }

    challenges, err := s.repo.GetAllChallenges()
    if err != nil {
        return nil, errors.New("failed to fetch challenges")
    }

    return challenges, nil
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

    challengeData, err := s.repo.GetChallengeByID(challengeID)
    if err != nil {
        return nil, errors.New("challenge not found")
    }

    return challengeData, nil
}

func (s *AdminChallengeService) GetAllSubmissions(c *fiber.Ctx) ([]challenge.Submission, error) {
    // Verify admin access
    claims, err := utils.GetClaimsFromHeader(c)
    if err != nil {
        return nil, errors.New("unauthorized")
    }
    
    if claims.Role != "admin" {
        return nil, errors.New("access denied: admin role required")
    }

    submissions, err := s.repo.GetAllSubmissions()
    if err != nil {
        return nil, errors.New("failed to fetch submissions")
    }

    return submissions, nil
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