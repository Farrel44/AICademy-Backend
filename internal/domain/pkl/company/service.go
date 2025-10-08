package pkl

import (
	"errors"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/pkl"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CompanyPklService struct {
	repo  *pkl.PklRepository
	redis *utils.RedisClient
}

func NewCompanyPklService(repo *pkl.PklRepository, redis *utils.RedisClient) *CompanyPklService {
	return &CompanyPklService{
		repo:  repo,
		redis: redis,
	}
}

func (s *CompanyPklService) GetCompanyInternships(c *fiber.Ctx, offset, limit int, search string) ([]pkl.Internship, int64, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, 0, errors.New("failed to get user data")
	}

	user, err := s.repo.GetCompanyUserByID(userID)
	if err != nil {
		return nil, 0, errors.New("failed to get company data")
	}

	return s.repo.GetInternshipsByCompanyID(user.CompanyProfile.ID, offset, limit, search)
}

func (s *CompanyPklService) GetInternshipApplications(c *fiber.Ctx, internshipID uuid.UUID) ([]pkl.InternshipApplication, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	user, err := s.repo.GetCompanyUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get company data")
	}

	// Verify that the internship belongs to this company
	internship, err := s.repo.GetInternshipByID(internshipID)
	if err != nil {
		return nil, errors.New("internship not found")
	}

	if internship.CompanyProfileID != user.CompanyProfile.ID {
		return nil, errors.New("unauthorized: internship does not belong to this company")
	}

	return s.repo.GetSubmissionsByInternshipID(internshipID)
}

func (s *CompanyPklService) UpdateApplicationStatus(c *fiber.Ctx, applicationID uuid.UUID, status pkl.ApplicationStatus) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return errors.New("failed to get user data")
	}

	user, err := s.repo.GetCompanyUserByID(userID)
	if err != nil {
		return errors.New("failed to get company data")
	}

	// Get the application and verify it belongs to this company's internship
	application, err := s.repo.GetSubmissionByID(applicationID)
	if err != nil {
		return errors.New("application not found")
	}

	// Verify that the internship belongs to this company
	internship, err := s.repo.GetInternshipByID(application.InternshipID)
	if err != nil {
		return errors.New("internship not found")
	}

	if internship.CompanyProfileID != user.CompanyProfile.ID {
		return errors.New("unauthorized: cannot modify applications for internships not owned by this company")
	}

	// Only allow approved or rejected status from companies
	if status != pkl.ApplicationStatusApproved && status != pkl.ApplicationStatusRejected {
		return errors.New("invalid status: companies can only approve or reject applications")
	}

	return s.repo.UpdateSubmissionStatus(applicationID, status, &userID, "company")
}

func (s *CompanyPklService) CreateInternshipPosition(c *fiber.Ctx, req CreateInternshipRequest) (*pkl.Internship, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	user, err := s.repo.GetCompanyUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get company data")
	}

	internship := &pkl.Internship{
		CompanyProfileID: user.CompanyProfile.ID,
		Title:            req.Title,
		Description:      req.Description,
		Type:             req.Type,
		Deadline:         req.Deadline,
		PostedAt:         time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.repo.CreateInternshipPosition(internship); err != nil {
		return nil, err
	}

	return internship, nil
}

func (s *CompanyPklService) UpdateInternshipPosition(c *fiber.Ctx, internshipID uuid.UUID, req UpdateInternshipRequest) (*pkl.Internship, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	user, err := s.repo.GetCompanyUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get company data")
	}

	internship, err := s.repo.GetInternshipByID(internshipID)
	if err != nil {
		return nil, errors.New("internship not found")
	}

	// Verify that the internship belongs to this company
	if internship.CompanyProfileID != user.CompanyProfile.ID {
		return nil, errors.New("unauthorized: internship does not belong to this company")
	}

	// Update fields
	internship.Title = req.Title
	internship.Description = req.Description
	internship.Type = req.Type
	internship.Deadline = req.Deadline
	internship.UpdatedAt = time.Now()

	if err := s.repo.UpdateInternshipPosition(internship); err != nil {
		return nil, err
	}

	return internship, nil
}

func (s *CompanyPklService) DeleteInternshipPosition(c *fiber.Ctx, internshipID uuid.UUID) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return errors.New("failed to get user data")
	}

	user, err := s.repo.GetCompanyUserByID(userID)
	if err != nil {
		return errors.New("failed to get company data")
	}

	internship, err := s.repo.GetInternshipByID(internshipID)
	if err != nil {
		return errors.New("internship not found")
	}

	// Verify that the internship belongs to this company
	if internship.CompanyProfileID != user.CompanyProfile.ID {
		return errors.New("unauthorized: internship does not belong to this company")
	}

	return s.repo.DeleteInternshipPosition(internshipID)
}
