package teacher_challenge

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/challenge"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type TeacherChallengeService struct {
	repo        *challenge.ChallengeRepository
	redisClient *utils.RedisClient
}

func NewTeacherChallengeService(repo *challenge.ChallengeRepository, redisClient *utils.RedisClient) *TeacherChallengeService {
	return &TeacherChallengeService{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (s *TeacherChallengeService) CreateChallenge(userID uuid.UUID, req *CreateChallengeRequest) (*challenge.Challenge, error) {
	// Get teacher profile ID from user ID
	teacherProfileID, err := s.repo.GetTeacherProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("teacher profile not found")
	}

	newChallenge := &challenge.Challenge{
		Title:              req.Title,
		Description:        req.Description,
		Deadline:           req.Deadline,
		Prize:              req.Prize,
		MaxParticipants:    req.MaxParticipants,
		CreatedByTeacherID: teacherProfileID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err = s.repo.CreateChallenge(newChallenge)
	if err != nil {
		return nil, errors.New("failed to create challenge")
	}

	return newChallenge, nil
}

func (s *TeacherChallengeService) UpdateChallenge(userID uuid.UUID, challengeID uuid.UUID, req *UpdateChallengeRequest) (*challenge.Challenge, error) {
	// Get teacher profile ID from user ID
	teacherProfileID, err := s.repo.GetTeacherProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("teacher profile not found")
	}

	// Get existing challenge and verify ownership
	existingChallenge, err := s.repo.GetTeacherChallengeByID(challengeID, *teacherProfileID)
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

func (s *TeacherChallengeService) DeleteChallenge(userID uuid.UUID, challengeID uuid.UUID) error {
	// Get teacher profile ID from user ID
	teacherProfileID, err := s.repo.GetTeacherProfileByUserID(userID)
	if err != nil {
		return errors.New("teacher profile not found")
	}

	// Verify ownership
	_, err = s.repo.GetTeacherChallengeByID(challengeID, *teacherProfileID)
	if err != nil {
		return errors.New("challenge not found or access denied")
	}

	err = s.repo.DeleteChallenge(challengeID)
	if err != nil {
		return errors.New("failed to delete challenge")
	}

	return nil
}

func (s *TeacherChallengeService) GetMyChallenges(userID uuid.UUID, page, limit int, search string) (*utils.PaginationResponse, error) {
	// Get teacher profile ID from user ID
	teacherProfileID, err := s.repo.GetTeacherProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("teacher profile not found")
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

	cacheKey := fmt.Sprintf("teacher_challenges:%s:%d:%d:%s", teacherProfileID.String(), page, limit, search)

	if cached, err := s.redisClient.Get(cacheKey); err == nil {
		var result utils.PaginationResponse
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	offset := (page - 1) * limit
	challenges, total, err := s.repo.GetChallengesByTeacherOptimized(*teacherProfileID, offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to fetch challenges")
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
		s.redisClient.Set(cacheKey, string(resultJSON), time.Minute*5)
	}

	return result, nil
}

func (s *TeacherChallengeService) GetChallengeByID(teacherID uuid.UUID, challengeID uuid.UUID) (*challenge.Challenge, error) {
	// Use the new method with auto winner check instead of direct repository call
	challengeData, err := s.repo.GetChallengeByIDWithWinnerCheck(challengeID)
	if err != nil {
		return nil, errors.New("challenge not found")
	}

	// Verify teacher ownership after getting challenge
	if challengeData.CreatedByTeacherID == nil || *challengeData.CreatedByTeacherID != teacherID {
		return nil, errors.New("access denied: challenge not owned by teacher")
	}

	return challengeData, nil
}

func (s *TeacherChallengeService) GetMySubmissions(teacherID uuid.UUID, page, limit int, search string, challengeID *uuid.UUID) (*utils.PaginationResponse, error) {
	// Check and auto announce winners before getting submissions
	s.repo.CheckAndAutoAnnounceWinners()

	// Validate search parameters
	validation, err := utils.ValidateSearchParams(search, page, limit)
	if err != nil {
		return nil, err
	}

	page = validation.Page
	limit = validation.Limit
	search = validation.Query

	cacheKey := fmt.Sprintf("teacher_challenge_submissions:%s:%d:%d:%s:%v", teacherID.String(), page, limit, search, challengeID)

	if cached, err := s.redisClient.Get(cacheKey); err == nil {
		var result utils.PaginationResponse
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	offset := (page - 1) * limit
	submissions, total, err := s.repo.GetSubmissionsByTeacherOptimized(teacherID, offset, limit, search, challengeID)
	if err != nil {
		return nil, errors.New("failed to fetch submissions")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := &utils.PaginationResponse{
		Data:       submissions,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	if resultJSON, err := json.Marshal(result); err == nil {
		s.redisClient.Set(cacheKey, string(resultJSON), time.Minute*5)
	}

	return result, nil
}

func (s *TeacherChallengeService) ScoreSubmission(teacherID uuid.UUID, req *ScoreSubmissionRequest) error {
	// Validate score range
	if req.Points < 1 || req.Points > 100 {
		return errors.New("points must be between 1 and 100")
	}

	err := s.repo.ScoreSubmission(req.SubmissionID, req.Points, teacherID, false)
	if err != nil {
		return errors.New("failed to score submission")
	}

	return nil
}
