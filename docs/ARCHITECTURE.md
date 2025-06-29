# Architecture Documentation

## Overview

This project follows the Hexagonal Architecture (Ports and Adapters) pattern combined with Domain-Driven Design (DDD) and Event-Driven Architecture (EDA) principles.

## Core Principles

### Hexagonal Architecture

- **Driving Adapters** (Primary/Left side): HTTP handlers, CLI commands, test harnesses
- **Core Business Logic**: Domain models, application services, ports
- **Driven Adapters** (Secondary/Right side): Databases, message queues, external APIs

Guidelines:
- Prioritize separation of concerns, dependency inversion, and testability in all code generation.
- Ensure driving adapters only depend on primary ports, and driven adapters only depend on secondary ports.

### Domain-Driven Design

- **Entities**: Objects with identity that change over time (e.g., `Friend`)
- **Value Objects**: Immutable objects that describe things (e.g., `FriendData`)
- **Aggregates**: Clusters of domain objects that are treated as a single unit
- **Domain Events**: Important business events that have occurred

Guidelines:

- Ensure that domain models are rich and encapsulate business logic.
- Use value objects and aggregates where appropriate.
- Avoid leaking domain logic into adapters or ports.
- Embed the BaseEntity defined in `./internal/core/domain/shared/base_model.go` in all domain entities to ensure they have a consistent structure and behavior.
- BaseEntities should be initialized with the entitiy's domain id, so that the id within the instance of the BaseEntity has the same type as the domain id type. The domain id type is meant to ensure type safety and prevent accidental misuse of IDs across different entities. It should be named after the entity (e.g., `FriendID` for `Friend` entity).
- Type-safe IDs will embed the base uuid.UUID structure so its methods become available to callers of the new type.
- Other public attributes within the entity should be grouped into a Value Object that is embedded into the struct, this will enable in the use of this value objects in Port definitions.
- All models shall implement validations defined as annotations. Validations are run in model constructors using the validation functions defined in `/support/validation/validator.go`. This has the effect of always reusing the same validation object.

### Event-Driven Architecture

Domain events are published when significant business events occur. They decouple components and enable async processing

Guidelines:

- Encourage the use of events to decouple components.
- Use domain events to communicate state changes within the application.
- Express domain events in past tense (e.g., `UserRegistered`, `OrderPlaced`).

## Project Structure

```txt
├── cmd/                    # Application entry point
│   └── main.go            # Main application bootstrap with dependency wiring
├── internal/
│   ├── core/              # Business logic (no external dependencies)
│   │   ├── domain/        # Domain models organized by bounded context
│   │   │   ├── authorization/ # Authorization domain (roles, permissions, context)
│   │   │   ├── {sub-context-1}/   # Sub-context 1 domain logic (entity, events, value objects)
│   │   │   ├── {sub-context-2}/   # Sub-context 2 domain logic (entity, events, value objects)
│   │   │   ├── ...        # Sub-context N domain logic (entity, events, value objects)
│   │   │   └── shared/    # Shared domain infrastructure (base entities, errors)
│   │   ├── application/   # Use cases, application services
│   │   └── ports/         # Interface definitions
│   │       ├── primary/   # Driving ports (inbound)
│   │       └── secondary/ # Driven ports (outbound)
│   ├── adapters/          # External concerns
│   │   ├── driving/       # Inbound adapters (HTTP, CLI, etc.)
│   │   └── driven/        # Outbound adapters (DB, messaging, etc.)
│   └── support/           # Infrastructure and cross-cutting concerns
│       ├── config/        # Application configuration management
│       ├── logging/       # Logging setup and configuration
│       ├── server/        # HTTP server lifecycle management
│       ├── auth/          # Authentication and authorization utilities
│       ├── errors/        # Base error types and error handling
│       └── validation/    # Data validation utilities
├── docs/                  # Repository documentation
└── build/                 # Build artifacts
```

## Dependency Flow

```txt
Driving Adapters → Primary Ports → Application Services → Secondary Ports → Driven Adapters
     (HTTP)           (Greeter)       (HelloService)    (FriendRepository)    (InMemory)
```

## Key Design Decisions

1. **Dependency Inversion**: Core business logic doesn't depend on external concerns
2. **Interface Segregation**: Small, focused interfaces for each port (`Greeter`, `HealthChecker`, `FriendRepository`, `EventPublisher`)
3. **Testability**: Easy to mock dependencies and test in isolation (comprehensive test suites with mocks)
4. **Event Publishing**: Domain events for loose coupling and auditability (`FriendCreated`, `FriendDataUpdated`, `FriendDeleted`)
5. **Rich Domain Models**: Business logic encapsulated in domain objects (Friend entity with behavior methods)
6. **Graceful Shutdown**: Proper server lifecycle management with signal handling
7. **Cross-cutting Concerns**: Infrastructure concerns (logging, configuration, auth, validation) are treated separately from business logic
8. **Configuration Driven**: Behavior is configurable through environment variables
9. **Type Safety**: Strong typing with domain-specific IDs and value objects
10. **Authorization**: Role-based access control with permission checks at the application service level

## Application Bootstrap

The `cmd/main.go` orchestrates the application startup in a clean, sequential manner:

1. Load configuration from environment variables
2. Initialize structured logging
3. Wire dependencies directly (no separate container abstraction)
4. Start HTTP server with graceful shutdown handling

This approach keeps dependency wiring simple and co-located with the application entry point, following clean architecture principles while maintaining separation of concerns.

## API Design

The HTTP adapter exposes the following RESTful endpoints:

### Authentication & Authorization

- JWT-based authentication using configurable secret, issuer, and audience
- Role-based access control with middleware enforcement
- Support for optional authentication on certain endpoints

## Domain Organization

The domain layer is organized into distinct packages, each encapsulating a specific concern following DDD principles:

### Design Principles

- **Bounded Contexts**: Each package represents a clear bounded context
- **Separation of Concerns**: Domain logic is isolated from infrastructure
- **Rich Domain Models**: Entities encapsulate business logic and behavior
- **Event-Driven**: Domain events enable loose coupling and auditability
- **Type Safety**: Value objects and strong typing prevent invalid states
- **Direct Imports**: Components import domain packages directly (e.g., `friend`, `authorization`)

## Testing Strategy

- **Unit Tests**: Comprehensive test coverage for domain logic and application services using mocks
  - Domain models with behavior testing (`Friend`, `AuthorizationContext`)
  - Application services with authorization testing (`HelloService`, `HealthService`)
  - Mock implementations for repositories and event publishers
- **Integration Tests**: Test adapters with real implementations
  - HTTP handlers and middleware testing
  - Authentication and authorization flow testing
- **Contract Tests**: Ensure adapters properly implement ports
  - Repository implementations conform to `FriendRepository` interface
  - Event publishers conform to `EventPublisher` interface
