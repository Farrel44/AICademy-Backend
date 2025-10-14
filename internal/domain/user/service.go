package user

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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserService struct {
	repo     *UserRepository
	s3Client *s3.Client
	bucket   string
	baseURL  string
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   UserRole  `json:"role"`
	jwt.RegisteredClaims
}

func NewUserService(repo *UserRepository) *UserService {
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

	return &UserService{
		repo:     repo,
		s3Client: client,
		bucket:   bucketName,
		baseURL:  baseURL,
	}
}

func (s *UserService) uploadFile(file *multipart.FileHeader, folder string) (string, error) {
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

func (s *UserService) GetUserByToken(c *fiber.Ctx) (*User, error) {
	userId, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user id")
	}
	user, err := s.repo.GetUserByID(userId)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}
	return user, nil
}

func (s *UserService) GetEnhancedUserProfile(c *fiber.Ctx) (*EnhancedUserResponse, error) {
	userId, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("failed to get user id")
	}

	user, err := s.repo.GetUserByID(userId)
	if err != nil {
		return nil, errors.New("failed to get user data")
	}

	response := &EnhancedUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	if user.StudentProfile != nil {
		enhancedProfile, err := s.buildEnhancedStudentProfile(user.StudentProfile, userId)
		if err != nil {
			return nil, err
		}
		response.StudentProfile = enhancedProfile
		response.Username = s.generateUsername(user.StudentProfile.Fullname, user.StudentProfile.NIS)
	}

	if user.AlumniProfile != nil {
		enhancedProfile, err := s.buildEnhancedAlumniProfile(user.AlumniProfile)
		if err != nil {
			return nil, err
		}
		response.AlumniProfile = enhancedProfile
		response.Username = s.generateUsername(user.AlumniProfile.Fullname, "alumni")
	}

	return response, nil
}

func (s *UserService) GetPublicStudentProfileByNIS(nis string) (*PublicStudentProfileResponse, error) {
	// Get student profile by NIS
	studentProfile, err := s.repo.GetStudentProfileByNIS(nis)
	if err != nil {
		return nil, errors.New("student not found")
	}

	// Get projects
	projects, err := s.repo.GetUserProjectsByStudentProfileID(studentProfile.ID)
	if err != nil {
		projects = []UserProjectInfo{} // Empty slice if error
	}

	// Get certifications
	certifications, err := s.repo.GetUserCertificationsByStudentProfileID(studentProfile.ID)
	if err != nil {
		certifications = []UserCertificationInfo{} // Empty slice if error
	}

	// Get recommended role
	recommendedRole, _ := s.repo.GetStudentRecommendedRole(studentProfile.UserID)

	// Generate profile URL
	profileURL := os.Getenv("FRONTEND_URL") + "/profile/" + nis

	return &PublicStudentProfileResponse{
		NIS:             studentProfile.NIS,
		Fullname:        studentProfile.Fullname,
		Class:           studentProfile.Class,
		ProfilePicture:  studentProfile.ProfilePicture,
		Headline:        studentProfile.Headline,
		Bio:             studentProfile.Bio,
		Projects:        projects,
		Certifications:  certifications,
		ProfileURL:      profileURL,
		JoinedAt:        studentProfile.CreatedAt,
		LastActive:      studentProfile.UpdatedAt,
		RecommendedRole: recommendedRole,
		ShowEmail:       false,
		ShowCV:          false,
		IsPublicProfile: true,
	}, nil
}

func (s *UserService) buildEnhancedStudentProfile(profile *StudentProfile, userID uuid.UUID) (*EnhancedStudentProfile, error) {
	enhancedProfile := &EnhancedStudentProfile{
		ID:             profile.ID,
		UserID:         profile.UserID,
		Fullname:       profile.Fullname,
		NIS:            profile.NIS,
		Class:          profile.Class,
		ProfilePicture: profile.ProfilePicture,
		Headline:       profile.Headline,
		Bio:            profile.Bio,
		CVFile:         profile.CVFile,
		CreatedAt:      profile.CreatedAt,
		UpdatedAt:      profile.UpdatedAt,
	}

	// Get recommended role
	recommendedRole, err := s.repo.GetStudentRecommendedRole(userID)
	if err == nil && recommendedRole != nil {
		enhancedProfile.RecommendedRole = recommendedRole
	}

	// Get projects
	projects, err := s.repo.GetUserProjectsByStudentProfileID(profile.ID)
	if err == nil {
		enhancedProfile.Projects = projects
	} else {
		enhancedProfile.Projects = []UserProjectInfo{} // Empty slice instead of nil
	}

	// Get certifications
	certifications, err := s.repo.GetUserCertificationsByStudentProfileID(profile.ID)
	if err == nil {
		enhancedProfile.Certifications = certifications
	} else {
		enhancedProfile.Certifications = []UserCertificationInfo{} // Empty slice instead of nil
	}

	enhancedProfile.ProfileURL = os.Getenv("FRONTEND_URL") + "/profile/" + profile.NIS

	return enhancedProfile, nil
}

