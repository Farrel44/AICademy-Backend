package teacher_challenge

import (
	"errors"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherChallengeService struct {
	repo *challenge.ChallengeRepository
}

func NewTeacherChallengeService(repo *challenge.ChallengeRepository) *TeacherChallengeService {
	return &TeacherChallengeService{
		repo: repo,
	}
}

func (s *TeacherChallengeService) CreateChallenge(c *fiber.Ctx, req *CreateChallengeRequest) (*challenge.Challenge, error) {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return nil, errors.New("access denied: teacher role required")
	}

	teacherID := claims.UserID

	newChallenge := &challenge.Challenge{
		ThumbnailImage:     req.ThumbnailImage,
		Title:              req.Title,
		Description:        req.Description,
		Deadline:           req.Deadline,
		Prize:              req.Prize,
		MaxParticipants:    req.MaxParticipants,
		CreatedByTeacherID: &teacherID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err = s.repo.CreateChallenge(newChallenge)
	if err != nil {
		return nil, errors.New("failed to create challenge")
	}

	return newChallenge, nil
}

func (s *TeacherChallengeService) UpdateChallenge(c *fiber.Ctx, challengeID uuid.UUID, req *UpdateChallengeRequest) (*challenge.Challenge, error) {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return nil, errors.New("access denied: teacher role required")
	}

	// Get existing challenge and verify ownership
	existingChallenge, err := s.repo.GetTeacherChallengeByID(challengeID, claims.UserID)
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

func (s *TeacherChallengeService) DeleteChallenge(c *fiber.Ctx, challengeID uuid.UUID) error {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return errors.New("access denied: teacher role required")
	}

	// Verify ownership
	_, err = s.repo.GetTeacherChallengeByID(challengeID, claims.UserID)
	if err != nil {
		return errors.New("challenge not found or access denied")
	}

	err = s.repo.DeleteChallenge(challengeID)
	if err != nil {
		return errors.New("failed to delete challenge")
	}

	return nil
}

func (s *TeacherChallengeService) GetMyChallenges(c *fiber.Ctx) ([]challenge.Challenge, error) {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return nil, errors.New("access denied: teacher role required")
	}

	teacherID := claims.UserID

	challenges, err := s.repo.GetChallengesByTeacher(teacherID)
	if err != nil {
		return nil, errors.New("failed to fetch challenges")
	}

	return challenges, nil
}

func (s *TeacherChallengeService) GetChallengeByID(c *fiber.Ctx, challengeID uuid.UUID) (*challenge.Challenge, error) {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return nil, errors.New("access denied: teacher role required")
	}

	teacherID := claims.UserID

	challengeData, err := s.repo.GetTeacherChallengeByID(challengeID, teacherID)
	if err != nil {
		return nil, errors.New("challenge not found or access denied")
	}

	return challengeData, nil
}

func (s *TeacherChallengeService) GetMySubmissions(c *fiber.Ctx, challengeID *uuid.UUID) ([]challenge.Submission, error) {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return nil, errors.New("access denied: teacher role required")
	}

	var submissions []challenge.Submission

	if challengeID != nil {
		// Verify teacher owns this challenge
		_, err = s.GetChallengeByID(c, *challengeID)
		if err != nil {
			return nil, err
		}

		submissions, err = s.repo.GetSubmissionsByChallenge(*challengeID)
	} else {
		// Get submissions for all teacher's challenges
		teacherChallenges, err := s.GetMyChallenges(c)
		if err != nil {
			return nil, err
		}

		for _, ch := range teacherChallenges {
			challengeSubmissions, err := s.repo.GetSubmissionsByChallenge(ch.ID)
			if err != nil {
				continue
			}
			submissions = append(submissions, challengeSubmissions...)
		}
	}

	if err != nil {
		return nil, errors.New("failed to fetch submissions")
	}

	return submissions, nil
}

func (s *TeacherChallengeService) ScoreSubmission(c *fiber.Ctx, req *ScoreSubmissionRequest) error {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return errors.New("access denied: teacher role required")
	}

	// Validate score range
	if req.Points < 1 || req.Points > 100 {
		return errors.New("points must be between 1 and 100")
	}

	err = s.repo.ScoreSubmission(req.SubmissionID, req.Points, claims.UserID, false)
	if err != nil {
		return errors.New("failed to score submission")
	}

	return nil
}

func (s *TeacherChallengeService) GetLeaderboard(c *fiber.Ctx, challengeID *uuid.UUID) ([]challenge.LeaderboardEntry, error) {
	// Verify teacher access
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	if claims.Role != "teacher" {
		return nil, errors.New("access denied: teacher role required")
	}

	var leaderboard []challenge.LeaderboardEntry

	if challengeID != nil {
		// Verify teacher owns this challenge
		_, err = s.GetChallengeByID(c, *challengeID)
		if err != nil {
			return nil, err
		}

		leaderboard, err = s.repo.GetLeaderboard(*challengeID)
	} else {
		// For now, return error as global leaderboard should be admin-only
		return nil, errors.New("challenge_id is required for teacher leaderboard")
	}

	if err != nil {
		return nil, errors.New("failed to fetch leaderboard")
	}

	return leaderboard, nil
}
