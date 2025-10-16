package experience

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service interface {
	CreateExperience(studentProfileID uuid.UUID, req *CreateExperienceRequest) (*ExperienceResponse, error)
	GetExperienceByID(id, studentProfileID uuid.UUID) (*ExperienceResponse, error)
	GetExperiencesByStudentID(studentProfileID uuid.UUID, page, limit int) (*ExperienceListResponse, error)
	UpdateExperience(id, studentProfileID uuid.UUID, req *UpdateExperienceRequest) (*ExperienceResponse, error)
	DeleteExperience(id, studentProfileID uuid.UUID) error
}

type service struct {
	repo         Repository
	redis        *utils.RedisClient
	cacheManager *utils.CacheManager
}

func NewService(repo Repository, redis *utils.RedisClient) Service {
	return &service{
		repo:         repo,
		redis:        redis,
		cacheManager: utils.NewCacheManager(redis),
	}
}

func (s *service) CreateExperience(studentProfileID uuid.UUID, req *CreateExperienceRequest) (*ExperienceResponse, error) {
	// Validate business rules
	if err := s.validateExperienceDates(req.StartDate, req.EndDate, req.IsCurrent); err != nil {
		return nil, err
	}

	experience := &Experience{
		ID:               uuid.New(),
		StudentProfileID: studentProfileID,
		CompanyName:      req.CompanyName,
		Position:         req.Position,
		Department:       req.Department,
		EmploymentType:   req.EmploymentType,
		Location:         req.Location,
		LocationType:     req.LocationType,
		Description:      req.Description,
		Responsibilities: req.Responsibilities,
		Achievements:     req.Achievements,
		Skills:           req.Skills,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		IsCurrent:        req.IsCurrent,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// If marked as current, end date should be nil
	if req.IsCurrent {
		experience.EndDate = nil
	}

	if err := s.repo.CreateExperience(experience); err != nil {
		return nil, fmt.Errorf("failed to create experience: %w", err)
	}

	s.invalidateExperienceCache(studentProfileID)
	return s.experienceToResponse(experience), nil
}

func (s *service) GetExperienceByID(id, studentProfileID uuid.UUID) (*ExperienceResponse, error) {
	cacheKey := fmt.Sprintf("experience:%s", id.String())

	var experience Experience
	if err := s.redis.GetJSON(cacheKey, &experience); err == nil {
		if experience.StudentProfileID == studentProfileID {
			return s.experienceToResponse(&experience), nil
		}
	}

	experience2, err := s.repo.GetExperienceByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("experience not found")
		}
		return nil, fmt.Errorf("failed to get experience: %w", err)
	}

	if experience2.StudentProfileID != studentProfileID {
		return nil, fmt.Errorf("experience not found")
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, *experience2, "medium")
	return s.experienceToResponse(experience2), nil
}

