package pkl

import (
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AlumniPklService struct {
	repo  *pkl.PklRepository
	redis *utils.RedisClient
}

func NewAlumniPklService(repo *pkl.PklRepository, redis *utils.RedisClient) *AlumniPklService {
	return &AlumniPklService{
		repo:  repo,
		redis: redis,
	}
}

func (s *AlumniPklService) CheckRateLimit(userID string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
	key := fmt.Sprintf("rate_limit:%s", userID)
	count, err := s.redis.Incr(key)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	if count == 1 {
		s.redis.SetExpire(key, window)
	}

	ttl, _ := s.redis.GetTTL(key)
	resetTime = time.Now().Add(ttl)
	remaining = limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	allowed = count <= int64(limit)
	return allowed, remaining, resetTime, nil
}

func (s *AlumniPklService) ApplyAlumniInternshipPosition(c *fiber.Ctx, internshipId uuid.UUID) (*pkl.InternshipApplication, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("Failed to get user data")
	}

	fmt.Print("internship id")
	fmt.Print(internshipId)

	var req ApplyInternshipRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, errors.New("invalid request body")
	}
	if err := utils.ValidateStruct(req); err != nil {
		return nil, err
	}

	user, err := s.repo.GetAlumniUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	internship, err := s.repo.GetInternshipByID(internshipId)
	if err != nil {
		return nil, errors.New("invalid internship id")
	}

	// Alumni can only apply for Freelance and Job positions
	if internship.Type == pkl.InternshipTypePKL {
		return nil, errors.New("alumni cannot apply for PKL positions")
	}

	if internship.Deadline != nil && time.Now().After(*internship.Deadline) {
		return nil, errors.New("the internship application period has ended")
	}

	exist, err := s.repo.HasExistingAlumniApplication(internshipId, user.AlumniProfile.ID)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("you have already applied to this internship")
	}

	app := &pkl.InternshipApplication{
		InternshipID:    internship.ID,
		AlumniProfileID: &user.AlumniProfile.ID,
		AppliedAt:       time.Now(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.ApplyInternshipByID(app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *AlumniPklService) GetAvailablePositions(c *fiber.Ctx, offset, limit int, search string) ([]pkl.Internship, int64, error) {
	// Only return Freelance and Job positions for alumni
	return s.repo.GetAlumniInternships(offset, limit, search)
}

func (s *AlumniPklService) GetAlumniApplications(c *fiber.Ctx, offset, limit int) ([]pkl.InternshipApplication, int64, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, 0, errors.New("failed to get user data")
	}

	user, err := s.repo.GetAlumniUserByID(userID)
	if err != nil {
		return nil, 0, errors.New("failed to get user data")
	}

	return s.repo.GetAlumniApplicationsByProfileID(user.AlumniProfile.ID, offset, limit)
}

func (s *AlumniPklService) GetApplicationByID(c *fiber.Ctx, applicationID uuid.UUID) (*pkl.InternshipApplication, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	user, err := s.repo.GetAlumniUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	application, err := s.repo.GetSubmissionByID(applicationID)
	if err != nil {
		return nil, errors.New("application not found")
	}

	if application.AlumniProfileID == nil || *application.AlumniProfileID != user.AlumniProfile.ID {
		return nil, errors.New("unauthorized access to application")
	}

	return application, nil
}
