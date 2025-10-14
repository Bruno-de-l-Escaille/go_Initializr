-- Drop indexes first
DROP INDEX IF EXISTS gin_example_entity_events_name;
DROP INDEX IF EXISTS gin_example_entity_events_email;
DROP INDEX IF EXISTS gin_example_entity_events_payload;
DROP INDEX IF EXISTS idx_example_entity_events_created_at;
DROP INDEX IF EXISTS idx_example_entity_events_entity_uuid;

-- Drop the table
DROP TABLE IF EXISTS example_entity_events;