package project

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProjectService struct {
	repo     *ProjectRepository
	s3Client *s3.Client
	bucket   string
	baseURL  string
}

func NewProjectService(repo *ProjectRepository) *ProjectService {
	bucketName := os.Getenv("R2_BUCKET_NAME")
	accountId := os.Getenv("R2_ACCOUNT_ID")
	accessKeyId := os.Getenv("R2_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
	baseURL := os.Getenv("OBJECT_STORAGE_URL")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS config: %v", err))
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	return &ProjectService{
		repo:     repo,
		s3Client: client,
		bucket:   bucketName,
		baseURL:  baseURL,
	}
}

func (s *ProjectService) uploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Upload to R2
	_, err = s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
		Body:   src,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.baseURL, filename), nil
}

// Project methods
func (s *ProjectService) CreateProject(c *fiber.Ctx, req *CreateProjectRequest) (*ProjectResponse, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user id from token")
	}

	studentProfileID, err := s.repo.GetStudentProfileIDByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to get student profile")
	}

	project := &Project{
		OwnerStudentProfileID: studentProfileID,
		ProjectName:           req.ProjectName,
		Description:           req.Description,
		LinkURL:               req.LinkURL,
		StartDate:             req.StartDate,
		EndDate:               req.EndDate,
		CreatedAt:             time.Now(),
	}

	if err := s.repo.CreateProject(project); err != nil {
		return nil, err
	}

	// Upload photos if provided
	for _, photo := range req.Photos {
		if photo == nil {
			continue
		}

		url, err := s.uploadFile(photo, "projects")
		if err != nil {
			continue // Skip failed uploads
		}

		projectPhoto := &ProjectPhoto{
			ProjectID: project.ID,
			URL:       url,
			CreatedAt: time.Now(),
		}
		s.repo.AddProjectPhoto(projectPhoto)
	}

	// Fetch the complete project with relations
	createdProject, err := s.repo.GetProjectByID(project.ID)
	if err != nil {
		return nil, err
	}

	return s.projectToResponse(createdProject), nil
}

func (s *ProjectService) GetProjectByID(id uuid.UUID) (*ProjectResponse, error) {
	project, err := s.repo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}
	return s.projectToResponse(project), nil
}

func (s *ProjectService) GetMyProjects(c *fiber.Ctx) ([]ProjectResponse, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user id from token")
	}

	studentProfileID, err := s.repo.GetStudentProfileIDByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to get student profile")
	}

	projects, err := s.repo.GetProjectsByOwnerID(studentProfileID)
	if err != nil {
		return nil, err
	}

	var responses []ProjectResponse
	for _, project := range projects {
		responses = append(responses, *s.projectToResponse(&project))
	}

	return responses, nil
}

func (s *ProjectService) UpdateProject(id uuid.UUID, req *UpdateProjectRequest) (*ProjectResponse, error) {
	project, err := s.repo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	if req.ProjectName != "" {
		project.ProjectName = req.ProjectName
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.LinkURL != nil {
		project.LinkURL = req.LinkURL
	}
	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		project.EndDate = *req.EndDate
	}

	if err := s.repo.UpdateProject(project); err != nil {
		return nil, err
	}

	// Upload new photos if provided
	for _, photo := range req.Photos {
		if photo == nil {
			continue
		}

		url, err := s.uploadFile(photo, "projects")
		if err != nil {
			continue
		}

		projectPhoto := &ProjectPhoto{
			ProjectID: project.ID,
			URL:       url,
			CreatedAt: time.Now(),
		}
		s.repo.AddProjectPhoto(projectPhoto)
	}

	updatedProject, err := s.repo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	return s.projectToResponse(updatedProject), nil
}

func (s *ProjectService) DeleteProject(id uuid.UUID) error {
	return s.repo.DeleteProject(id)
}

func (s *ProjectService) AddProjectContributor(c *fiber.Ctx, projectID uuid.UUID, req *AddProjectContributorRequest) error {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return errors.New("failed to get user id from token")
	}

	studentProfileID, err := s.repo.GetStudentProfileIDByUserID(userID)
	if err != nil {
		return errors.New("failed to get student profile")
	}

	// Check if user owns the project
	project, err := s.repo.GetProjectByID(projectID)
	if err != nil {
		return errors.New("project not found")
	}

	if project.OwnerStudentProfileID != studentProfileID {
		return errors.New("unauthorized to add contributors to this project")
	}

	contributor := &ProjectContributor{
		ProjectID:        projectID,
		StudentProfileID: req.StudentProfileID,
		ProjectRole:      req.ProjectRole,
		ProfilingRoleID:  req.ProfilingRoleID,
		Description:      req.Description,
	}

	return s.repo.AddProjectContributor(contributor)
}

