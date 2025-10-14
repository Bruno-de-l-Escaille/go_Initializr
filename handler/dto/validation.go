package dto

// ValidationError represents validation errors for models
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
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