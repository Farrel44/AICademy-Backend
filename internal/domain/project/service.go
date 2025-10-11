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
	repo         *ProjectRepository
	redis        *utils.RedisClient
	cacheManager *utils.CacheManager
	s3Client     *s3.Client
	bucket       string
	baseURL      string
}

func NewProjectService(repo *ProjectRepository, redis *utils.RedisClient) *ProjectService {
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
		repo:         repo,
		redis:        redis,
		cacheManager: utils.NewCacheManager(redis),
		s3Client:     client,
		bucket:       bucketName,
		baseURL:      baseURL,
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

// Cache invalidation functions
func (s *ProjectService) invalidateProjectCache(projectID uuid.UUID) {
	itemKey := fmt.Sprintf("project:%s", projectID.String())
	s.redis.Delete(itemKey)

	s.redis.Delete("project:statistics")
	s.cacheManager.InvalidateByPattern("projects:*")
}

func (s *ProjectService) invalidateCertificationCache(certificationID uuid.UUID) {
	itemKey := fmt.Sprintf("certification:%s", certificationID.String())
	s.redis.Delete(itemKey)

	s.redis.Delete("certification:statistics")
	s.cacheManager.InvalidateByPattern("certifications:*")
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

	// Add contributors if provided
	for _, contributorReq := range req.Contributors {
		if err := s.addContributorToProject(project.ID, &contributorReq); err != nil {
			// Log error but continue (don't fail the whole project creation)
			fmt.Printf("Failed to add contributor %s: %v\n", contributorReq.StudentID, err)
			continue
		}
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

	// Invalidate cache after successful creation
	s.invalidateProjectCache(project.ID)

	// Fetch the complete project with relations
	createdProject, err := s.repo.GetProjectByID(project.ID)
	if err != nil {
		return nil, err
	}

	return s.projectToResponse(createdProject), nil
}

func (s *ProjectService) addContributorToProject(projectID uuid.UUID, req *CreateContributorRequest) error {
	// Get student profile by NIS or other identifier
	studentProfile, err := s.repo.GetStudentProfileByNIS(req.StudentID)
	if err != nil {
		// Try by email if NIS fails
		studentProfile, err = s.repo.GetStudentProfileByEmail(req.StudentID)
		if err != nil {
			return fmt.Errorf("student with ID %s not found", req.StudentID)
		}
	}

	contributor := &ProjectContributor{
		ProjectID:        projectID,
		StudentProfileID: studentProfile.ID,
		RoleID:           req.RoleID,
	}

	return s.repo.AddProjectContributor(contributor)
}

func (s *ProjectService) GetProjectByID(id uuid.UUID) (*ProjectResponse, error) {
	project, err := s.repo.GetProjectByID(id)
	if err != nil {
		return nil, err
	}
	return s.projectToResponse(project), nil
}

func (s *ProjectService) GetMyProjects(c *fiber.Ctx, page, limit int, search string) (*utils.PaginationResponse, error) {
	claims, err := utils.GetClaimsFromHeader(c)
	if err != nil {
		return nil, errors.New("unauthorized")
	}

	// Validate search parameters
	validation, err := utils.ValidateSearchParams(search, page, limit)
	if err != nil {
		return nil, err
	}

	page = validation.Page
	limit = validation.Limit
	search = validation.Query

	if err != nil {
		return nil, errors.New("rate limit check failed")
	}

	// Limit pagination for caching
	if page > 10 || limit > 100 {
		return s.getProjectsFromDB(claims.UserID, page, limit, search)
	}

	// Generate cache keys
	cacheKey := s.cacheManager.GenerateCacheKey("user_projects", claims.UserID, page, limit, search)
	countKey := s.cacheManager.GenerateCacheKey("user_projects_count", claims.UserID, search)

	// Try to get from cache first
	var cachedResult utils.PaginationResponse
	if err := s.redis.GetJSON(cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	}

	// Get cached count first to avoid expensive COUNT query
	var total int64
	if err := s.redis.GetJSON(countKey, &total); err != nil {
		total, err = s.repo.CountProjectsByOwnerID(claims.UserID, search)
		if err != nil {
			return nil, errors.New("failed to count projects")
		}
		s.cacheManager.SetWithSmartTTL(countKey, total, "short")
	}

	offset := (page - 1) * limit

	// Get projects data
	projects, err := s.repo.GetProjectsByOwnerIDOptimized(claims.UserID, offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get projects")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := &utils.PaginationResponse{
		Data:       projects,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	// Cache the results with smart TTL
	s.cacheManager.SetWithSmartTTL(cacheKey, result, "medium")

	return result, nil
}

func (s *ProjectService) getProjectsFromDB(userID uuid.UUID, page, limit int, search string) (*utils.PaginationResponse, error) {
	total, err := s.repo.CountProjectsByOwnerID(userID, search)
	if err != nil {
		return nil, errors.New("failed to count projects")
	}

	offset := (page - 1) * limit
	projects, err := s.repo.GetProjectsByOwnerIDOptimized(userID, offset, limit, search)
	if err != nil {
		return nil, errors.New("failed to get projects")
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &utils.PaginationResponse{
		Data:       projects,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
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

	// Invalidate cache after update
	s.invalidateProjectCache(project.ID)

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

	// Get student profile by NIS or other identifier
	studentProfile, err := s.repo.GetStudentProfileByNIS(req.StudentID)
	if err != nil {
		// Try by email if NIS fails
		studentProfile, err = s.repo.GetStudentProfileByEmail(req.StudentID)
		if err != nil {
			return fmt.Errorf("student with ID %s not found", req.StudentID)
		}
	}

	contributor := &ProjectContributor{
		ProjectID:        projectID,
		StudentProfileID: studentProfile.ID,
		RoleID:           req.RoleID,
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

	// Invalidate cache after successful creation
	s.invalidateCertificationCache(certification.ID)

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

	// Invalidate cache after update
	s.invalidateCertificationCache(certification.ID)

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
	var contributors []ProjectContributorResponse
	for _, contributor := range project.Contributors {
		contributorResp := ProjectContributorResponse{
			StudentProfileID: contributor.StudentProfileID,
			RoleID:           contributor.RoleID,
		}

		if contributor.StudentProfile != nil {
			contributorResp.StudentName = contributor.StudentProfile.Fullname
			contributorResp.StudentNIS = contributor.StudentProfile.NIS
			contributorResp.StudentClass = contributor.StudentProfile.Class
		}

		if contributor.TargetRole != nil {
			contributorResp.Role = &TargetRoleResponse{
				ID:          contributor.TargetRole.ID,
				Name:        contributor.TargetRole.Name,
				Description: contributor.TargetRole.Description,
				Category:    contributor.TargetRole.Category,
			}
		}

		contributors = append(contributors, contributorResp)
	}

	return &ProjectResponse{
		ID:                    project.ID,
		OwnerStudentProfileID: project.OwnerStudentProfileID,
		ProjectName:           project.ProjectName,
		Description:           project.Description,
		LinkURL:               project.LinkURL,
		StartDate:             project.StartDate,
		EndDate:               project.EndDate,
		CreatedAt:             project.CreatedAt,
		Contributors:          contributors,
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
