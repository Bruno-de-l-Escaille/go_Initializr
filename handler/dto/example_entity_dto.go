package dto

// ExampleEntityCreateRequest represents the request body for creating an ExampleEntity
// swagger:model
type ExampleEntityCreateRequest struct {
	Name        string `json:"name" binding:"required" example:"Example Name" validation:"^[a-zA-Z\\s]{2,100}$" validation_message:"Name must be between 2 and 100 alphabetic characters."`
	Description string `json:"description" example:"This is an example description" validation:"^.{0,500}$" validation_message:"Description must be at most 500 characters."`
	Email       string `json:"email" binding:"required" example:"example@example.com" validation:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$" validation_message:"Please provide a valid email address."`
	Status      string `json:"status" example:"active" validation:"^(active|inactive|pending)$" validation_message:"Status must be either 'active', 'inactive', or 'pending'."`
}

// ExampleEntityUpdateRequest represents the request body for updating an ExampleEntity
// swagger:model
type ExampleEntityUpdateRequest struct {
	Name        *string `json:"name,omitempty" example:"Updated Example Name" validation:"^[a-zA-Z\\s]{2,100}$" validation_message:"Name must be between 2 and 100 alphabetic characters."`
	Description *string `json:"description,omitempty" example:"Updated description" validation:"^.{0,500}$" validation_message:"Description must be at most 500 characters."`
	Email       *string `json:"email,omitempty" example:"updated@example.com" validation:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$" validation_message:"Please provide a valid email address."`
	Status      *string `json:"status,omitempty" example:"inactive" validation:"^(active|inactive|pending)$" validation_message:"Status must be either 'active', 'inactive', or 'pending'."`
}