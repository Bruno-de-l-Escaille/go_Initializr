package event

import (
	"encoding/json"
	"fmt"
	"go_Initializr/models"
	"go_Initializr/pkg/uuidv7"
	"go_Initializr/repository"
	eventsQ "go_Initializr/repository/queue"
	"go_Initializr/service"

	"github.com/bytedance/sonic"
	"github.com/rs/zerolog/log"
)

type EventService struct {
	Repository       repository.EventRepositoryInterface
	EntityRepository repository.BaseRepository
	EventPublisher   eventsQ.EventPublisher
}

// Ensure EventService implements EventServiceInterface
var _ service.EventServiceInterface = (*EventService)(nil)

// CreateAndProcessEvent creates a new event and processes it
func (s *EventService) CreateAndProcessEvent(entity, operation string, payload string, ref string, actorUUID string) (*models.Event, error) {
	event, err := s.createEvent(operation, payload, ref, actorUUID)
	if err != nil {
		return nil, err
	}
	if operation == "DELETE" {
		event.Payload = json.RawMessage("{}") // Set payload to empty JSON for DELETE operation
	}
	// Save the event
	err = s.EntityRepository.SaveEntityEvent(event, entity)
	if err != nil {
		return nil, err
	}

	// Publish event to Redis if publisher is available
	if s.EventPublisher != nil {
		// Create event type based on entity and operation
		eventType := fmt.Sprintf("%s_%s", entity, operation)
		eventData, err := sonic.Marshal(event)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal event for publishing")
		} else {

			// Create and publish the event using new format with UUID
			redisEvent := eventsQ.NewEvent(event.UUID, event.EntityUUID, eventType, eventData)
			// log.Info().Str("uuid", event.UUID).Msg("Publishing event to queue...")
			if err := s.EventPublisher.Publish(redisEvent); err != nil {
				log.Error().Err(err).Msg("Failed to publish event")
			}
		}
	}

	return event, nil
}

// CreateAndProcessEventsBulk creates and processes multiple events in a single operation
func (s *EventService) CreateAndProcessEventsBulk(entity string, operations []string, payloads []string, refs []string, actorUUID string) ([]*models.Event, error) {
	if len(operations) != len(payloads) || len(operations) != len(refs) {
		return nil, fmt.Errorf("operations, payloads, and refs must have the same length")
	}

	events := make([]*models.Event, len(operations))

	// Create all events first
	for i := range operations {
		event, err := s.createEvent(operations[i], payloads[i], refs[i], actorUUID)
		if err != nil {
			return nil, err
		}
		events[i] = event
	}

	// Publish events to Redis if publisher is available
	if s.EventPublisher != nil {
		for _, event := range events {
			// Create event type based on entity and operation
			eventType := fmt.Sprintf("%s_%s", entity, event.Operation)
			eventData, err := sonic.Marshal(event)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal event for publishing")
				continue
			}

			// Create and publish the event using new format with UUID
			redisEvent := eventsQ.NewEvent(event.UUID, event.EntityUUID, eventType, eventData)
			if err := s.EventPublisher.Publish(redisEvent); err != nil {
				log.Error().Err(err).Msg("Failed to publish event")
			}
		}
	}

	return events, nil
}

// Helper function to create an event
func (s *EventService) createEvent(operation string, payload string, ref string, actorUUID string) (*models.Event, error) {
	// Create base event properties
	uuid, _ := uuidv7.GenerateB62()

	// Create event - convert payload string to json.RawMessage
	event := &models.Event{
		UUID:      uuid,
		Operation: operation,
		Payload:   json.RawMessage(payload), // Convert string to json.RawMessage
		ActorUUID: actorUUID,
	}
	// Handle reference (entity UUID)
	if operation == "CREATE" {
		event.EntityUUID, _ = uuidv7.GenerateB62()
		// For CREATE operations during authentication, the actor should be the entity being created
		if actorUUID == "auth" {
			event.ActorUUID = event.EntityUUID
		}
	} else {
		event.EntityUUID = ref
		// For non-CREATE operations, if actorUUID is "auth", use the event UUID as fallback
		if actorUUID == "auth" {
			event.ActorUUID = uuid
		}
	}
	return event, nil
}