// Certification methods
func (s *ProjectService) CreateCertification(c *fiber.Ctx, req *CreateCertificationRequest) (*CertificationResponse, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user id from token")
	}

	studentProfileID, err := s.repo.GetStudentProfileIDByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to get student profile")
	}

	certification := &Certification{
		StudentProfileID:    studentProfileID,
		Name:                req.Name,
		IssuingOrganization: req.IssuingOrganization,
		IssueDate:           req.IssueDate,
		ExpirationDate:      req.ExpirationDate,
		CredentialID:        req.CredentialID,
		CredentialURL:       req.CredentialURL,
		CreatedAt:           time.Now(),
	}

	if err := s.repo.CreateCertification(certification); err != nil {
		return nil, err
	}

	// Upload photos if provided
	for _, photo := range req.Photos {
		if photo == nil {
			continue
		}

		url, err := s.uploadFile(photo, "certifications")
		if err != nil {
			continue
		}

		certPhoto := &CertificationPhoto{
			CertificationID: certification.ID,
			URL:             url,
			CreatedAt:       time.Now(),
		}
		s.repo.AddCertificationPhoto(certPhoto)
	}

	createdCertification, err := s.repo.GetCertificationByID(certification.ID)
	if err != nil {
		return nil, err
	}

	return s.certificationToResponse(createdCertification), nil
}

func (s *ProjectService) GetCertificationByID(id uuid.UUID) (*CertificationResponse, error) {
	certification, err := s.repo.GetCertificationByID(id)
	if err != nil {
		return nil, err
	}
	return s.certificationToResponse(certification), nil
}

func (s *ProjectService) GetMyCertifications(c *fiber.Ctx) ([]CertificationResponse, error) {
	userID, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user id from token")
	}

	studentProfileID, err := s.repo.GetStudentProfileIDByUserID(userID)
	if err != nil {
		return nil, errors.New("failed to get student profile")
	}

	certifications, err := s.repo.GetCertificationsByStudentID(studentProfileID)
	if err != nil {
		return nil, err
	}

	var responses []CertificationResponse
	for _, cert := range certifications {
		responses = append(responses, *s.certificationToResponse(&cert))
	}

	return responses, nil
}

func (s *ProjectService) UpdateCertification(id uuid.UUID, req *UpdateCertificationRequest) (*CertificationResponse, error) {
	certification, err := s.repo.GetCertificationByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		certification.Name = *req.Name
	}
	if req.IssuingOrganization != nil {
		certification.IssuingOrganization = *req.IssuingOrganization
	}
	if req.IssueDate != nil {
		certification.IssueDate = *req.IssueDate
	}
	if req.ExpirationDate != nil {
		certification.ExpirationDate = req.ExpirationDate
	}
	if req.CredentialID != nil {
		certification.CredentialID = req.CredentialID
	}
	if req.CredentialURL != nil {
		certification.CredentialURL = req.CredentialURL
	}

	if err := s.repo.UpdateCertification(certification); err != nil {
		return nil, err
	}

	// Upload new photos if provided
	for _, photo := range req.Photos {
		if photo == nil {
			continue
		}

		url, err := s.uploadFile(photo, "certifications")
		if err != nil {
			continue
		}

		certPhoto := &CertificationPhoto{
			CertificationID: certification.ID,
			URL:             url,
			CreatedAt:       time.Now(),
		}
		s.repo.AddCertificationPhoto(certPhoto)
	}

	updatedCertification, err := s.repo.GetCertificationByID(id)
	if err != nil {
		return nil, err
	}

	return s.certificationToResponse(updatedCertification), nil
}

func (s *ProjectService) DeleteCertification(id uuid.UUID) error {
	return s.repo.DeleteCertification(id)
}

// Helper methods
func (s *ProjectService) projectToResponse(project *Project) *ProjectResponse {
	return &ProjectResponse{
		ID:                    project.ID,
		OwnerStudentProfileID: project.OwnerStudentProfileID,
		ProjectName:           project.ProjectName,
		Description:           project.Description,
		LinkURL:               project.LinkURL,
		StartDate:             project.StartDate,
		EndDate:               project.EndDate,
		CreatedAt:             project.CreatedAt,
		Contributors:          project.Contributors,
		Photos:                project.Photos,
	}
}

func (s *ProjectService) certificationToResponse(cert *Certification) *CertificationResponse {
	return &CertificationResponse{
		ID:                  cert.ID,
		StudentProfileID:    cert.StudentProfileID,
		Name:                cert.Name,
		IssuingOrganization: cert.IssuingOrganization,
		IssueDate:           cert.IssueDate,
		ExpirationDate:      cert.ExpirationDate,
		CredentialID:        cert.CredentialID,
		CredentialURL:       cert.CredentialURL,
		CreatedAt:           cert.CreatedAt,
		IsExpired:           cert.IsExpired(),
		IsExpiringSoon:      cert.IsExpiringSoon(),
		Photos:              cert.Photos,
	}
}
