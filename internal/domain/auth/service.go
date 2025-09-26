package auth

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"github.com/Farrel44/AICademy-Backend/internal/utils"

	"github.com/google/uuid"
)

const DefaultStudentPassword = "telkom@2025"

type AuthService struct {
	repo *AuthRepository
}

func NewAuthService(repo *AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) RegisterAlumni(req RegisterAlumniRequest) (*AuthResponse, error) {
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
		return nil, errors.New("failed to create user")
	}

	alumniProfile := &user.AlumniProfile{
		UserID:   newUser.ID,
		Fullname: req.Fullname,
	}

	err = s.repo.CreateAlumniProfile(alumniProfile)
	if err != nil {
		return nil, errors.New("failed to create alumni profile")
	}

	token, err := utils.GenerateToken(newUser)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: token,
		User: UserProfile{
			ID:    newUser.ID,
			Email: newUser.Email,
			Role:  string(newUser.Role),
		},
	}, nil
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	foundUser, err := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(req.Password, foundUser.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(foundUser)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	response := &AuthResponse{
		Token: token,
		User: UserProfile{
			ID:    foundUser.ID,
			Email: foundUser.Email,
			Role:  string(foundUser.Role),
		},
	}

	if foundUser.Role == user.RoleStudent && utils.CheckPassword(DefaultStudentPassword, foundUser.PasswordHash) {
		response.RequirePasswordChange = true
	}

	return response, nil
}

func (s *AuthService) CreateTeacher(req CreateTeacherRequest) error {
	existingUser, _ := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	newUser := &user.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
		Role:         user.RoleTeacher,
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return errors.New("failed to create user")
	}

	teacherProfile := &user.TeacherProfile{
		UserID:   newUser.ID,
		Fullname: req.Fullname,
	}

	err = s.repo.CreateTeacherProfile(teacherProfile)
	if err != nil {
		return errors.New("failed to create teacher profile")
	}

	return nil
}

func (s *AuthService) CreateStudent(req CreateStudentRequest) error {
	exists, err := s.repo.CheckNISExists(req.NIS)
	if err != nil {
		return errors.New("failed to check NIS")
	}
	if exists {
		return errors.New("student with this NIS already exists")
	}

	existingUser, _ := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(DefaultStudentPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	newUser := &user.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
		Role:         user.RoleStudent,
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return errors.New("failed to create user")
	}

	studentProfile := &user.StudentProfile{
		UserID:   newUser.ID,
		Fullname: req.Fullname,
		NIS:      req.NIS,
		Class:    req.Class,
	}

	err = s.repo.CreateStudentProfile(studentProfile)
	if err != nil {
		return errors.New("failed to create student profile")
	}

	return nil
}

func (s *AuthService) CreateCompany(req CreateCompanyRequest) error {
	existingUser, _ := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	newUser := &user.User{
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
		Role:         user.RoleCompany,
	}

	err = s.repo.CreateUser(newUser)
	if err != nil {
		return errors.New("failed to create user")
	}

	companyProfile := &user.CompanyProfile{
		UserID:          newUser.ID,
		CompanyName:     req.CompanyName,
		CompanyLogo:     req.CompanyLogo,
		CompanyLocation: req.CompanyLocation,
		Description:     req.Description,
	}

	err = s.repo.CreateCompanyProfile(companyProfile)
	if err != nil {
		return errors.New("failed to create company profile")
	}

	return nil
}

func (s *AuthService) CreateStudentsFromCSV(csvReader *csv.Reader) (int, []string, error) {
	var users []user.User
	var profiles []user.StudentProfile
	var validationErrors []string

	records, err := csvReader.ReadAll()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	hashedPassword, err := utils.HashPassword(DefaultStudentPassword)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to hash default password: %v", err)
	}

	for i, record := range records {
		if i == 0 {
			continue
		}

		if len(record) < 4 {
			validationErrors = append(validationErrors, fmt.Sprintf("Row %d: insufficient columns", i+1))
			continue
		}

		csvRow := StudentCSVRow{
			NIS:      strings.TrimSpace(record[0]),
			Class:    strings.TrimSpace(record[1]),
			Email:    strings.ToLower(strings.TrimSpace(record[2])),
			Fullname: strings.TrimSpace(record[3]),
		}

		if err := utils.ValidateStruct(csvRow); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		exists, err := s.repo.CheckNISExists(csvRow.NIS)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Row %d: failed to check NIS", i+1))
			continue
		}
		if exists {
			validationErrors = append(validationErrors, fmt.Sprintf("Row %d: NIS %s already exists", i+1, csvRow.NIS))
			continue
		}

		existingUser, _ := s.repo.GetUserByEmail(csvRow.Email)
		if existingUser != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Row %d: email %s already exists", i+1, csvRow.Email))
			continue
		}

		userID := uuid.New()
		newUser := user.User{
			ID:           userID,
			Email:        csvRow.Email,
			PasswordHash: hashedPassword,
			Role:         user.RoleStudent,
		}

		studentProfile := user.StudentProfile{
			UserID:   userID,
			Fullname: csvRow.Fullname,
			NIS:      csvRow.NIS,
			Class:    csvRow.Class,
		}

		users = append(users, newUser)
		profiles = append(profiles, studentProfile)
	}

	if len(users) == 0 {
		return 0, validationErrors, fmt.Errorf("no valid records to create")
	}

	err = s.repo.CreateStudentsBulk(users, profiles)
	if err != nil {
		return 0, validationErrors, fmt.Errorf("failed to create students: %v", err)
	}

	return len(users), validationErrors, nil
}

func (s *AuthService) ChangePassword(userID uuid.UUID, req ChangePasswordRequest) error {
	foundUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !utils.CheckPassword(req.CurrentPassword, foundUser.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	if req.NewPassword != req.ConfirmPassword {
		return errors.New("new password and confirmation do not match")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	err = s.repo.UpdatePassword(userID, hashedPassword)
	if err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *AuthService) ForgotPassword(req ForgotPasswordRequest) error {
	existingUser, err := s.repo.GetUserByEmail(strings.ToLower(req.Email))
	if err != nil {
		return nil
	}

	resetToken, err := utils.GenerateRandomString(20)
	if err != nil {
		return errors.New("failed to generate reset token")
	}

	encodedToken := utils.Encode(resetToken)
	expiresAt := time.Now().Add(15 * time.Minute)

	err = s.repo.SaveResetToken(existingUser.Email, encodedToken, expiresAt)
	if err != nil {
		return errors.New("failed to save reset token")
	}

	err = utils.SendResetPasswordEmail(existingUser, resetToken)
	if err != nil {
		return errors.New("failed to send reset email")
	}

	return nil
}

func (s *AuthService) ResetPassword(token string, req ResetPasswordRequest) error {
	if req.Password != req.PasswordConfirm {
		return errors.New("password and confirmation do not match")
	}

	encodedToken := utils.Encode(token)

	foundUser, err := s.repo.GetUserByResetToken(encodedToken)
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

	err = s.repo.ClearResetToken(foundUser.ID)
	if err != nil {
		return errors.New("failed to clear reset token")
	}

	return nil
}