func (s *service) GetExperiencesByStudentID(studentProfileID uuid.UUID, page, limit int) (*ExperienceListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("experiences:%s:page:%d:limit:%d", studentProfileID.String(), page, limit)

	var cachedResponse ExperienceListResponse
	if err := s.redis.GetJSON(cacheKey, &cachedResponse); err == nil {
		return &cachedResponse, nil
	}

	experiences, total, err := s.repo.GetExperiencesByStudentID(studentProfileID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiences: %w", err)
	}

	responses := make([]ExperienceResponse, len(experiences))
	for i, exp := range experiences {
		responses[i] = *s.experienceToResponse(&exp)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	response := &ExperienceListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	s.cacheManager.SetWithSmartTTL(cacheKey, *response, "medium")
	return response, nil
}

func (s *service) UpdateExperience(id, studentProfileID uuid.UUID, req *UpdateExperienceRequest) (*ExperienceResponse, error) {
	// Check if experience exists and belongs to student
	exists, err := s.repo.ExistsForStudent(id, studentProfileID)
	if err != nil {
		return nil, fmt.Errorf("failed to check experience: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("experience not found")
	}

	// Get existing experience
	experience, err := s.repo.GetExperienceByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get experience: %w", err)
	}

	// Update fields if provided
	s.updateExperienceFields(experience, req)

	// Validate business rules with potentially updated dates
	if err := s.validateExperienceDates(experience.StartDate, experience.EndDate, experience.IsCurrent); err != nil {
		return nil, err
	}

	// If marked as current, end date should be nil
	if experience.IsCurrent {
		experience.EndDate = nil
	}

	experience.UpdatedAt = time.Now()

	if err := s.repo.UpdateExperience(experience); err != nil {
		return nil, fmt.Errorf("failed to update experience: %w", err)
	}

	s.invalidateExperienceCache(studentProfileID)
	return s.experienceToResponse(experience), nil
}

func (s *service) DeleteExperience(id, studentProfileID uuid.UUID) error {
	// Check if experience exists and belongs to student
	exists, err := s.repo.ExistsForStudent(id, studentProfileID)
	if err != nil {
		return fmt.Errorf("failed to check experience: %w", err)
	}
	if !exists {
		return fmt.Errorf("experience not found")
	}

	if err := s.repo.DeleteExperience(id); err != nil {
		return fmt.Errorf("failed to delete experience: %w", err)
	}

	s.invalidateExperienceCache(studentProfileID)
	return nil
}

func (s *service) updateExperienceFields(experience *Experience, req *UpdateExperienceRequest) {
	if req.CompanyName != nil {
		experience.CompanyName = *req.CompanyName
	}
	if req.Position != nil {
		experience.Position = *req.Position
	}
	if req.Department != nil {
		experience.Department = *req.Department
	}
	if req.EmploymentType != nil {
		experience.EmploymentType = *req.EmploymentType
	}
	if req.Location != nil {
		experience.Location = *req.Location
	}
	if req.LocationType != nil {
		experience.LocationType = *req.LocationType
	}
	if req.Description != nil {
		experience.Description = *req.Description
	}
	if req.Responsibilities != nil {
		experience.Responsibilities = *req.Responsibilities
	}
	if req.Achievements != nil {
		experience.Achievements = *req.Achievements
	}
	if req.Skills != nil {
		experience.Skills = *req.Skills
	}
	if req.StartDate != nil {
		experience.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		experience.EndDate = req.EndDate
	}
	if req.IsCurrent != nil {
		experience.IsCurrent = *req.IsCurrent
	}
}

func (s *service) validateExperienceDates(startDate time.Time, endDate *time.Time, isCurrent bool) error {
	now := time.Now()

	// Start date cannot be in the future
	if startDate.After(now) {
		return fmt.Errorf("start date cannot be in the future")
	}

	// If not current, end date is required
	if !isCurrent && endDate == nil {
		return fmt.Errorf("end date is required when experience is not current")
	}

	// If end date is provided, it cannot be before start date
	if endDate != nil {
		if endDate.Before(startDate) {
			return fmt.Errorf("end date cannot be before start date")
		}
		// End date cannot be in the future
		if endDate.After(now) {
			return fmt.Errorf("end date cannot be in the future")
		}
	}

	// If marked as current, end date should not be provided
	if isCurrent && endDate != nil {
		return fmt.Errorf("end date should not be provided for current experience")
	}

	return nil
}

func (s *service) experienceToResponse(experience *Experience) *ExperienceResponse {
	return &ExperienceResponse{
		ID:               experience.ID,
		StudentProfileID: experience.StudentProfileID,
		CompanyName:      experience.CompanyName,
		Position:         experience.Position,
		Department:       experience.Department,
		EmploymentType:   experience.EmploymentType,
		Location:         experience.Location,
		LocationType:     experience.LocationType,
		Description:      experience.Description,
		Responsibilities: experience.Responsibilities,
		Achievements:     experience.Achievements,
		Skills:           experience.Skills,
		StartDate:        experience.StartDate,
		EndDate:          experience.EndDate,
		IsCurrent:        experience.IsCurrent,
		CreatedAt:        experience.CreatedAt,
		UpdatedAt:        experience.UpdatedAt,
	}
}

func (s *service) invalidateExperienceCache(studentProfileID uuid.UUID) {
	s.redis.Delete("experience:statistics")
	s.cacheManager.InvalidateByPattern("experiences:*")
	s.cacheManager.InvalidateByPattern(fmt.Sprintf("experiences:%s:*", studentProfileID.String()))
}
