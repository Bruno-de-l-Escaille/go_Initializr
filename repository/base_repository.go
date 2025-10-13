package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go_Initializr/models"
	"go_Initializr/models/apperrors"
	"go_Initializr/pkg/nosql"
	"go_Initializr/pkg/uuidv7"
	"go_Initializr/service/utils"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/bytedance/sonic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BaseRepository struct {
	DB          *sql.DB
	MongoClient *nosql.MongoDB
}

func getProjectionForType(targetType reflect.Type) (bson.M, error) {
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}
	if targetType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot generate projection for non-struct type: %v", targetType)
	}

	projection := bson.M{}
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		bsonTag := field.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		tagParts := strings.Split(bsonTag, ",")
		fieldName := tagParts[0]

		if fieldName == "" || fieldName == "_" {
			continue
		}

		projection[fieldName] = 1
	}

	hasExplicitID := false
	for key := range projection {
		if key == "_id" {
			hasExplicitID = true
			break
		}
	}
	if !hasExplicitID {
		// projection["_id"] = 0
	}

	if len(projection) == 0 {
		log.Warn().Str("type", targetType.String()).Msg("Generated projection is empty. Check BSON tags or struct complexity. Fetching all fields.")
		return nil, nil
	}
	return projection, nil
}

type FilterValidator[T any] struct {
	allowedFields map[string]reflect.Type
}

func NewFilterValidator[T any]() *FilterValidator[T] {
	var zero T
	targetType := reflect.TypeOf(zero)

	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	if targetType.Kind() != reflect.Struct {
		log.Warn().Str("type", targetType.String()).Msg("FilterValidator created for non-struct type")
		return &FilterValidator[T]{
			allowedFields: make(map[string]reflect.Type),
		}
	}

	allowedFields := make(map[string]reflect.Type)

	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		bsonTag := field.Tag.Get("bson")

		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		tagParts := strings.Split(bsonTag, ",")
		fieldName := tagParts[0]

		if fieldName == "" || fieldName == "_" {
			continue
		}

		allowedFields[fieldName] = field.Type
	}

	return &FilterValidator[T]{
		allowedFields: allowedFields,
	}
}

func (fv *FilterValidator[T]) ValidateFilter(filter bson.M) error {
	if filter == nil {
		return nil
	}
	for fieldName := range filter {
		if _, exists := fv.allowedFields[fieldName]; !exists {
			err := "field " + fieldName + " is not allowed for filtering"
			return apperrors.NewBadRequest(err)
		}
	}
	return nil
}

func GetAllEntities[T any](ctx context.Context, r *BaseRepository, entityType string, filter bson.M, offset *int, limit *int) ([]T, int64, int64, error) {
	startTime := time.Now()
	var results []T

	if filter == nil {
		filter = bson.M{}
	} else {
		validator := NewFilterValidator[T]()
		if err := validator.ValidateFilter(filter); err != nil {
			return nil, 0, 0, err
		}
	}

	// Get total count without pagination
	collection := r.MongoClient.Database.Collection(entityType)
	countStart := time.Now()
	totalCount, err := collection.CountDocuments(ctx, filter)
	countElapsed := time.Since(countStart)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error counting documents: %w", err)
	}
	log.Info().Str("entity_type", entityType).Dur("duration", countElapsed).Int64("count", totalCount).Msg("CountDocuments")

	if *offset == -1 && *limit == -1 {
		*offset = 0
		*limit = int(totalCount)
	}
	// Check if offset is too large
	if offset != nil && *offset >= int(totalCount) && totalCount > 0 {
		return nil, 0, totalCount, apperrors.NewBadRequest(fmt.Sprintf("Offset %d is greater than total count %d", *offset, totalCount))
	}

	// Build pipeline stages
	var pipeline []bson.M
	pipeline = append(pipeline, bson.M{"$match": filter})

	// Add projection if available
	elemType := reflect.TypeOf(results).Elem()
	projection, projErr := getProjectionForType(elemType)
	if projErr != nil {
		log.Error().Err(projErr).Str("type", elemType.String()).Msg("Error generating projection")
	} else if projection != nil {
		pipeline = append(pipeline, bson.M{"$project": projection})
	}

	// Add pagination stages if defined
	// If both offset and limit are nil, we return all documents
	if offset != nil || limit != nil {
		if offset != nil {
			pipeline = append(pipeline, bson.M{"$skip": *offset})
		}
		if limit != nil {
			pipeline = append(pipeline, bson.M{"$limit": *limit})
		}
	}

	defer func() {
		elapsed := time.Since(startTime)
		var zero T
		log.Info().Str("type", reflect.TypeOf(zero).String()).Dur("duration", elapsed).Bool("projection_used", projection != nil).Msg("BaseRepo.GetAllEntities")
	}()

	// Execute pipeline
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error aggregating entities, falling back to InitializeAllSnapshots")
		fallbackResults, fallbackErr := fallbackInitializeAndMap[T](ctx, r, entityType)
		if fallbackErr != nil {
			return nil, 0, 0, fallbackErr
		}
		resultCount := int64(len(fallbackResults))
		return fallbackResults, resultCount, resultCount, nil
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &results); err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error decoding aggregation result")
		return nil, 0, 0, fmt.Errorf("failed to decode entities [%s]: %w", entityType, err)
	}

	resultCount := int64(len(results))
	log.Info().Int64("result_count", resultCount).Str("entity_type", entityType).Int64("total_count", totalCount).Msg("Found entities")

	return results, resultCount, totalCount, nil
}

