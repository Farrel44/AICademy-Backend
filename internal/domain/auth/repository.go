package auth

import (
	"aicademy-backend/internal/domain/user"

	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user *user.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*user.User, error) {
	var user user.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetUserByID(id uint) (*user.User, error) {
	var user user.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
