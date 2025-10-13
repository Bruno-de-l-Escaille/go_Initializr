# Go Initializr

This repository contains two branches for bootstrapping a Go API project:

## Branches

### 1. `Go_starter_api`

This branch provides a clean architecture for a REST API in Go.

**Features:**

*   **Clean Architecture:** Follows the principles of clean architecture to create a decoupled and maintainable codebase.
*   **Native SQL:** Uses native SQL for database interactions instead of an ORM, providing more control and performance.
*   **UUID v7:** Implements UUID v7 for primary keys.
*   **GoPeople-based Authentication:** Integrates JWT validation using a public key, based on GoPeople's authentication system.
*   **Swagger Documentation:** Includes Swagger for API documentation.
*   **Structured Logging:** Uses `zerolog` for structured, high-performance logging.

### 2. `event-sourcing-go-starter-api`

This branch provides a starter project based on an event-sourcing and CQRS architecture, similar to the one used at GoPeople.

**Features:**

*   **Event Sourcing & CQRS:** Implements the Event Sourcing and Command Query Responsibility Segregation (CQRS) patterns.
*   **Native SQL:** Uses native SQL for database interactions.
*   **UUID v7:** Implements UUID v7.
*   **GoPeople-based Authentication:** Integrates JWT validation using a public key.
*   **Swagger Documentation:** Includes Swagger for API documentation.
*   **Structured Logging:** Uses `zerolog` for structured logging.