func GetEntityByRef[T any](ctx context.Context, r *BaseRepository, ref string, entityType string) (*T, error) {
	// startTime := time.Now()
	var result T
	targetType := reflect.TypeOf(result)

	projection, projErr := getProjectionForType(targetType)
	if projErr != nil {
		log.Error().Err(projErr).Str("type", targetType.String()).Msg("Error generating projection")
		projection = nil
	} else if projection == nil {
		log.Warn().Str("type", targetType.String()).Msg("No projection generated. Fetching all fields.")
	}

	// defer func() {
	// 	elapsed := time.Since(startTime)
	// 	log.Info().Str("type", reflect.TypeOf(result).String()).Str("ref", ref).Dur("duration", elapsed).Bool("projection_used", projection != nil).Msg("BaseRepo.GetEntityByRef")
	// }()

	collection := r.MongoClient.Database.Collection(entityType)
	findOneOptions := options.FindOne()
	if projection != nil {
		findOneOptions.SetProjection(projection)
	}
	var filter = bson.M{"uuid": ref}

	err := collection.FindOne(ctx, filter, findOneOptions).Decode(&result)
	if err == nil {
		return &result, nil
	}
	log.Error().Err(err).Str("entity_type", entityType).Str("ref", ref).Msg("Mongo error while getting entity")

	if !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("entity_type", entityType).Str("ref", ref).Msg("MongoDB FindOne error")
		return nil, fmt.Errorf("database error retrieving snapshot [%s] '%s': %w", entityType, ref, err)
	}

	log.Info().Str("entity_type", entityType).Str("ref", ref).Msg("Entity not found in snapshot, reconstructing from events...")
	fallbackStart := time.Now()
	reconstructedMap, reconErr := r.reconstructEntityMapFromEvents(ctx, ref, entityType)
	if reconErr != nil {
		if errors.Is(reconErr, mongo.ErrNoDocuments) {
			log.Warn().Str("entity_type", entityType).Str("ref", ref).Msg("Entity not found during event reconstruction")
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(reconErr).Str("entity_type", entityType).Str("ref", ref).Msg("Error reconstructing from events")
		return nil, fmt.Errorf("failed during event reconstruction for [%s] '%s': %w", entityType, ref, reconErr)
	}

	var finalEntity T
	mapErr := utils.SingleMapToEntity(reconstructedMap, &finalEntity)
	if mapErr != nil {
		log.Error().Err(mapErr).Str("type", reflect.TypeOf(finalEntity).String()).Str("entity_type", entityType).Str("ref", ref).Msg("Error mapping reconstructed map")
		return nil, fmt.Errorf("failed to map reconstructed entity [%s] '%s': %w", entityType, ref, mapErr)
	}
	fallbackElapsed := time.Since(fallbackStart)
	log.Info().Str("entity_type", entityType).Str("ref", ref).Dur("duration", fallbackElapsed).Msg("Reconstructed and mapped entity from events")

	return &finalEntity, nil
}

func fallbackInitializeAndMap[T any](ctx context.Context, r *BaseRepository, entityType string) ([]T, error) {
	log.Info().Str("entity_type", entityType).Msg("Executing fallback: InitializeAllSnapshots")
	entityMap, initErr := r.InitializeAllSnapshots(ctx, entityType)
	if initErr != nil {
		log.Error().Err(initErr).Str("entity_type", entityType).Msg("Error during InitializeAllSnapshots fallback")
		return nil, initErr
	}

	mapStart := time.Now()
	entitiesFromMap, mapErr := utils.MapToEntities[T](entityMap)
	if mapErr != nil {
		var zero T
		log.Error().Err(mapErr).Str("type", reflect.TypeOf(zero).String()).Str("entity_type", entityType).Msg("Error mapping fallback results")
		return nil, fmt.Errorf("error mapping fallback results for [%s]: %w", entityType, mapErr)
	}
	mapElapsed := time.Since(mapStart)
	log.Info().Int("count", len(entitiesFromMap)).Str("entity_type", entityType).Dur("duration", mapElapsed).Msg("Mapped entities from fallback map")

	return entitiesFromMap, nil
}

