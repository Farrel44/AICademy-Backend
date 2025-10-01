package auth

import (
	"errors"
	"strings"
	"time"

	authRepo "github.com/Farrel44/AICademy-Backend/internal/domain/auth"
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/google/uuid"
)

type CommonAuthService struct {
	repo *authRepo.AuthRepository
}

func NewCommonAuthService(repo *authRepo.AuthRepository) *CommonAuthService {
	return &CommonAuthService{repo: repo}
}

func (s *CommonAuthService) Login(req LoginRequest) (*AuthResponse, error) {
	foundUser, err := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(req.Password, foundUser.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	// Generate access token dan refresh token
	tokenPair, err := utils.GenerateTokenPair(foundUser)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Simpan refresh token ke database
	refreshTokenRecord := &user.RefreshToken{
		UserID:    foundUser.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 hari
	}

	err = s.repo.CreateRefreshToken(refreshTokenRecord)
	if err != nil {
		return nil, errors.New("failed to save refresh token")
	}

	var userName string
	switch foundUser.Role {
	case user.RoleStudent:
		profile, err := s.repo.GetStudentProfileByUserID(foundUser.ID)
		if err == nil {
			userName = profile.Fullname
		}
	case user.RoleAlumni:
		profile, err := s.repo.GetAlumniProfileByUserID(foundUser.ID)
		if err == nil {
			userName = profile.Fullname
		}
	case user.RoleTeacher:
		profile, err := s.repo.GetTeacherProfileByUserID(foundUser.ID)
		if err == nil {
			userName = profile.Fullname
		}
	}

	response := &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User: UserProfile{
			ID:    foundUser.ID,
			Email: foundUser.Email,
			Role:  string(foundUser.Role),
			Name:  userName,
		},
	}

	if foundUser.Role == user.RoleStudent && utils.CheckPassword("telkom@2025", foundUser.PasswordHash) {
		response.RequirePasswordChange = true
	}

	return response, nil
}

func (s *CommonAuthService) ChangePassword(userID uuid.UUID, req ChangePasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("password confirmation does not match")
	}

	foundUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !utils.CheckPassword(req.CurrentPassword, foundUser.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.repo.UpdatePassword(userID, hashedPassword)
}

func (s *CommonAuthService) GetMe(userID uuid.UUID) (interface{}, error) {
	foundUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	switch foundUser.Role {
	case user.RoleStudent:
		profile, err := s.repo.GetStudentProfileByUserID(userID)
		if err != nil {
			return nil, errors.New("student profile not found")
		}

		return StudentProfileResponse{
			ID:             profile.ID,
			UserID:         profile.UserID,
			Fullname:       profile.Fullname,
			NIS:            profile.NIS,
			Class:          profile.Class,
			ProfilePicture: profile.ProfilePicture,
			Headline:       profile.Headline,
			Bio:            profile.Bio,
			CVFile:         profile.CVFile,
			Email:          foundUser.Email,
			Role:           string(foundUser.Role),
			CreatedAt:      profile.CreatedAt,
			UpdatedAt:      profile.UpdatedAt,
		}, nil

	case user.RoleAlumni:
		profile, err := s.repo.GetAlumniProfileByUserID(userID)
		if err != nil {
			return nil, errors.New("alumni profile not found")
		}

		return AlumniProfileResponse{
			ID:             profile.ID,
			UserID:         profile.UserID,
			Fullname:       profile.Fullname,
			ProfilePicture: profile.ProfilePicture,
			Headline:       profile.Headline,
			Bio:            profile.Bio,
			CVFile:         profile.CVFile,
			Email:          foundUser.Email,
			Role:           string(foundUser.Role),
			CreatedAt:      profile.CreatedAt,
			UpdatedAt:      profile.UpdatedAt,
		}, nil

	case user.RoleTeacher:
		profile, err := s.repo.GetTeacherProfileByUserID(userID)
		if err != nil {
			return nil, errors.New("teacher profile not found")
		}

		return TeacherProfileResponse{
			ID:             profile.ID,
			UserID:         profile.UserID,
			Fullname:       profile.Fullname,
			ProfilePicture: profile.ProfilePicture,
			Email:          foundUser.Email,
			Role:           string(foundUser.Role),
			CreatedAt:      profile.CreatedAt,
		}, nil

	case user.RoleCompany:
		profile, err := s.repo.GetCompanyProfileByUserID(userID)
		if err != nil {
			return nil, errors.New("company profile not found")
		}

		return CompanyProfileResponse{
			ID:              profile.ID,
			UserID:          profile.UserID,
			CompanyName:     profile.CompanyName,
			CompanyLogo:     profile.CompanyLogo,
			CompanyLocation: profile.CompanyLocation,
			Description:     profile.Description,
			Email:           foundUser.Email,
			Role:            string(foundUser.Role),
			CreatedAt:       profile.CreatedAt,
		}, nil

	case user.RoleAdmin:
		return AdminProfileResponse{
			ID:        foundUser.ID,
			Email:     foundUser.Email,
			Role:      string(foundUser.Role),
			CreatedAt: foundUser.CreatedAt,
		}, nil

	default:
		return nil, errors.New("invalid user role")
	}
}

func (s *CommonAuthService) ForgotPassword(req ForgotPasswordRequest) error {
	foundUser, err := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	resetToken, err := utils.GenerateResetToken()
	if err != nil {
		return errors.New("failed to generate reset token")
	}

	expiry := time.Now().Add(1 * time.Hour) // 1 hour expiry

	err = s.repo.SaveResetToken(foundUser.Email, resetToken, expiry)
	if err != nil {
		return errors.New("failed to save reset token")
	}

	// Send email
	err = utils.SendResetPasswordEmail(foundUser, resetToken)
	if err != nil {
		return errors.New("failed to send reset email")
	}

	return nil
}

func (s *CommonAuthService) ResetPassword(token string, req ResetPasswordRequest) error {
	if req.Password != req.PasswordConfirm {
		return errors.New("password confirmation does not match")
	}

	foundUser, err := s.repo.GetUserByResetToken(token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	err = s.repo.UpdatePassword(foundUser.ID, hashedPassword)
	if err != nil {
		return errors.New("failed to update password")
	}

	// Clear reset token
	return s.repo.ClearResetToken(foundUser.ID)
}

func (s *CommonAuthService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token is required")
	}

	// Hapus refresh token dari database
	return s.repo.DeleteRefreshToken(refreshToken)
}

func (s *CommonAuthService) RefreshToken(req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate refresh token
	refreshTokenRecord, err := s.repo.GetRefreshTokenByToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Get user data
	foundUser, err := s.repo.GetUserByID(refreshTokenRecord.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new access token
	accessToken, err := utils.GenerateAccessToken(foundUser)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   15 * 60, // 15 menit dalam detik
	}, nil
}
