package auth

import (
	"aicademy-backend/internal/domain/user"
	"aicademy-backend/internal/utils"
	"errors"
)

type AuthService struct {
	repo *AuthRepository
}

func NewAuthService(repo *AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	existingUser, _ := s.repo.GetUserByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	newUser := &user.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     user.UserRole(req.Role),
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(newUser)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserProfile{
			ID:    newUser.ID,
			Name:  newUser.Name,
			Email: newUser.Email,
			Role:  string(newUser.Role),
		},
	}, nil
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User: UserProfile{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		},
	}, nil
}