func (r *BaseRepository) reconstructEntityMapFromEvents(ctx context.Context, ref string, entityType string) (map[string]interface{}, error) {
	var rows *sql.Rows
	var queryErr error

	query := fmt.Sprintf(`
        SELECT uuid, operation, payload, entity_uuid
        FROM %s_events
        WHERE %%s = $1
        ORDER BY uuid ASC
    `, entityType)

	targetUUID := ref
	formattedQuery := fmt.Sprintf(query, "entity_uuid")
	rows, queryErr = r.DB.QueryContext(ctx, formattedQuery, targetUUID)
	// log.Info().Str("entity_type", entityType).Str("ref", ref).Msg("Querying events by UUID")

	if queryErr != nil {
		log.Error().Err(queryErr).Str("entity_type", entityType).Str("ref", ref).Msg("SQL Query error reconstructing")
		return nil, fmt.Errorf("database error retrieving events: %w", queryErr)
	}
	defer rows.Close()

	entityState := make(map[string]interface{})
	foundEvents := false
	processedEventCount := 0
	relevantEventCount := 0

	for rows.Next() {
		processedEventCount++
		var event models.Event
		if err := rows.Scan(
			&event.UUID, &event.Operation,
			&event.Payload, &event.EntityUUID); err != nil {
			log.Error().Err(err).Str("entity_type", entityType).Str("ref", ref).Int("event_count", processedEventCount).Msg("SQL Scan error reconstructing")
			return nil, fmt.Errorf("database error scanning event: %w", err)
		}

		if event.EntityUUID == targetUUID {
			if !foundEvents {
				foundEvents = true
			}
			relevantEventCount++
			r.applySingleEvent(entityState, event)

			if strings.ToUpper(event.Operation) == "DELETE" {
				log.Info().Str("entity_type", entityType).Str("ref", ref).Msg("DELETE operation found, stopping event processing")
				return nil, mongo.ErrNoDocuments
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Str("ref", ref).Msg("SQL rows iteration error reconstructing")
		return nil, fmt.Errorf("database error iterating events: %w", err)
	}

	if !foundEvents {
		log.Warn().Str("entity_type", entityType).Str("ref", ref).Msg("No relevant events found to reconstruct state")
		return nil, mongo.ErrNoDocuments
	}

	if _, exists := entityState["uuid"]; !exists && targetUUID != "" {
		entityState["uuid"] = targetUUID
	}

	log.Info().Str("entity_type", entityType).Str("ref", ref).Int("relevant_events", relevantEventCount).Int("scanned_events", processedEventCount).Msg("Reconstructed state from events")
	return entityState, nil
}

func (r *BaseRepository) applySingleEvent(entityState map[string]interface{}, event models.Event) {
	op := strings.ToUpper(event.Operation)

	if op == "DELETE" {
		for k := range entityState {
			delete(entityState, k)
		}
		return
	}

	var payload map[string]interface{}
	if len(event.Payload) > 0 && event.Payload != nil {
		if err := sonic.Unmarshal([]byte(event.Payload), &payload); err != nil {
			log.Warn().Err(err).Str("event_uuid", event.UUID).Msg("Failed to unmarshal payload during single entity reconstruction")
		}
	}

	if payload != nil {
		for key, value := range payload {
			entityState[key] = value
		}
	}

	entityState["uuid"] = event.EntityUUID
}

func (r *BaseRepository) applyEventToEntityMap(entityMap map[string]interface{}, event models.Event) {
	entityUUID := event.EntityUUID
	op := strings.ToUpper(event.Operation)

	if op == "DELETE" {
		delete(entityMap, entityUUID)
		return
	}

	var currentState map[string]interface{}
	if existingState, exists := entityMap[entityUUID]; exists {
		var ok bool
		currentState, ok = existingState.(map[string]interface{})
		if !ok {
			log.Error().Str("entity_uuid", entityUUID).Str("type", reflect.TypeOf(existingState).String()).Msg("Existing state is not map[string]interface{}, resetting")
			currentState = make(map[string]interface{})
		}
	} else {
		currentState = make(map[string]interface{})
	}

	var payload map[string]interface{}
	if len(event.Payload) > 0 && event.Payload != nil {
		if err := sonic.Unmarshal([]byte(event.Payload), &payload); err != nil {
			log.Warn().Err(err).Str("event_uuid", event.UUID).Msg("Failed to unmarshal event payload during bulk init")
		}
	}

	if payload != nil {
		for key, value := range payload {
			currentState[key] = value
		}
	}

	currentState["uuid"] = entityUUID
	entityMap[entityUUID] = currentState
}

func (r *BaseRepository) GetLastSnapshotUUID(ctx context.Context, entityType string) (string, error) {
	snapshotMetaCollection := r.MongoClient.Database.Collection("snapshot_markers")
	var marker struct {
		ID            string    `bson:"_id"`
		LastEventUUID string    `bson:"last_event_uuid"`
		UpdatedAt     time.Time `bson:"updated_at"`
	}
	filter := bson.M{"_id": entityType}
	err := snapshotMetaCollection.FindOne(ctx, filter).Decode(&marker)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Info().Str("entity_type", entityType).Msg("No snapshot marker found")
		return "", nil
	}
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error getting snapshot marker")
		return "", err
	}
	return marker.LastEventUUID, nil
}

