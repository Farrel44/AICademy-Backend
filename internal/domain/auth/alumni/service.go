package alumni

import (
	"errors"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	commonAuth "github.com/Farrel44/AICademy-Backend/internal/domain/common/auth"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"
)

type AlumniAuthService struct {
	repo *auth.AuthRepository
}

func NewAlumniAuthService(repo *auth.AuthRepository) *AlumniAuthService {
	return &AlumniAuthService{repo: repo}
}

func (s *AlumniAuthService) RegisterAlumni(req RegisterAlumniRequest) (*commonAuth.AuthResponse, error) {
	existingUser, _ := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	newUser := &user.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
		Role:         user.RoleAlumni,
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return nil, errors.New("failed to create user account")
	}
	alumniProfile := &user.AlumniProfile{
		UserID:   newUser.ID,
		Fullname: req.Fullname,
	}

	err = s.repo.CreateAlumniProfile(alumniProfile)
	if err != nil {
		return nil, errors.New("failed to create alumni profile")
	}

	// Generate access token dan refresh token
	tokenPair, err := utils.GenerateTokenPair(newUser)
	if err != nil {
		return nil, errors.New("failed to generate authentication token")
	}

	// Simpan refresh token ke database
	refreshTokenRecord := &user.RefreshToken{
		UserID:    newUser.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 hari
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
			ID:    newUser.ID,
			Email: newUser.Email,
			Role:  string(newUser.Role),
		},
	}, nil
}
