# Design Guidelines

This document provides general guidelines for implementing Domain-Driven Design (DDD) with Hexagonal Architecture in Go, based on the patterns and principles demonstrated in this codebase.

## Overview

This project follows DDD principles with a hexagonal (ports and adapters) structure, emphasizing:

- **Domain-centricity**: Business logic is isolated in the domain layer
- **Dependency inversion**: Core business logic has no dependencies on external concerns
- **Bounded contexts**: Clear boundaries between different business domains
- **Event-driven integration**: Loose coupling between contexts via domain events

## Core Principles

### 1. Dependency Direction

- **Inward dependencies only**: All dependencies point toward the domain core
- **Interface segregation**: Define minimal, focused interfaces
- **Dependency injection**: Use constructor injection for all dependencies

### 2. Layer Responsibilities

#### Domain Layer (`internal/core/{context}/domain/`)

- Contains pure business logic and rules
- Defines domain entities, value objects, and domain services
- Publishes domain events for cross-context communication
- Has zero dependencies on infrastructure or application layers

#### Application Layer (`internal/core/{context}/application/`)

- Orchestrates business workflows
- Handles use case execution
- Coordinates between its corresponding domain service and external systems
- Manages transactions and event publishing

#### Ports (`internal/core/{context}/ports/`)

- **Primary ports**: Interfaces for driving the application (e.g., use case interfaces)
- **Secondary ports**: Interfaces for driven dependencies (e.g., repositories, external services)

#### Adapters (`internal/adapters/`)

- **Driving adapters**: HTTP handlers, CLI commands, message consumers
- **Driven adapters**: Database implementations, external API clients, message publishers

## Directory Structure

```txt
├── cmd/                   # Application entry point
│   └── main.go            # Main application bootstrap with dependency wiring
├── internal/
│   ├── core/                           # Business logic
│   │   ├── {context}/                  # Bounded context
│   │   │   ├── domain/                 # Domain entities, value objects, events
│   │   │   ├── application/            # Use cases, application services, authorization
│   │   │   └── ports/                  # Interfaces
│   │   │       ├── primary/            # Inbound interfaces (use cases)
│   │   │       └── secondary/          # Outbound interfaces (repositories, etc.)
│   ├── adapters/                       # Infrastructure implementations
│   │   ├── driving/                    # Inbound adapters (HTTP, CLI, etc.)
│   │   ├── driven/                     # Outbound adapters (DB, external APIs)
│   │   └── integration/                # Cross-context integration adapters (ACLs)
│   └── support/                        # Technical infrastructure and reusable components
│       ├── auth/                       # Authorization utilities
│       ├── basedomain/                 # Base domain types and utilities
│       ├── config/                     # Configuration management
│       ├── errors/                     # Base error types and definitions
│       ├── logging/                    # Logging utilities
│       ├── validation/                 # Input validation
│       └── server/                     # HTTP server setup and middleware
├── pkg/                   # Public packages for reuse
├── docs/                  # Repository documentation
└── test/                  # Integration tests
    └── integration/       # Cross-context integration tests
```

## Bounded Context Guidelines

### Context Identification

- Each context should represent a distinct business domain
- Contexts should have minimal coupling
- Use domain events for cross-context communication
- Each context should be independently deployable (if needed)

### Context Boundaries

- **Clear ownership**: Each entity belongs to exactly one context
- **Anti-corruption layers**: Use adapters, placed in `adapters/integration/`, to translate between contexts
- **Shared kernel**: Minimize shared concepts, isolate in `shared/` directory

## Asynchronous Cross-Domain Integration (Event-Driven)

### Domain Events

- Events represent business-significant occurrences
- Events should be named in past tense (e.g., `CargoBooked`, `HandlingEventRegistered`)
- Include all necessary data to avoid chatty integration
- Use ACL (Anti-Corruption Layer) for consuming events from other contexts

### Internal Event Handlers

- Place internal event handlers in `adapters/integration/`
- Implement handlers for domain events to trigger business logic

### Event Bus Pattern

- Use an event bus to decouple event producers from consumers
- Ensure the bus implementation provides reliable delivery and supports retries

```go
type EventBus interface {
    Publish(ctx context.Context, event DomainEvent) error
    Subscribe(eventType string, handler EventHandler)
}

type EventHandler interface {
    Handle(ctx context.Context, event DomainEvent) error
}
```