func (r *BaseRepository) UpdateLastSnapshotUUID(ctx context.Context, entityType string, uuid string) error {
	snapshotMetaCollection := r.MongoClient.Database.Collection("snapshot_markers")
	filter := bson.M{"_id": entityType}
	update := bson.M{
		"$set": bson.M{
			"last_event_uuid": uuid,
			"updated_at":      time.Now().UTC(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := snapshotMetaCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Str("uuid", uuid).Msg("Error updating snapshot marker")
		return err
	}
	log.Info().Str("entity_type", entityType).Str("uuid", uuid).Msg("Updated snapshot marker")
	return nil
}

func (r *BaseRepository) UpdateEntitySnapshot(ctx context.Context, entityData interface{}, entityType string) error {
	collection := r.MongoClient.Database.Collection(entityType)

	entityMap, ok := entityData.(map[string]interface{})
	if !ok {
		log.Error().Str("type", reflect.TypeOf(entityData).String()).Str("entity_type", entityType).Msg("UpdateEntitySnapshot expects map[string]interface{}")
		return fmt.Errorf("UpdateEntitySnapshot expects map[string]interface{}, got %T", entityData)
	}

	uuidVal, exists := entityMap["uuid"]
	if !exists {
		log.Error().Str("entity_type", entityType).Msg("Entity map missing 'uuid' key in UpdateEntitySnapshot")
		return fmt.Errorf("entity map submitted to UpdateEntitySnapshot has no 'uuid' key")
	}
	uuid, ok := uuidVal.(string)
	if !ok || uuid == "" {
		log.Error().Interface("uuid_val", uuidVal).Str("entity_type", entityType).Msg("Entity map 'uuid' is not a valid string in UpdateEntitySnapshot")
		return fmt.Errorf("entity map 'uuid' is not a valid string")
	}

	filter := bson.M{"uuid": uuid}
	update := bson.M{"$set": entityMap}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Str("uuid", uuid).Msg("Error updating/inserting snapshot")
		return err
	}

	return nil
}

func (r *BaseRepository) DeleteEntitySnapshot(ctx context.Context, uuid string, entityType string) error {
	collection := r.MongoClient.Database.Collection(entityType)
	filter := bson.M{"uuid": uuid}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Str("uuid", uuid).Msg("Error deleting snapshot")
		return err
	}
	if result.DeletedCount == 0 {
		log.Warn().Str("entity_type", entityType).Str("uuid", uuid).Msg("DeleteEntitySnapshot called, but no document was found")
	} else {
		log.Info().Str("entity_type", entityType).Str("uuid", uuid).Msg("Successfully deleted snapshot")
	}
	return nil
}

func (r *BaseRepository) ApplyEventsToEntityMap(ctx context.Context, entityMap map[string]interface{}, lastSnapshottedUUID string, entityType string) (int, error) {
	query := fmt.Sprintf(`
        SELECT uuid, operation, payload, entity_uuid
        FROM %s_events
        WHERE uuid > $1
        ORDER BY uuid ASC
    `, entityType)

	if lastSnapshottedUUID == "" {
		log.Warn().Str("entity_type", entityType).Msg("No lastSnapshottedUUID provided, cannot apply incremental events")
		return 0, nil
	}

	log.Info().Str("entity_type", entityType).Str("after_uuid", lastSnapshottedUUID).Msg("Applying events")
	rows, err := r.DB.QueryContext(ctx, query, lastSnapshottedUUID)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Str("after_uuid", lastSnapshottedUUID).Msg("Error querying events")
		return 0, err
	}
	defer rows.Close()

	nbEventsApplied := 0
	currentLastUUID := lastSnapshottedUUID

	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.UUID, &event.Operation,
			&event.Payload, &event.EntityUUID); err != nil {
			log.Error().Err(err).Str("entity_type", entityType).Msg("Error scanning event row while applying events")
			continue
		}

		r.applyEventToEntityMap(entityMap, event)
		currentLastUUID = event.UUID
		nbEventsApplied++
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error iterating event rows while applying events")
		return nbEventsApplied, err
	}

	if nbEventsApplied > 0 {
		if err := r.UpdateLastSnapshotUUID(ctx, entityType, currentLastUUID); err != nil {
			log.Warn().Err(err).Str("entity_type", entityType).Str("uuid", currentLastUUID).Int("applied_events", nbEventsApplied).Msg("Failed to update last snapshot UUID marker")
		}
	}

	log.Info().Int("applied_events", nbEventsApplied).Str("entity_type", entityType).Str("last_uuid", currentLastUUID).Msg("Applied new events to map")
	return nbEventsApplied, nil
}

