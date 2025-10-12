package user

import (
	"errors"
	"os"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserService struct {
	repo *UserRepository
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   UserRole  `json:"role"`
	jwt.RegisteredClaims
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
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
	profileURL := os.Getenv("FRONTEND_URL") + nis

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

	enhancedProfile.ProfileURL = os.Getenv("FRONTEND_URL") + profile.NIS

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
	enhancedProfile.ProfileURL = "https://aiademy.smktelkom-pwt.sch.id/" + username

	return enhancedProfile, nil
}

func (s *UserService) generateUsername(fullname, identifier string) string {
	// Simple username generation logic
	return fullname + "_" + identifier
}

func (s *UserService) GetStudentWithRecommendedRole(c *fiber.Ctx) (*EnhancedUserResponse, error) {
	return s.GetEnhancedUserProfile(c)
}

func (s *UserService) UpdateUserProfile(c *fiber.Ctx, req *UpdateStudentRequest) (*StudentProfile, error) {
	userId, err := utils.GetUserIDFromToken(c)
	if err != nil {
		return nil, errors.New("error getting user id")
	}
	user, _ := s.repo.GetUserByID(userId)
	if user == nil {
		return nil, errors.New("Failed to fetch current user data")
	}
	if req.Bio != nil {
		user.StudentProfile.Bio = *req.Bio
	}

	if req.CvFile != nil {
		user.StudentProfile.CVFile = req.CvFile
	}

	if req.Headline != nil {
		user.StudentProfile.Headline = *req.Headline
	}

	if req.ProfilePicture != nil {
		user.StudentProfile.ProfilePicture = *req.ProfilePicture
	}

	user.StudentProfile.UpdatedAt = time.Now()

	err = s.repo.UpdateStudentProfile(user.StudentProfile)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return user.StudentProfile, nil
}
