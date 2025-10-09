package user

import (
	"errors"
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

func (s *UserService) GetStudentWithRecommendedRole(c *fiber.Ctx) (*EnhancedUserResponse, error) {
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
		enhancedProfile := &EnhancedStudentProfile{
			ID:             user.StudentProfile.ID,
			UserID:         user.StudentProfile.UserID,
			Fullname:       user.StudentProfile.Fullname,
			NIS:            user.StudentProfile.NIS,
			Class:          user.StudentProfile.Class,
			ProfilePicture: user.StudentProfile.ProfilePicture,
			Headline:       user.StudentProfile.Headline,
			Bio:            user.StudentProfile.Bio,
			CVFile:         user.StudentProfile.CVFile,
			CreatedAt:      user.StudentProfile.CreatedAt,
			UpdatedAt:      user.StudentProfile.UpdatedAt,
		}

		recommendedRole, err := s.repo.GetStudentRecommendedRole(userId)
		if err != nil {
			return nil, errors.New("failed to get recommended role")
		}

		enhancedProfile.RecommendedRole = recommendedRole

		response.StudentProfile = enhancedProfile
	}

	return response, nil
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
