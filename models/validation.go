package models

import (
	"time"

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
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	e.UpdatedAt = time.Now()
}

// ValidateCreate validates ExampleEntity for creation
func (e *ExampleEntityCreateRequest) ValidateCreate() []ValidationError {
	var errors []ValidationError

	if len(e.Name) < 2 || len(e.Name) > 100 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name must be between 2 and 100 characters",
		})
	}

	if len(e.Description) > 500 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "Description must be at most 500 characters",
		})
	}

	if e.Status != "" && e.Status != "active" && e.Status != "inactive" && e.Status != "pending" {
		errors = append(errors, ValidationError{
			Field:   "status",
			Message: "Status must be either 'active', 'inactive', or 'pending'",
		})
	}

	return errors
}
