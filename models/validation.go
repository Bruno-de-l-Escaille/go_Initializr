package models

import (
	"golang.org/x/crypto/bcrypt"
)

// ValidationError represents validation errors for models
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks if a password matches the hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SetDefaults sets default values for ExampleEntity
func (e *ExampleEntity) SetDefaults() {
	if e.Status == "" {
		e.Status = "pending"
	}
}

