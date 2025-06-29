package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword calls HashPassword
func TestHashPassword(t *testing.T) {
	input := "password123"
	// expect, _ := bcrypt.GenerateFromPassword([]byte(input), 1)
	hashedPassword, err := HashPassword(input)

	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input)) != nil {
		t.Errorf("%v", err)
	}
}
