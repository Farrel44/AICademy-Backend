package student

import (
	"errors"
	"strings"
	"time"

	"aicademy-backend/internal/domain/auth"
	commonAuth "aicademy-backend/internal/domain/common/auth"
	"aicademy-backend/internal/domain/user"
	"aicademy-backend/internal/utils"

	"github.com/google/uuid"
)

type StudentAuthService struct {
	repo *auth.AuthRepository
}

func NewStudentAuthService(repo *auth.AuthRepository) *StudentAuthService {
	return &StudentAuthService{repo: repo}
}

const DefaultStudentPassword = "telkom@2025"

func (s *StudentAuthService) CreateStudent(req CreateStudentRequest) error {
	existingUser, _ := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	nisExists, err := s.repo.CheckNISExists(req.NIS)
	if err != nil {
		return errors.New("failed to check NIS")
	}
	if nisExists {
		return errors.New("student with this NIS already exists")
	}

	hashedPassword, err := utils.HashPassword(DefaultStudentPassword)
	if err != nil {
		return errors.New("failed to generate password")
	}

	newUser := &user.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
		Role:         user.RoleStudent,
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return errors.New("failed to create user account")
	}

	studentProfile := &user.StudentProfile{
		UserID:   newUser.ID,
		NIS:      req.NIS,
		Fullname: req.Fullname,
		Class:    req.Class,
	}

	err = s.repo.CreateStudentProfile(studentProfile)
	if err != nil {
		return errors.New("failed to create student profile")
	}

	return nil
}

func (s *StudentAuthService) ChangeDefaultPassword(userID uuid.UUID, req ChangeDefaultPasswordRequest) (*commonAuth.AuthResponse, error) {

	foundUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if foundUser.Role != user.RoleStudent {
		return nil, errors.New("user is not a student")
	}

	if !utils.CheckPassword(DefaultStudentPassword, foundUser.PasswordHash) {
		return nil, errors.New("password has already been changed from default")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, errors.New("failed to hash new password")
	}

	err = s.repo.UpdatePassword(foundUser.ID, hashedPassword)
	if err != nil {
		return nil, errors.New("failed to update password")
	}

	tokenPair, err := utils.GenerateTokenPair(foundUser)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	refreshTokenRecord := &user.RefreshToken{
		UserID:    foundUser.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = s.repo.CreateRefreshToken(refreshTokenRecord)
	if err != nil {
		return nil, errors.New("failed to save refresh token")
	}

	return &commonAuth.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User: commonAuth.UserProfile{
			ID:    foundUser.ID,
			Email: foundUser.Email,
			Role:  string(foundUser.Role),
		},
		RequirePasswordChange: false,
	}, nil
}
