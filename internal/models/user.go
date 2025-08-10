package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	UUID           string    `json:"-"`
	Login          string    `json:"-"`
	HashedPassword string    `json:"-"`
	Balance        int64     `json:"-"`
	CreatedAt      time.Time `json:"-"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