## Synchronous Cross-Domain Integration

- Use ACL adapters for synchronous calls between contexts

### Integration Handlers

- Place integration event handlers in `adapters/integration/`
- Implement Anti-Corruption Layer (ACL) pattern
- Handle failures gracefully with retries and dead letter queues
- Always use ACL when consuming events or data from other bounded contexts.
- ACL adapters should be stateless and focus purely on translation and validation.

## Implementation Patterns

### Entities and Value Objects

- Embed the BaseEntity defined in `./internal/support/basedomain/base_entity.go` in all domain entities to ensure they have a consistent structure and behavior.
- Initialize BaseEntities with the entity's domain id, so that the id within the instance of the BaseEntity has the same type as the domain id type. The domain id type ensures type safety and prevents accidental misuse of IDs across different entities. Name it after the entity (e.g., `TrackingID` for `Cargo` entity).
- Type-safe IDs must embed the base uuid.UUID structure so its methods become available to callers of the new type.
- Group other public attributes within the entity into a Value Object that is embedded into the struct, this will enable their use in Port definitions.
- Add validation annotations to models. Validations are run in model constructors using the validation functions defined in `/support/validation/validator.go`, which call a single validation object. This reuses the cache of the validator to avoid performance issues with repeated validation calls.
- Add other annotations to models as needed, such as JSON tags for serialization, or custom tags for specific behaviors. This is particularly useful when other layers wrap domain objects in DTOs or when using reflection-based libraries.

```./internal/support/basedomain/base_entity.go

// ..

type BaseEntity[T EntityID] struct {
  Id        T         `json:"id" validate:"required"`
  CreatedAt time.Time `json:"created_at" validate:"required"`
  UpdatedAt time.Time `json:"updated_at" validate:"required,gtfield=CreatedAt"`

  events []DomainEvent `json:"-"`
}

// ..

```

```go

// ..
type SomeEntityID struct {
  uuid.UUID

// ..

// SomeEntity - has identity and lifecycle
type SomeEntity struct {
  shared.BaseEntity[SomeEntityID] `json:",inline"`
  Data SomeData `json:"data"`
}

// ..

// SomeData - immutable, compared by value
type SomeData struct {
  SomeString    string `json:"some_string" validate:"required"`
  SomeInt       int    `json:"some_int" validate:"gte=18,lte=120"`
  // ... validation in constructor
}


```

### Repository Pattern

```go
// Define interface in domain ports
type Repository interface {
    Save(ctx context.Context, entity Entity) error
    FindByID(ctx context.Context, id EntityID) (*Entity, error)
}

// Implement in driven adapters
type InMemoryRepository struct {
    data map[string]*Entity
}
```

### Service Layer

```go
// Application service coordinates use cases
type ApplicationService struct {
    repository Repository
    eventBus   EventBus
}

func (s *ApplicationService) ExecuteUseCase(ctx context.Context, cmd Command) error {
    // 1. Load domain objects
    // 2. Execute business logic
    // 3. Save changes
    // 4. Publish events
}
```

### Outbox pattern

- Use the outbox pattern to ensure reliable event publishing
- Store events in a dedicated outbox table within the same transaction as the domain changes

## HTTP API Guidelines

### Request/Response DTOs

- Define separate DTOs for API contracts. If domain objects provide a good fit, wrap them in DTOs to avoid exposing domain logic directly.
- Use validation tags for input validation
- Convert between DTOs and domain objects in handlers

### Authentication and Authorization

- Use middleware for cross-cutting concerns
- Implement role-based access control
- Use JWT tokens with proper validation

### Endpoints

- Use RESTful principles for endpoint Design
- Use nouns for resources (e.g., `/cargo`, `/shipment`)
- Use HTTP methods to represent actions (GET for retrieval, POST for creation, PUT/PATCH for updates, DELETE for deletion)
- Use plural nouns for collections (e.g., `/cargos`, `/shipments`)

### Error Handling

- Return consistent error response format
- Map domain errors to appropriate HTTP status codes
- Include error codes for client-side handling

## Testing Strategy

### Unit Tests

- Test domain logic in isolation
- Test domain logic with authentication and authorization mocked
- Mock external dependencies
- Focus on business rule validation

### Integration Tests

