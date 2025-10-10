# Go Initializr

A Go REST API boilerplate project based on the gopeople architecture, featuring a complete CRUD implementation for an ExampleEntity.

## Features

- **Clean Architecture**: Follows a layered architecture with models, repositories, services, and handlers
- **JWT Authentication**: ES256 (ECDSA) asymmetric JWT validation using public key from gopeople auth service
- **REST API**: Complete CRUD operations with proper HTTP status codes
- **Database Integration**: PostgreSQL support with prepared statements and migrations
- **UUID v7**: Time-sortable UUIDs for better database performance
- **Live Reload**: Air integration for development with automatic reloading
- **Database Migrations**: Structured database schema management
- **Dependency Injection**: Centralized dependency management following gopeople patterns
- **Logging**: Structured logging with zerolog including user context
- **Middleware**: CORS, request logging, recovery, JWT validation, and request ID tracking
- **Configuration**: Environment-based configuration
- **Swagger Documentation**: API documentation with Swagger/OpenAPI and JWT security

## Project Structure

```
go_Initializr/
├── .air.toml               # Air configuration for live reloading
├── docs/                   # Swagger documentation
├── handler/                # HTTP handlers (controllers)
│   ├── middleware/         # Custom middleware
│   └── example_entity_handler.go
├── internal/               # Internal packages
│   ├── injection/          # Dependency injection
│   └── logging/            # Logging configuration
├── migrations/             # Database migrations
│   ├── 000001_create_example_entities.up.sql
│   ├── 000001_create_example_entities.down.sql
│   └── migrate.go         # Migration utilities
├── models/                 # Data models and validation
├── pkg/                    # Public packages
│   ├── db/                # Database utilities (connection, UUID v7, transactions)
│   └── initializer/       # Application initialization
├── repository/             # Data access layer
├── routes/                # Route definitions
├── service/               # Business logic layer
├── tmp/                   # Air temporary files (auto-generated)
├── .gitignore            # Git ignore rules
├── example.env           # Environment variables template
├── go.mod               # Go module definition
├── Makefile            # Build and development commands
├── main.go             # Application entry point
└── README.md          # This file
```

## Getting Started

### Prerequisites

- Go 1.23.0 or later
- PostgreSQL 12 or later

### Installation

1. **Clone or copy the boilerplate code**
2. **Navigate to the project directory**
   ```bash
   cd go_Initializr
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Set up environment variables**
   ```bash
   cp example.env .env
   ```
   Edit `.env` with your database configuration.

5. **Obtain JWT Public Key**
   
   This service validates JWTs signed by the gopeople auth service. You need the public key:
   
   ```bash
   # Copy public key from gopeople service
   cp /path/to/gopeople/keys/public.pem keys/public.pem
   ```
   
   Set the path in `.env`:
   ```env
   JWT_PUBLIC_KEY_PATH=./keys/public.pem
   ```
   
   **Note**: This service only validates JWTs. It does NOT generate or sign tokens. See `README_JWT.md` and `JWT_ARCHITECTURE.md` for details.

6. **Create the database**
   Create a PostgreSQL database named `go_Initializr` (or whatever you set in DB_NAME).

7. **Run database migrations**
   ```bash
   go run migrations/migrate.go
   ```

### Running the Application

**Development (with live reload using Air):**
```bash
air
```

**Production:**
```bash
go run main.go
```

**Build binary:**
```bash
go build -o go_initializr
./go_initializr
```

The server will start on `http://localhost:8080` by default.

### API Endpoints

#### ExampleEntity Endpoints

- `GET /api/example-entities` - Get all entities with optional filtering and pagination
- `POST /api/example-entities` - Create a new entity
- `GET /api/example-entities/{uuid}` - Get entity by UUID
- `PUT /api/example-entities/{uuid}` - Update entity
- `DELETE /api/example-entities/{uuid}` - Delete entity

#### Other Endpoints

- `GET /health` - Health check
- `GET /swagger/index.html` - API documentation (after generating docs)

### JWT Authentication

The API uses **ES256 (ECDSA with P-256)** for JWT validation. This is an asymmetric cryptography approach where:

- **gopeople service** signs JWTs using a private key
- **go_Initializr service** validates JWTs using a public key (this service)
- Tokens are included in the `Authorization` header as a Bearer token

**Algorithm**: ES256 (ECDSA with SHA-256 on P-256 curve)  
**Validation**: Public key cryptography (no shared secrets)

