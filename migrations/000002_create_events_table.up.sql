-- Create example_entity_events table (following gopeople pattern)
CREATE TABLE IF NOT EXISTS example_entity_events (
    uuid VARCHAR(22) PRIMARY KEY,  -- Base62 encoded UUID v7 (22 chars)
    operation VARCHAR(50) NOT NULL,  -- Operation type (e.g., "create", "update", "delete")
    payload JSONB NOT NULL,  -- Event payload as JSON
    entity_uuid VARCHAR(22) NOT NULL,  -- UUID of the entity this event relates to
    actor_uuid VARCHAR(22),  -- UUID of the actor who triggered this event
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP  -- Timestamp when event was created
);

-- Create standard index on entity_uuid
CREATE INDEX idx_example_entity_events_entity_uuid ON example_entity_events(entity_uuid);

-- Create index on created_at for chronological queries
CREATE INDEX idx_example_entity_events_created_at ON example_entity_events(created_at);

-- Create GIN indexes for efficient JSONB searching
CREATE INDEX gin_example_entity_events_payload ON example_entity_events USING GIN (payload);

-- Create specific indexes for common fields
CREATE INDEX gin_example_entity_events_email ON example_entity_events USING GIN ((payload -> 'email'));
CREATE INDEX gin_example_entity_events_name ON example_entity_events USING GIN ((payload -> 'name'));

-- Add comments to document the table structure
COMMENT ON TABLE example_entity_events IS 'Event store for ExampleEntity following event sourcing pattern';
COMMENT ON COLUMN example_entity_events.uuid IS 'Unique identifier for the event (Base62 UUID v7)';
COMMENT ON COLUMN example_entity_events.operation IS 'Type of operation (CREATE, UPDATE, DELETE)';
COMMENT ON COLUMN example_entity_events.payload IS 'Event data as JSONB for flexible schema';
COMMENT ON COLUMN example_entity_events.entity_uuid IS 'UUID of the entity this event relates to';
COMMENT ON COLUMN example_entity_events.actor_uuid IS 'UUID of the user/system that triggered this event';
COMMENT ON COLUMN example_entity_events.created_at IS 'Timestamp when the event was created';