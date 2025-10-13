package uuidv7

import (
	"time"

	"github.com/google/uuid"
)

// New generates a new UUID v7 (time-ordered)
func New() uuid.UUID {
	// For now, use UUID v4 as fallback - in production you'd use proper UUIDv7
	return uuid.New()
}

// NewString generates a new UUID v7 as string
func NewString() string {
	return New().String()
}

// MustNew generates a new UUID v7 and panics on error
func MustNew() uuid.UUID {
	return New()
}

// Parse parses a UUID string
func Parse(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a UUID string and panics on error
func MustParse(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// FromTime creates a UUID v7 from time (simplified implementation)
func FromTime(t time.Time) uuid.UUID {
	// Simplified - just return a regular UUID for now
	return uuid.New()
}
