package events

import "encoding/json"

// Event represents a generic event interface
type Event interface {
	GetUUID() string
	GetType() string
	GetData() []byte
	GetEntityUUID() string
}

// BaseEvent is a concrete implementation of the Event interface
type BaseEvent struct {
	UUID       string          `json:"uuid"`
	EntityUUID string          `json:"entity_uuid"`
	Type       string          `json:"type"`
	Data       json.RawMessage `json:"data,omitempty"`
}

// GetUUID returns the event UUID
func (e BaseEvent) GetUUID() string {
	return e.UUID
}

// GetType returns the event type
func (e BaseEvent) GetType() string {
	return e.Type
}

// GetData returns the event data
func (e BaseEvent) GetData() []byte {
	return e.Data
}

// GetEntityUUID returns the entity UUID
func (e BaseEvent) GetEntityUUID() string {
	return e.EntityUUID
}

// NewEvent creates a new event with the given UUID, type and data
func NewEvent(uuid string, entityUUID string, eventType string, data []byte) Event {
	return BaseEvent{
		UUID:       uuid,
		EntityUUID: entityUUID,
		Type:       eventType,
		Data:       data,
	}
}
