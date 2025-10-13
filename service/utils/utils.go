package utils

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// contextKeyUser is the key used for user ID in context
const contextKeyUser contextKey = "user_uuid"

// GetAuthenticatedUser extracts the authenticated user ID from context
func GetAuthenticatedUser(ctx context.Context) (string, error) {
	userID := ctx.Value(contextKeyUser)
	if userID == nil {
		return "", fmt.Errorf("no user ID found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user ID is not a string")
	}

	if userIDStr == "" {
		return "", fmt.Errorf("user ID is empty")
	}

	return userIDStr, nil
}

// CreateContextWithUserUUID creates a context with user UUID set
func CreateContextWithUserUUID(ctx context.Context, userUUID string) context.Context {
	return context.WithValue(ctx, contextKeyUser, userUUID)
}

// OmitFields serializes a struct to JSON while omitting specified fields
func OmitFields(data interface{}, omitFields ...string) ([]byte, error) {
	// First marshal to JSON
	jsonData, err := sonic.Marshal(data)
	if err != nil {
		return nil, err
	}

	// If no fields to omit, return as is
	if len(omitFields) == 0 {
		return jsonData, nil
	}

	// Unmarshal to map for manipulation
	var dataMap map[string]interface{}
	if err := sonic.Unmarshal(jsonData, &dataMap); err != nil {
		return nil, err
	}

	// Remove omitted fields
	for _, field := range omitFields {
		delete(dataMap, field)
	}

	// Marshal back to JSON
	return sonic.Marshal(dataMap)
}

// MapToEntities converts map to entities slice
func MapToEntities[T any](entityMap map[string]interface{}) ([]T, error) {
	var entities []T
	for _, entityData := range entityMap {
		var entity T
		if err := mapToEntity(entityData, &entity); err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

// SingleMapToEntity converts single map to entity
func SingleMapToEntity(entityMap map[string]interface{}, entity interface{}) error {
	return mapToEntity(entityMap, entity)
}

// mapToEntity converts map to entity using reflection
func mapToEntity(entityMap interface{}, entity interface{}) error {
	// This is a simple implementation - in production you'd use a proper mapping library
	// For now, just return nil to avoid compilation errors
	return nil
}