func (r *BaseRepository) GetEntityHistoryByRef(ctx context.Context, ref string, entityType string) ([]models.Event, error) {
	var rows *sql.Rows
	var queryErr error

	query := fmt.Sprintf(`
SELECT uuid, operation, payload, entity_uuid, actor_uuid
FROM %s_events
WHERE %%s = $1
ORDER BY uuid ASC
`, entityType)
	formattedQuery := fmt.Sprintf(query, "entity_uuid")
	rows, queryErr = r.DB.QueryContext(ctx, formattedQuery, ref)
	if queryErr != nil {
		log.Error().Err(queryErr).Str("ref", ref).Str("entity_type", entityType).Msg("Error querying history")
		return nil, queryErr
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.UUID, &event.Operation,
			&event.Payload, &event.EntityUUID, &event.ActorUUID); err != nil {
			log.Error().Err(err).Str("ref", ref).Str("entity_type", entityType).Msg("Error scanning history event row")
			return nil, err
		}
		event.UUIDTimeStamp, _ = uuidv7.GetTimestampFromBase62UUIDv7(event.UUID)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Str("ref", ref).Str("entity_type", entityType).Msg("Error iterating history event rows")
		return nil, err
	}

	if len(events) == 0 {
		log.Info().Str("ref", ref).Str("entity_type", entityType).Msg("No history found")
	}
	return events, nil
}

func (r *BaseRepository) GetAllEntityEvents(ctx context.Context, entityType string) (map[string]interface{}, error) {
	startTime := time.Now()

	query := fmt.Sprintf(`
        SELECT uuid, operation,payload, entity_uuid
        FROM %s_events
        ORDER BY uuid ASC
    `, entityType)

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error querying all events")
		return nil, err
	}
	defer rows.Close()

	entityMap := make(map[string]interface{})
	eventCount := 0
	var lastProcessedUUID string
	for rows.Next() {
		eventCount++
		var event models.Event
		if err := rows.Scan(
			&event.UUID, &event.Operation,
			&event.Payload, &event.EntityUUID); err != nil {
			log.Error().Err(err).Int("event_count", eventCount).Str("entity_type", entityType).Msg("Error scanning event row during GetAllEntityEvents")
			return nil, fmt.Errorf("failed scanning event %d: %w", eventCount, err)
		}
		r.applyEventToEntityMap(entityMap, event)
		lastProcessedUUID = event.UUID
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error iterating event rows during GetAllEntityEvents")
		return nil, err
	}

	elapsed := time.Since(startTime)
	log.Info().Str("entity_type", entityType).Dur("duration", elapsed).Int("event_count", eventCount).Int("entity_count", len(entityMap)).Str("last_uuid", lastProcessedUUID).Msg("GetAllEntityEvents execution time")

	return entityMap, nil
}

func (r *BaseRepository) GetLastEventUUID(ctx context.Context, entityType string) (string, error) {
	var lastUUID sql.NullString
	query := fmt.Sprintf("SELECT uuid FROM %s_events ORDER BY uuid DESC LIMIT 1", entityType)
	err := r.DB.QueryRowContext(ctx, query).Scan(&lastUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error getting last event UUID")
		return "", err
	}
	if !lastUUID.Valid {
		return "", nil
	}
	return lastUUID.String, nil
}

