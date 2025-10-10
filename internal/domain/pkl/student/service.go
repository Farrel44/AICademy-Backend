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

func (s *StudentPklService) GetAvailableInternships(page, limit int, search string) (*utils.PaginationResponse, error) {
	// Validate search parameters
	validation, err := utils.ValidateSearchParams(search, page, limit)
	if err != nil {
		return nil, err
	}

	page = validation.Page
	limit = validation.Limit
	search = validation.Query

	// Generate cache keys
	cacheKey := utils.GenerateSearchCacheKey("internships", search, page, limit)
	countKey := utils.GenerateCountCacheKey("internships", search)

	// Try to get from cache first
	if cachedResult, err := utils.GetCachedSearchResult(s.redis, cacheKey); err == nil {
		return &utils.PaginationResponse{
			Data:       cachedResult.Data,
			Total:      cachedResult.TotalCount,
			Page:       cachedResult.Page,
			Limit:      cachedResult.Limit,
			TotalPages: int((cachedResult.TotalCount + int64(limit) - 1) / int64(limit)),
		}, nil
	}

	offset := (page - 1) * limit

	// Get cached count first to avoid expensive COUNT query
	var total int64
	if cachedCount, err := utils.GetCachedCount(s.redis, countKey); err == nil {
		total = cachedCount
	} else {
		// Only count if not in cache
		total, err = s.repo.CountInternships(search)
		if err != nil {
			return nil, errors.New("gagal mengambil jumlah data magang")
		}
		// Cache the count
		utils.CacheCount(s.redis, countKey, total)
	}

	// Get internships data
	internships, err := s.repo.GetInternshipsOptimized(offset, limit, search)
	if err != nil {
		return nil, errors.New("gagal mengambil data magang")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Cache the results
	utils.CacheSearchResult(s.redis, cacheKey, internships, total, page, limit)

	return &utils.PaginationResponse{
		Data:       internships,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
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
