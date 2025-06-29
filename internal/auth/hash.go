package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("empty string not allowed")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return "", fmt.Errorf("bcrypt error: %v", err)
	}

	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	if len(password) == 0 {
		return errors.New("empty string not allowed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("incorrect password: %v", err)
	}

	return nil
}
