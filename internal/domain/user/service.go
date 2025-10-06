package user

import (
	"errors"
	"os"
	"strings"

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

func (s *UserService) validateToken(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET not found in environment")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *UserService) GetUserFromToken(c *fiber.Ctx) (*User, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, errors.New("authorization header format must be Bearer {token}")
	}

	tokenString := parts[1]

	claims, err := s.validateToken(tokenString)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}
