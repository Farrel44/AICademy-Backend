package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTUser interface {
	GetID() uuid.UUID
	GetEmail() string
	GetRole() string
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

var (
	jwtSecretBytes []byte
	jwtOnce        sync.Once
)

func initJWTSecret() {
	jwtOnce.Do(func() {
		jwtSecretBytes = []byte(os.Getenv("JWT_SECRET"))
	})
}

func GenerateAccessToken(user JWTUser) (string, error) {
	initJWTSecret()

	if len(jwtSecretBytes) == 0 {
		return "", errors.New("JWT_SECRET not set")
	}

	expireTime := time.Now().Add(15 * time.Minute)

	claims := Claims{
		UserID: user.GetID(),
		Email:  user.GetEmail(),
		Role:   user.GetRole(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aicademy",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretBytes)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateTokenPair membuat access token dan refresh token sekaligus
func GenerateTokenPair(user JWTUser) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 menit dalam detik
	}, nil
}

// Legacy function untuk backward compatibility
func GenerateToken(user JWTUser) (string, error) {
	return GenerateAccessToken(user)
}

func ValidateToken(tokenString string) (*Claims, error) {
	initJWTSecret()
	if len(jwtSecretBytes) == 0 {
		return nil, errors.New("JWT_SECRET not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretBytes, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func GetClaimsFromHeader(c *fiber.Ctx) (*Claims, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is required")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || parts[1] == "" {
		return nil, errors.New("authorization header format must be Bearer {token}")
	}

	return ValidateToken(parts[1])
}