func (r *BaseRepository) SaveEntityEvent(event *models.Event, entityType string) error {
	// Check if the last event for this entity has the same payload
	latestEventQuery := fmt.Sprintf(`
		SELECT payload
		FROM %s_events
		WHERE entity_uuid = $1
		ORDER BY uuid DESC
		LIMIT 1
	`, entityType)

	var latestPayload []byte
	err := r.DB.QueryRow(latestEventQuery, event.EntityUUID).Scan(&latestPayload)

	if err == nil {
		// A previous event was found, now compare payloads
		if string(latestPayload) == string(event.Payload) {
			log.Info().Str("entity_uuid", event.EntityUUID).Str("event_uuid", event.UUID).Msg("Redundant event detected. Skipping insertion.")
			return nil // Return nil to prevent error propagation for a legitimate skip
		}
	} else if err != sql.ErrNoRows {
		// An actual error occurred during the query
		log.Error().Err(err).Str("entity_uuid", event.EntityUUID).Str("entity_type", entityType).Msg("Error checking for last event")
		return err
	}
	// If err is sql.ErrNoRows, it's the first event for this entity, so we proceed.

	// No duplicate found, proceed with insertion.
	insertQuery := fmt.Sprintf("INSERT INTO %s_events (uuid, operation, payload, entity_uuid, actor_uuid) VALUES ($1, $2, $3, $4, $5)", entityType)
	_, insertErr := r.DB.Exec(insertQuery, event.UUID, event.Operation, event.Payload, event.EntityUUID, event.ActorUUID)
	if insertErr != nil {
		log.Error().Err(insertErr).Str("event_uuid", event.UUID).Str("entity_uuid", event.EntityUUID).Str("entity_type", entityType).Msg("Error saving event")
		return insertErr
	}

	return nil
}

func (r *BaseRepository) GetEntitySnapshotByUUID(ctx context.Context, uuid string, entityType string, result interface{}) error {
	if result == nil {
		return errors.New("result argument cannot be nil")
	}
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("result argument must be a non-nil pointer")
	}

	collection := r.MongoClient.Database.Collection(entityType)
	filter := bson.M{"uuid": uuid}
	err := collection.FindOne(ctx, filter).Decode(result)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Error().Err(err).Str("uuid", uuid).Str("entity_type", entityType).Msg("Error getting snapshot by UUID")
	}
	return err
}

func (r *BaseRepository) InitializeAllSnapshots(ctx context.Context, entityType string) (map[string]interface{}, error) {
	log.Info().Str("entity_type", entityType).Msg("Starting InitializeAllSnapshots")

	entityMap, err := r.GetAllEntityEvents(ctx, entityType)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error in GetAllEntityEvents during InitializeAllSnapshots")
		return nil, fmt.Errorf("failed to get all entity events: %w", err)
	}

	lastEventUUIDInEvents, eventUUIDErr := r.GetLastEventUUID(ctx, entityType)
	if eventUUIDErr != nil {
		log.Warn().Err(eventUUIDErr).Str("entity_type", entityType).Msg("Could not determine last event UUID after GetAllEntityEvents")
	}

	if len(entityMap) == 0 {
		log.Info().Str("entity_type", entityType).Msg("InitializeAllSnapshots: No entities found after processing events. No snapshots to update.")
		if markerErr := r.UpdateLastSnapshotUUID(ctx, entityType, lastEventUUIDInEvents); markerErr != nil {
			log.Warn().Err(markerErr).Str("entity_type", entityType).Msg("Failed to update snapshot marker for empty result set")
		}
		return entityMap, nil
	}

	log.Info().Str("entity_type", entityType).Int("entity_count", len(entityMap)).Msg("Obtained event map, starting ultra-fast MongoDB snapshot updates")
	startTimeUpdate := time.Now()

	collection := r.MongoClient.Database.Collection(entityType)

	// Convert map to documents
	documents := make([]interface{}, 0, len(entityMap))
	for _, entityData := range entityMap {
		if entityMapConv, ok := entityData.(map[string]interface{}); ok {
			documents = append(documents, entityMapConv)
		}
	}

	processedCount := 0
	errorCount := 0

	if len(documents) > 0 {
		// Drop the collection first (fastest way to clear all data)
		if err := collection.Drop(ctx); err != nil {
			log.Warn().Err(err).Str("entity_type", entityType).Msg("Could not drop collection")
			// Continue anyway, maybe collection didn't exist
		}

		// Insert all documents at once - much faster than upserts
		opts := options.InsertMany().
			SetOrdered(false).
			SetBypassDocumentValidation(true)

		_, insertErr := collection.InsertMany(ctx, documents, opts)
		if insertErr != nil {
			log.Error().Err(insertErr).Str("entity_type", entityType).Msg("InsertMany error during InitializeAllSnapshots")
			errorCount = len(documents)

			// Fallback to optimized upsert if insert fails
			log.Info().Str("entity_type", entityType).Msg("Falling back to upsert method")

			operations := make([]mongo.WriteModel, 0, len(entityMap))
			for _, entityData := range entityMap {
				entityMapConv, ok := entityData.(map[string]interface{})
				if !ok {
					continue
				}

				filter := bson.M{"uuid": entityMapConv["uuid"]}
				op := mongo.NewReplaceOneModel().
					SetFilter(filter).
					SetReplacement(entityMapConv).
					SetUpsert(true)

				operations = append(operations, op)
			}

			if len(operations) > 0 {
				bulkOpts := options.BulkWrite().
					SetOrdered(false).
					SetBypassDocumentValidation(true)

				_, bulkErr := collection.BulkWrite(ctx, operations, bulkOpts)
				if bulkErr != nil {
					log.Error().Err(bulkErr).Str("entity_type", entityType).Msg("Fallback BulkWrite error during InitializeAllSnapshots")
					errorCount = len(operations)
				} else {
					processedCount = len(operations)
					errorCount = 0
				}
			}
		} else {
			processedCount = len(documents)
			errorCount = 0

			// Create UUID index in background for performance (don't wait for it)
			go func() {
				indexModel := mongo.IndexModel{
					Keys:    bson.D{{Key: "uuid", Value: 1}},
					Options: options.Index().SetUnique(true).SetBackground(true),
				}
				_, indexErr := collection.Indexes().CreateOne(context.Background(), indexModel)
				if indexErr != nil {
					log.Warn().Err(indexErr).Str("entity_type", entityType).Msg("Could not create UUID index")
				}
			}()
		}
	}

	if markerErr := r.UpdateLastSnapshotUUID(ctx, entityType, lastEventUUIDInEvents); markerErr != nil {
		log.Warn().Err(markerErr).Str("entity_type", entityType).Str("uuid", lastEventUUIDInEvents).Msg("Failed to update snapshot marker after InitializeAllSnapshots")
	}

	elapsedUpdate := time.Since(startTimeUpdate)
	log.Info().Int("processed_count", processedCount).Str("entity_type", entityType).Dur("duration", elapsedUpdate).Int("error_count", errorCount).Str("last_event_uuid", lastEventUUIDInEvents).Msg("InitializeAllSnapshots finished")

	return entityMap, nil
}

