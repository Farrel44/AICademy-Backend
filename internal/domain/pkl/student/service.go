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

type StudentPklService struct {
	repo  *pkl.PklRepository
	redis *utils.RedisClient
}

func NewStudentPklService(repo *pkl.PklRepository, redis *utils.RedisClient) *StudentPklService {
	return &StudentPklService{
		repo:  repo,
		redis: redis,
	}
}

func (s *StudentPklService) CheckRateLimit(userID string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
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

func (s *StudentPklService) ApplyStudentInternshipPosition(c *fiber.Ctx, internshipId uuid.UUID) (*pkl.InternshipApplication, error) {
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
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	internship, err := s.repo.GetInternshipByID(internshipId)
	if err != nil {
		return nil, errors.New("invalid internship id")
	}

	if internship.Deadline != nil && time.Now().After(*internship.Deadline) {
		return nil, errors.New("the internship application period has ended")
	}

	exist, err := s.repo.HasExistingApplication(internshipId, user.StudentProfile.ID)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("you have already applied to this internship")
	}

	app := &pkl.InternshipApplication{
		InternshipID:     internship.ID,
		StudentProfileID: &user.StudentProfile.ID,
		AppliedAt:        time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.ApplyInternshipByID(app); err != nil {
		return nil, err
	}

	return app, nil
}
