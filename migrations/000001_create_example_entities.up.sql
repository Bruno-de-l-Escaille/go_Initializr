CREATE TABLE IF NOT EXISTS example_entities (
    uuid VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    email VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('active', 'inactive', 'pending')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_example_entities_email ON example_entities(email);
CREATE INDEX IF NOT EXISTS idx_example_entities_status ON example_entities(status);
CREATE INDEX IF NOT EXISTS idx_example_entities_created_at ON example_entities(created_at);