- Test complete workflows end-to-end
- Implement integration tests so they can run against both mocked and real dependencies:
  - Mocked integration tests: Use in-memory databases or mocks for external services
  - Real integration tests: Use real databases and services to verify end-to-end behavior
- Verify event-driven integration

### Test data generation

- All test data should be generated via application services, not hardcoded, to ensure consistency and avoid brittle tests
- Use factories or builders to create test data Objects
- When generating random data, ensure randomness is controlled (e.g., using a fixed seed) to make tests deterministic and repeatable

### API Tests

- Test HTTP endpoints with real authentication
- Verify request/response formats
- Test error scenarios

### Test Coverage

- Aim for high coverage, prioritizing critical paths

### Early Integration with Other Systems

- Implement a skeleton for adapters so that other systems can start integrating early
- Generate dummy data for integration tests to allow other teams to test their systems against the application
- Mocked application services should use actual domain objects to ensure that the integration points are well-defined
- Allow the application to run in mock mode, where it can be started without any external dependencies
- Update mocked application services periodically to reflect changes in the domain model, ensuring that integration points remain valid

## Configuration Management

### Environment-based Configuration

- Use environment variables for runtime configuration
- Provide sensible defaults for development
- Validate configuration at startup

### Dependency Injection

- Use constructor injection
- Create factory functions for complex object graphs
- Separate configuration from business logic

## Security Considerations

### Authentication

- Use industry-standard JWT tokens
- Implement proper token validation
- Support token expiration and refresh
- Inject authenticated user context into application services

### Authorization

- Implement domain-driven authorization where each bounded context defines its own permissions
- Each bounded context should have a `permissions.go` file in its `application` package with:
  - Context-specific permission types (e.g., `BookingPermission`, `RoutingPermission`)
  - Require*Permission functions that accept claims and permission
- Use `/internal/support/auth` for shared authentication logic (claims extraction, verification)
- Application services extract claims from context using `auth.ExtractClaims(ctx)` and then check domain permissions
- Enforce authorization at the application service layer before executing domain operations
- Use middleware for consistent token validation and context injection
- Audit access attempts using structured logging

### Input Validation

- Validate all inputs at the boundary
- Use structured validation with clear error messages
- Sanitize inputs to prevent injection attacks

## Performance Guidelines

### Database Access

- Use repository pattern for data access abstraction
- Implement proper connection pooling
- Consider read/write separation for scalability

### Caching

- Cache at appropriate layers (application, not domain)
- Use cache-aside pattern
- Implement cache invalidation strategy

### Event Processing

- Use asynchronous event processing when possible
- Implement proper error handling and retries
- Monitor event processing delays

## Monitoring and Observability

### Logging

- Use structured logging (JSON format)
- Log at appropriate levels (DEBUG, INFO, WARN, ERROR)
- Include correlation IDs for request tracing

### Metrics

- Track business metrics and technical metrics
- Monitor endpoint performance
- Track event processing metrics

### Health Checks

- Implement health check endpoints
- Check dependencies (database, external services)
- Provide detailed health information

## Migration and Deployment

### Database Migrations

- Version control database schema changes
- Use forward-compatible migrations
- Test migrations in staging environment

### Backward Compatibility

- Maintain API compatibility during upgrades
- Use versioning for breaking changes
- Implement graceful degradation

### Zero-Downtime Deployment

- Design for rolling deployments
- Use feature flags for gradual rollouts
- Implement proper health checks

### Signal Handling

- Ensure graceful shutdown on termination signals
- Use context cancellation to stop processing gracefully

## Common Anti-Patterns to Avoid

### Domain Layer Violations

- Don't put infrastructure dependencies in domain
- Don't let domain objects know about HTTP or databases
- Don't use anemic domain models (just getters/setters)

### Architectural Violations

- Don't skip the ports layer
- Don't let adapters contain business logic
- Don't create circular dependencies between contexts

### Event-Driven Anti-Patterns

- Don't use events for synchronous communication
- Don't include too much data in events
- Don't create event chains that are hard to debug

## Conclusion

This design provides a solid foundation for building maintainable, testable, and scalable applications. The key is to maintain clear boundaries, respect dependency directions, and keep business logic isolated from technical concerns.

Remember: these guidelines should be adapted to your specific domain and requirements. The goal is consistency and maintainability, not rigid adherence to patterns.
