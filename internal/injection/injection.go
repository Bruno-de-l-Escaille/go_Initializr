package injection

// This package follows the gopeople injection pattern with:
// - Core components for shared dependencies
// - Business component modules for specific domains
// - Container to orchestrate all components
//
// Structure:
// - config.go: Application configuration
// - core.go: Core shared components (DB, base services)
// - example_entity.go: ExampleEntity domain components
// - container.go: Main dependency container