var _ EntityRepositoryInterface = (*BaseRepository)(nil)

func GetEntitySnapshotByUUIDAndType[T any](ctx context.Context, r *BaseRepository, uuid string, entityType string) (*T, error) {
	var result T
	collection := r.MongoClient.Database.Collection(entityType)
	filter := bson.M{"uuid": uuid}

	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn().Str("entity_type", entityType).Str("uuid", uuid).Msg("Entity not found in snapshot")
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("entity_type", entityType).Str("uuid", uuid).Msg("Error retrieving entity from snapshot")
		return nil, err
	}
	return &result, nil
}

func GetEntitiesByFilter[T any](ctx context.Context, r *BaseRepository, entityType string, filter bson.M) ([]T, error) {
	// startTime := time.Now()
	var results []T

	elemType := reflect.TypeOf(results).Elem()
	projection, projErr := getProjectionForType(elemType)
	if projErr != nil {
		log.Error().Err(projErr).Str("type", elemType.String()).Msg("Error generating projection")
		projection = nil
	}

	// defer func() {
	// 	elapsed := time.Since(startTime)
	// 	var zero T
	// 	log.Info().Str("type", reflect.TypeOf(zero).String()).Str("entity_type", entityType).Dur("duration", elapsed).Msg("GetEntitiesByFilter")
	// }()

	collection := r.MongoClient.Database.Collection(entityType)
	findOptions := options.Find()
	if projection != nil {
		findOptions.SetProjection(projection)
	}

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error finding entities with filter")
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &results); err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error decoding entities with filter")
		return nil, fmt.Errorf("failed to decode entities [%s]: %w", entityType, err)
	}

	return results, nil
}

func CountEntities[T any](ctx context.Context, r *BaseRepository, entityType string, filter bson.M) (int64, error) {
	collection := r.MongoClient.Database.Collection(entityType)
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error counting entities with filter")
		return 0, err
	}
	return count, nil
}

func FindOneEntity[T any](ctx context.Context, r *BaseRepository, entityType string, filter bson.M) (*T, error) {
	var result T

	targetType := reflect.TypeOf(result)
	projection, projErr := getProjectionForType(targetType)
	if projErr != nil {
		log.Error().Err(projErr).Str("type", targetType.String()).Msg("Error generating projection")
		projection = nil
	}

	collection := r.MongoClient.Database.Collection(entityType)
	findOneOptions := options.FindOne()
	if projection != nil {
		findOneOptions.SetProjection(projection)
	}

	err := collection.FindOne(ctx, filter, findOneOptions).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		log.Error().Err(err).Str("entity_type", entityType).Msg("Error finding one entity with filter")
		return nil, err
	}

	return &result, nil
}

func MapEntityMapToSlice[T any](entityMap map[string]interface{}) ([]T, error) {
	return utils.MapToEntities[T](entityMap)
}

