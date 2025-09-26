package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"

	"github.com/Farrel44/AICademy-Backend/internal/domain/user"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db           *gorm.DB
	rdb          *redis.Client
	cacheVersion string
	cacheTTL     time.Duration
}

func NewRepository(db *gorm.DB, rdb *redis.Client) *AuthRepository {
	return &AuthRepository{
		db:           db,
		rdb:          rdb,
		cacheVersion: "v1",
		cacheTTL:     5 * time.Minute,
	}
}

func (r *AuthRepository) studentsCacheKey(page, pageSize int, q string) string {
	return fmt.Sprintf("students:%s:p:%d:s:%d:q:%s",
		r.cacheVersion, page, pageSize, url.QueryEscape(strings.ToLower(strings.TrimSpace(q))))
}

func (r *AuthRepository) cacheGet(ctx context.Context, key string, dst any) (bool, error) {
	b, err := r.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(b, dst)
}

func (r *AuthRepository) cacheSet(ctx context.Context, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, key, b, r.cacheTTL).Err()
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

func (r *AuthRepository) getStudentData(ctx context.Context, page, pageSize int, q string) (StudentData, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

	cacheKey := r.studentsCacheKey(page, pageSize, q)
	var cached StudentData
	if ok, _ := r.cacheGet(ctx, cacheKey, &cached); ok {
		return cached, nil
	}

	dbq := r.db.WithContext(ctx).Model(&user.StudentProfile{})

	if q = strings.TrimSpace(q); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		dbq = dbq.Where(`
			LOWER(fullname) LIKE ? OR LOWER(email) LIKE ? OR nis LIKE ? OR LOWER(class) LIKE ?`,
			like, like, like, like)
	}

	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return StudentData{}, nil
	}

	var profiles []user.StudentProfile

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

// Additional methods for getting profiles by user ID
func (r *AuthRepository) GetStudentProfileByUserID(userID uuid.UUID) (*user.StudentProfile, error) {
	var profile user.StudentProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *AuthRepository) GetAlumniProfileByUserID(userID uuid.UUID) (*user.AlumniProfile, error) {
	var profile user.AlumniProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

func (r *AuthRepository) GetTeacherProfileByUserID(userID uuid.UUID) (*user.TeacherProfile, error) {
	var profile user.TeacherProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	return &profile, err
}

// Refresh Token Management
func (r *AuthRepository) CreateRefreshToken(refreshToken *user.RefreshToken) error {
	return r.db.Create(refreshToken).Error
}

func (r *AuthRepository) GetRefreshTokenByToken(token string) (*user.RefreshToken, error) {
	var refreshToken user.RefreshToken
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *AuthRepository) DeleteRefreshToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&user.RefreshToken{}).Error
}

func (r *AuthRepository) DeleteAllRefreshTokensByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&user.RefreshToken{}).Error
}

func (r *AuthRepository) CleanupExpiredRefreshTokens() error {
	return r.db.Where("expires_at <= ?", time.Now()).Delete(&user.RefreshToken{}).Error
}
