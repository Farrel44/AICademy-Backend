package auth

import (
	"github.com/Farrel44/AICademy-Backend/internal/domain/user"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(u *user.User) error {
	return r.db.Create(u).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*user.User, error) {
	var u user.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) GetUserByID(id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.db.First(&u, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	return r.db.Model(&user.User{}).Where("id = ?", userID).Update("password_hash", hashedPassword).Error
}

func (r *AuthRepository) CreateAlumniProfile(profile *user.AlumniProfile) error {
	return r.db.Create(profile).Error
}

func (r *AuthRepository) CreateStudentProfile(profile *user.StudentProfile) error {
	return r.db.Create(profile).Error
}

func (r *AuthRepository) CreateTeacherProfile(profile *user.TeacherProfile) error {
	return r.db.Create(profile).Error
}

func (r *AuthRepository) CreateCompanyProfile(profile *user.CompanyProfile) error {
	return r.db.Create(profile).Error
}
func (r *AuthRepository) CreateStudentsBulk(users []user.User, profiles []user.StudentProfile) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&users).Error; err != nil {
			return err
		}
		if err := tx.Create(&profiles).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *AuthRepository) SaveResetToken(email, token string, expiresAt time.Time) error {
	return r.db.Model(&user.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"password_reset_token": token,
		"password_reset_at":    expiresAt,
	}).Error
}

func (r *AuthRepository) GetUserByResetToken(token string) (*user.User, error) {
	var u user.User
	err := r.db.Where("password_reset_token = ? AND password_reset_at > ?", token, time.Now()).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) ClearResetToken(userID uuid.UUID) error {
	return r.db.Model(&user.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_reset_token": "",
		"password_reset_at":    nil,
	}).Error
}

func (r *AuthRepository) CheckNISExists(nis string) (bool, error) {
	var count int64
	err := r.db.Model(&user.StudentProfile{}).Where("nis = ?", nis).Count(&count).Error
	return count > 0, err
}
