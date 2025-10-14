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

## How to Use

This project is designed to be used as a template. Follow these steps to start your own project based on one of the branches.

### 1. Clone the Repository

Clone this repository into a new directory with your desired project name.

```bash
git clone https://github.com/Bruno-de-l-Escaille/go_Initializr.git <your-project-name>
cd <your-project-name>
```

### 2. Select Your Preferred Architecture

This repository offers two branches, each with a different architectural pattern. Choose the one that best fits your needs.

*   **For a standard REST API with a clean architecture**, use the `Go_starter_api` branch.
*   **For an event-sourcing and CQRS architecture**, use the `event-sourcing-go-starter-api` branch.

To start fresh with your chosen branch, you will create a new local branch from it and then remove the original remote.

**Example: Using `Go_starter_api`**

```bash
# Fetch all branches from the remote
git fetch origin

# Create a new branch from the desired remote branch
git checkout -b main origin/Go_starter_api

# At this point, you have a local `main` branch based on `Go_starter_api`.
# You can now remove the original remote to start your own version history.
git remote remove origin

# Verify that the remote has been removed
git remote -v
```

### 3. Update Project Name and Go Module Path

After cloning, you need to replace all occurrences of the default project name `go_Initializr` with your chosen `<your-project-name>`. This includes updating the Go module path in the `go.mod` file and any import paths within `.go` files.

You can automate this with the following commands.

**1. Replace in all files:**

Replace `<your-module-path>` with your full module path (e.g., `github.com/your-user/your-repo`).

*   **For macOS & Linux:**

    For example, to replace `go_Initializr` with `gotask`:
    ```bash
    find . -type f -name '*.go' -exec sed -i '' 's|go_Initializr|gotask|g' {} +
    sed -i '' 's|go_Initializr|gotask|' go.mod
    ```

**2. Tidy your Go module:**

Finally, tidy up your `go.mod` file to ensure all dependencies are correct.

```bash
go mod tidy
```

After these steps, you will have a clean project in your local directory, ready for you to initialize a new Git repository and push it to your own remote.
