package apperrors

import "fmt"

// AppError represents an application error
type AppError struct {
	Message string
	Code    string
}

func (e *AppError) Error() string {
	return e.Message
}

// NewInternal creates a new internal error
func NewInternal() *AppError {
	return &AppError{
		Message: "internal error",
		Code:    "INTERNAL",
	}
}

// NewBadRequest creates a new bad request error
func NewBadRequest(message string) *AppError {
	return &AppError{
		Message: fmt.Sprintf("bad request: %s", message),
		Code:    "BAD_REQUEST",
	}
}

// NewAuthorization creates a new authorization error
func NewAuthorization(message string) *AppError {
	return &AppError{
		Message: fmt.Sprintf("authorization error: %s", message),
		Code:    "AUTHORIZATION",
	}
}

// NewNotFound creates a new not found error
func NewNotFound(message string) *AppError {
	return &AppError{
		Message: fmt.Sprintf("not found: %s", message),
		Code:    "NOT_FOUND",
	}
}
