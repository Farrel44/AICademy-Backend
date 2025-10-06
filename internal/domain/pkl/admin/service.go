package pkl

import (
	"errors"
	"fmt"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/google/uuid"
)

type AdminPklService struct {
	repo  *pkl.PklRepository
	redis *utils.RedisClient
}

func NewAdminPklService(repo *pkl.PklRepository, redis *utils.RedisClient) *AdminPklService {
	return &AdminPklService{repo: repo, redis: redis}
}

func (s *AdminPklService) CheckRateLimit(userID string, limit int, window time.Duration) (allowed bool, remaining int, resetTime time.Time, err error) {
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

func (s *AdminPklService) CreateInternshipPosition(req *CreateInternshipRequest) (*pkl.Internship, error) {
	existingCompany, _ := s.repo.GetCompanyByID(req.CompanyID)
	if existingCompany == nil {
		return nil, errors.New("Companies Not Found")
	}

	newInternshipPosition := pkl.Internship{
		CompanyProfileID: existingCompany.ID,
		Title:            req.Title,
		Description:      req.Description,
		Type:             req.Type,
		Deadline:         req.Deadline,
		PostedAt:         time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := s.repo.CreateInternshipPosition(&newInternshipPosition)
	if err != nil {
		return nil, errors.New("failed to create intership position")
	}

	s.invalidateAllPklCache()
	return &newInternshipPosition, nil
}

func (s *AdminPklService) invalidateInternshipCache(internshipID uuid.UUID) {
	internshipKey := fmt.Sprintf("internship:%s", internshipID.String())
	s.redis.Delete(internshipKey)

	s.redis.Delete("internship_statistics")

	s.invalidateInternshipsListCache()
}

func (s *AdminPklService) invalidateInternshipsListCache() {
	commonKeys := []string{
		"internships:page:1:limit:10:search:",
		"internships:page:2:limit:10:search:",
		"internships:page:1:limit:20:search:",
		"internships:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminPklService) invalidateApplicationCache(applicationID uuid.UUID) {
	applicationKey := fmt.Sprintf("application:%s", applicationID.String())
	s.redis.Delete(applicationKey)

	s.redis.Delete("internship_statistics")

	s.invalidateApplicationsListCache()
}

func (s *AdminPklService) invalidateApplicationsListCache() {
	commonKeys := []string{
		"applications:page:1:limit:10:search:",
		"applications:page:2:limit:10:search:",
		"applications:page:1:limit:20:search:",
		"applications:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminPklService) invalidateReviewCache(reviewID uuid.UUID) {
	reviewKey := fmt.Sprintf("review:%s", reviewID.String())
	s.redis.Delete(reviewKey)

	s.redis.Delete("internship_statistics")

	s.invalidateReviewsListCache()
}

func (s *AdminPklService) invalidateReviewsListCache() {
	commonKeys := []string{
		"reviews:page:1:limit:10:search:",
		"reviews:page:2:limit:10:search:",
		"reviews:page:1:limit:20:search:",
		"reviews:page:2:limit:20:search:",
	}

	for _, key := range commonKeys {
		s.redis.Delete(key)
	}
}

func (s *AdminPklService) invalidateAllPklCache() {
	s.redis.Delete("internship_statistics")
	s.invalidateInternshipsListCache()
	s.invalidateApplicationsListCache()
	s.invalidateReviewsListCache()
}

func (s *AdminPklService) GetInternshipPositions(page, limit int, search string) (*PaginatedInternshipResponse, error) {
	internships, total, err := s.repo.GetAllInternships(page, limit, search)
	if err != nil {
		return nil, errors.New("failed to get internship positions")
	}

	// Fix type conversion from int64 to int
	totalPages := int(total+int64(limit)-1) / limit

	return &PaginatedInternshipResponse{
		Data:       internships,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminPklService) GetInternshipByID(id uuid.UUID) (*pkl.Internship, error) {
	internship, err := s.repo.GetInternshipByID(id)
	if err != nil {
		return nil, errors.New("internship position not found")
	}
	return internship, nil
}

func (s *AdminPklService) UpdateInternshipPosition(id uuid.UUID, req *UpdateInternshipRequest) error {
	existingInternship, err := s.repo.GetInternshipByID(id)
	if err != nil {
		return errors.New("internship position not found")
	}

	if req.CompanyID != nil {
		existingCompany, _ := s.repo.GetCompanyByID(*req.CompanyID)
		if existingCompany == nil {
			return errors.New("company not found")
		}
		existingInternship.CompanyProfileID = *req.CompanyID
	}

	if req.Title != nil {
		existingInternship.Title = *req.Title
	}

	if req.Description != nil {
		existingInternship.Description = *req.Description
	}

	if req.Type != nil {
		existingInternship.Type = *req.Type
	}

	if req.Deadline != nil {
		existingInternship.Deadline = req.Deadline
	}

	existingInternship.UpdatedAt = time.Now()

	err = s.repo.UpdateInternshipPosition(existingInternship)
	if err != nil {
		return errors.New("failed to update internship position")
	}

	s.invalidateInternshipCache(id)
	return nil
}

func (s *AdminPklService) DeleteInternshipPosition(id uuid.UUID) error {
	existingInternship, err := s.repo.GetInternshipByID(id)
	if err != nil {
		return errors.New("internship position not found")
	}

	err = s.repo.DeleteInternshipPosition(existingInternship.ID)
	if err != nil {
		return errors.New("failed to delete internship position")
	}

	s.invalidateInternshipCache(id)
	return nil
}
