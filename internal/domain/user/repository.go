package user

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewUserRepository(db *gorm.DB, rdb *redis.Client) *UserRepository {
	return &UserRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

func (r *UserRepository) GetUserByID(id uuid.UUID) (*User, error) {
	var u User
	err := r.db.
		Preload("StudentProfile").
		First(&u, "id = ?", id).Error
	if err != nil {
		fmt.Printf("err %s", err)
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdateStudentProfile(student *StudentProfile) error {
	return r.db.Save(student).Error
}
