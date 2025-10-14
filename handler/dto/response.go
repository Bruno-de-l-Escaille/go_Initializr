package dto

// Response represents a standard API response
type Response struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ResponseError represents an error response
type ResponseError struct {
	Error string `json:"error"`
}