func (s *UserService) buildEnhancedAlumniProfile(profile *AlumniProfile) (*EnhancedAlumniProfile, error) {
	enhancedProfile := &EnhancedAlumniProfile{
		ID:             profile.ID,
		UserID:         profile.UserID,
		Fullname:       profile.Fullname,
		ProfilePicture: profile.ProfilePicture,
		Headline:       profile.Headline,
		Bio:            profile.Bio,
		CVFile:         profile.CVFile,
		CreatedAt:      profile.CreatedAt,
		UpdatedAt:      profile.UpdatedAt,
		Projects:       []UserProjectInfo{},
		Certifications: []UserCertificationInfo{},
	}

	// Generate profile URL
	username := s.generateUsername(profile.Fullname, "alumni")
	enhancedProfile.ProfileURL = os.Getenv("FRONTEND_URL") + "/profile/" + username

	return enhancedProfile, nil
}

func (s *UserService) generateUsername(fullname, identifier string) string {
	// Simple username generation logic
	return fullname + "_" + identifier
}

func (s *UserService) GetStudentWithRecommendedRole(c *fiber.Ctx) (*EnhancedUserResponse, error) {
	return s.GetEnhancedUserProfile(c)
}

func (s *UserService) UpdateUserProfile(c *fiber.Ctx) (*StudentProfile, error) {
	userId, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("error getting user id")
	}

	user, err := s.repo.GetUserByID(userId)
	if err != nil || user == nil {
		return nil, errors.New("failed to fetch current user data")
	}

	if user.StudentProfile == nil {
		return nil, errors.New("student profile not found")
	}

	// Parse form data
	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("failed to parse form data")
	}

	// Handle text fields
	if bio := form.Value["bio"]; len(bio) > 0 && bio[0] != "" {
		user.StudentProfile.Bio = bio[0]
	}

	if headline := form.Value["headline"]; len(headline) > 0 && headline[0] != "" {
		user.StudentProfile.Headline = headline[0]
	}

	// Handle profile picture upload
	if profilePictures := form.File["profile_picture"]; len(profilePictures) > 0 {
		profilePicture := profilePictures[0]

		// Validate file type
		allowedTypes := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
			".webp": true,
		}

		ext := filepath.Ext(profilePicture.Filename)
		if !allowedTypes[ext] {
			return nil, errors.New("invalid file type. Only jpg, jpeg, png, gif, webp are allowed")
		}

		// Validate file size (max 5MB)
		if profilePicture.Size > 5*1024*1024 {
			return nil, errors.New("file size too large. Maximum 5MB allowed")
		}

		url, err := s.uploadFile(profilePicture, "profile-pictures")
		if err != nil {
			return nil, fmt.Errorf("failed to upload profile picture: %v", err)
		}

		user.StudentProfile.ProfilePicture = url
	}

	// Handle CV file upload
	if cvFiles := form.File["cv_file"]; len(cvFiles) > 0 {
		cvFile := cvFiles[0]

		// Validate file type for CV
		allowedCVTypes := map[string]bool{
			".pdf":  true,
			".doc":  true,
			".docx": true,
		}

		ext := filepath.Ext(cvFile.Filename)
		if !allowedCVTypes[ext] {
			return nil, errors.New("invalid CV file type. Only pdf, doc, docx are allowed")
		}

		// Validate file size (max 10MB for CV)
		if cvFile.Size > 10*1024*1024 {
			return nil, errors.New("CV file size too large. Maximum 10MB allowed")
		}

		url, err := s.uploadFile(cvFile, "cv-files")
		if err != nil {
			return nil, fmt.Errorf("failed to upload CV file: %v", err)
		}

		user.StudentProfile.CVFile = &url
	}

	user.StudentProfile.UpdatedAt = time.Now()

	err = s.repo.UpdateStudentProfile(user.StudentProfile)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return user.StudentProfile, nil
}
