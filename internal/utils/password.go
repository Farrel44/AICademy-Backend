package utils

import (
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	bcryptCost = bcrypt.DefaultCost
	costOnce   sync.Once
)

func initBcryptCost() {
	costOnce.Do(func() {
		env := os.Getenv("APP_ENV")
		if env == "development" || env == "testing" {
			bcryptCost = bcrypt.MinCost
		} else if costStr := os.Getenv("BCRYPT_COST"); costStr != "" {
			if cost, err := strconv.Atoi(costStr); err == nil && cost >= 4 && cost <= 15 {
				bcryptCost = cost
			}
		} else {
			bcryptCost = 10
		}
	})
}

func HashPassword(password string) (string, error) {
	initBcryptCost()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
