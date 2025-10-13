package repository

import (
	"database/sql"
	"fmt"
	"go_Initializr/models"
)

type EventRepository struct {
	DB *sql.DB
}

var _ EventRepositoryInterface = (*EventRepository)(nil)
// GetEventByUUIDAndType retrieves event data for a specific UUID and entity type
func (r *EventRepository) GetEventByUUIDAndType(uuid string, entityType string) (*models.Event, error) {
	// log.Info().Str("entity_type", entityType).Str("uuid", uuid).Msg("Getting event data")
	tableName := fmt.Sprintf("%s_events", entityType)

	// Define variables to scan into
	var eventUUID, operation string
	var payload []byte // Changed from string to []byte for json.RawMessage
	var entityUUID, actorUUID string

	// Update placeholder syntax from ? to $1 for PostgreSQL
	query := fmt.Sprintf("SELECT uuid, operation, payload, entity_uuid, actor_uuid FROM %s WHERE uuid = $1", tableName)
	err := r.DB.QueryRow(query, uuid).Scan(&eventUUID, &operation, &payload, &entityUUID, &actorUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event data from %s for uuid %s: %w", tableName, uuid, err)
	}

	// Construct the Event object
	event := &models.Event{
		UUID:       eventUUID,
		Operation:  operation,
		Payload:    payload, // Direct assignment of []byte to json.RawMessage
		EntityUUID: entityUUID,
		ActorUUID:  actorUUID,
	}

	return event, nil
}