**Protected Endpoints** (require JWT):
- `POST /api/example-entities` - Create entity
- `PUT /api/example-entities/{uuid}` - Update entity  
- `DELETE /api/example-entities/{uuid}` - Delete entity

**Public Endpoints** (no authentication required):
- `GET /api/public/example-entities` - Get all entities
- `GET /api/public/example-entities/{uuid}` - Get entity by UUID
- `GET /health` - Health check

**JWT Header (ES256):**
```json
{
  "alg": "ES256",
  "typ": "JWT"
}
```

**JWT Claims:**
```json
{
  "user_uuid": "user-uuid-here",
  "user_id": "user-id",
  "email": "user@example.com",
  "exp": 1640995200,
  "iat": 1640908800,
  "iss": "gopeople"
}
```

**How to get a JWT:**
1. Authenticate with the gopeople service (POST `/api/v1/auth/login`)
2. Use the returned JWT token in requests to go_Initializr

For more details on JWT validation architecture, see:
- `README_JWT.md` - JWT implementation details
- `JWT_ARCHITECTURE.md` - Service architecture and roles

### Example Usage

**Get all entities (public):**
```bash
curl http://localhost:8080/api/public/example-entities
```

**Get entity by UUID (public):**
```bash
curl http://localhost:8080/api/public/example-entities/{uuid}
```

**Create an entity (requires JWT from gopeople):**
```bash
# First, get JWT from gopeople auth service
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.access_token')

# Then use the JWT in go_Initializr
curl -X POST http://localhost:8080/api/example-entities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "John Doe",
    "description": "A sample entity",
    "email": "john@example.com",
    "status": "active"
  }'
```

**Update entity (requires JWT):**
```bash
curl -X PUT http://localhost:8080/api/example-entities/{uuid} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Updated Name"
  }'
```

**Delete entity (requires JWT):**
```bash
curl -X DELETE http://localhost:8080/api/example-entities/{uuid} \
  -H "Authorization: Bearer $TOKEN"
```

### Database Migrations

**Apply migrations:**
```bash
go run migrations/migrate.go
```

**Create a new migration manually:**
Create two files in `migrations/`:
- `000002_migration_name.up.sql` - Forward migration
- `000002_migration_name.down.sql` - Rollback migration

### UUID v7

The project uses UUID v7 for entity IDs, which provides:
- **Time-based sorting**: UUIDs are chronologically sortable
- **Better database performance**: Reduces index fragmentation
- **Extractable timestamps**: Can extract creation time from UUID

```go
// Generate UUID v7
id := db.MustUUIDv7()

// Extract timestamp from UUID v7
timestamp, err := db.ExtractTimestampFromUUIDv7(id)
```

### Generating Swagger Documentation

Generate documentation:
```bash
swag init
```

Access documentation at: `http://localhost:8080/swagger/index.html`

**Install swag if not already installed:**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Architecture Overview

### Models Layer (`models/`)
- Defines data structures and validation rules
- Contains request/response models for API endpoints
- Includes utility functions for password hashing and validation

### Repository Layer (`repository/`)
- Handles data persistence and retrieval
- Implements interfaces for testability
- Contains database-specific logic and queries

### Service Layer (`service/`)
- Contains business logic and validation
- Orchestrates operations between repositories
- Handles complex business rules and workflows

### Handler Layer (`handler/`)
- HTTP request/response handling
- Input validation and error handling
- Converts between HTTP and internal data structures

### Routing (`routes/`)
- Defines API endpoints and middleware
- Groups related routes together
- Applies middleware per route group

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `go_Initializr` |
| `DB_SSLMODE` | SSL mode for database | `disable` |
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `JWT_PUBLIC_KEY_PATH` | Path to ECDSA public key for JWT validation | `./keys/public.pem` |

## Development

### Adding New Entities

To add a new entity (e.g., `Product`):

1. **Define the model** in `models/models.go`
2. **Create repository interface and implementation** in `repository/`
3. **Implement service layer** in `service/`
4. **Create HTTP handlers** in `handler/`
5. **Add routes** in `routes/routes.go`
6. **Wire dependencies** in `internal/injection/injection.go`

### Testing

The project structure supports easy testing:
- Repository interfaces allow for mock implementations
- Service layer can be tested independently
- Handlers can be tested with HTTP test frameworks

### Logging

The application uses structured logging with zerolog:
- Logs are output in JSON format in production
- Console-friendly format in development
- Request logging middleware tracks all HTTP requests

## Contributing

1. Follow the existing code structure and patterns
2. Add proper error handling and logging
3. Write tests for new functionality
4. Update documentation as needed