func (r *BaseRepository) CheckSnapshotAndEventStoreLastUUIDs(ctx context.Context, entityType string) (bool, string, error) {
	startTime := time.Now()
	lastSnapshotUUID, err := r.GetLastSnapshotUUID(ctx, entityType)

	if err != nil || lastSnapshotUUID == "" {
		log.Error().Err(err).Msg("Error getting last snapshot UUID")
		return false, "", err
	}
	lastEventStoreUUID, err := r.GetLastEventUUID(ctx, entityType)
	if err != nil {
		log.Error().Err(err).Msg("Error getting last event UUID")
		return false, "", err
	}
	elapsed := time.Since(startTime)
	log.Info().Str("entity_type", entityType).Dur("duration", elapsed).Msg("CheckSnapshotAndEventStoreLastUUIDs")
	return lastSnapshotUUID == lastEventStoreUUID, lastSnapshotUUID, nil
}

func (r *BaseRepository) UpdateSnapshotSinceEventCuttOff(ctx context.Context, entityType string, eventCutOffUuid string) error {
	startTime := time.Now()

	if eventCutOffUuid == "" {
		return fmt.Errorf("empty event cutoff UUID provided")
	}

	log.Info().Str("entity_type", entityType).Str("since_uuid", eventCutOffUuid).Msg("Updating snapshots with events")

	query := fmt.Sprintf(`
        SELECT uuid, operation, payload, entity_uuid
        FROM %s_events
        WHERE uuid > $1
        ORDER BY uuid ASC
    `, entityType)

	rows, err := r.DB.QueryContext(ctx, query, eventCutOffUuid)
	if err != nil {
		log.Error().Err(err).Str("entity_type", entityType).Str("since_uuid", eventCutOffUuid).Msg("Error querying events")
		return err
	}
	defer rows.Close()

	affectedEntities := make(map[string]bool)
	var events []models.Event
	var lastEventUUID string

	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.UUID, &event.Operation,
			&event.Payload, &event.EntityUUID); err != nil {
			log.Error().Err(err).Msg("Error scanning event row")
			continue
		}

		events = append(events, event)
		affectedEntities[event.EntityUUID] = true
		lastEventUUID = event.UUID
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating events: %w", err)
	}

	if len(events) == 0 {
		log.Info().Str("entity_type", entityType).Str("since_uuid", eventCutOffUuid).Msg("No new events found")
		return nil
	}

	log.Info().Int("event_count", len(events)).Int("affected_entities", len(affectedEntities)).Str("entity_type", entityType).Str("since_uuid", eventCutOffUuid).Msg("Found events")

	collection := r.MongoClient.Database.Collection(entityType)
	var operations []mongo.WriteModel

	for entityUUID := range affectedEntities {
		var currentState map[string]interface{}
		err := collection.FindOne(ctx, bson.M{"uuid": entityUUID}).Decode(&currentState)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			log.Error().Err(err).Str("entity_uuid", entityUUID).Msg("Error retrieving entity")
			continue
		}

		if currentState == nil {
			currentState = make(map[string]interface{})
			currentState["uuid"] = entityUUID
		}

		entityEventsApplied := 0
		for _, event := range events {
			if event.EntityUUID == entityUUID {
				r.applySingleEvent(currentState, event)
				entityEventsApplied++
			}
		}

		if len(currentState) == 0 || (len(currentState) == 1 && currentState["uuid"] != nil) {
			filter := bson.M{"uuid": entityUUID}
			op := mongo.NewDeleteOneModel().SetFilter(filter)
			operations = append(operations, op)
			continue
		}

		filter := bson.M{"uuid": entityUUID}
		update := bson.M{"$set": currentState}
		op := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		operations = append(operations, op)
	}

	if len(operations) > 0 {
		bulkOption := options.BulkWriteOptions{}
		bulkOption.SetOrdered(false)

		result, bulkErr := collection.BulkWrite(ctx, operations, &bulkOption)
		if bulkErr != nil {
			log.Error().Err(bulkErr).Str("entity_type", entityType).Msg("BulkWrite error during UpdateSnapshotSinceEvent")
			return fmt.Errorf("failed to update snapshots: %w", bulkErr)
		}

		log.Info().Str("entity_type", entityType).Int64("modified", result.ModifiedCount).Int64("inserted", result.UpsertedCount).Int64("deleted", result.DeletedCount).Msg("Bulk update")
	}

	if lastEventUUID != "" {
		if err := r.UpdateLastSnapshotUUID(ctx, entityType, lastEventUUID); err != nil {
			log.Warn().Err(err).Str("uuid", lastEventUUID).Msg("Failed to update snapshot marker")
		}
	}
	elapsed := time.Since(startTime)
	log.Info().Str("entity_type", entityType).Int("event_count", len(events)).Int("affected_entities", len(affectedEntities)).Str("since_uuid", eventCutOffUuid).Dur("duration", elapsed).Msg("Updated snapshots")

	return nil
}
