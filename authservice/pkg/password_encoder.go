package pkg

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func PasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("не удалось зашифровать пароль")
	}
	return string(hash), nil
}

func CheckEqualsPassword(password string, hash string) bool {
	if !checkPasswordHash(password, hash) {
		return false
	}

	h, err := PasswordHash(password)
	if err != nil {
		return false
	}

	return h == hash
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}

	return true
}
