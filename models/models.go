package models

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


// Event represents an event in the system
type Event struct {
	UUID          string          `json:"uuid" example:"2zi4x7qG8dyx1q3q649iu"`                        // UUID if Event (UUID), if empty a new one will be generated
	EntityUUID    string          `json:"entity_uuid" example:"2zi4x7qG8dyx1q3q649iu"`                 // UUID of the entity this event relates to
	EntityType    string          `json:"entity_type" example:"example_entity"`                        // Type of the entity
	Operation     string          `json:"operation" enums:"create,update,delete" example:"update"`     // operation  ("create", "update", "delete")
	Payload       json.RawMessage `json:"-"`                                                           // Event payload (stored as JSON)
	Ref           string          `json:"ref" example:"example-ref"`                                   // Reference identifier
	ActorUUID     string          `json:"actor_uuid" example:"2zi4x7qG8dyx1q3q649iu"`                  // UUID of the actor who triggered this event
	UUIDTimeStamp time.Time       `json:"timestamp,omitempty" example:"2025-05-08T16:36:52.317+01:00"` // Timestamp extracted from UUID
}

// ExampleEntity represents an example entity in the system
// swagger:model
type ExampleEntity struct {
	UUID        string    `json:"uuid" bson:"uuid" swaggerignore:"true" validation:"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-7[0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$" validation_message:"Invalid UUID v7 format. Please provide a valid UUID v7."`
	Name        string    `json:"name" bson:"name" example:"Example Name" validation:"^[a-zA-Z\\s]{2,100}$" validation_message:"Name must be between 2 and 100 alphabetic characters."`
	Description string    `json:"description" bson:"description" example:"This is an example description" validation:"^.{0,500}$" validation_message:"Description must be at most 500 characters."`
	Email       string    `json:"email" bson:"email" example:"example@example.com" validation:"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$" validation_message:"Please provide a valid email address."`
	Status      string    `json:"status" bson:"status" example:"active" validation:"^(active|inactive|pending)$" validation_message:"Status must be either 'active', 'inactive', or 'pending'."`
}


// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserUUID string `json:"user_uuid"`
	jwt.RegisteredClaims
}
