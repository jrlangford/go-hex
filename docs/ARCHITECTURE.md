# Architecture Documentation

## Overview

This project follows the Hexagonal Architecture (Ports and Adapters) pattern combined with Domain-Driven Design (DDD) and Event-Driven Architecture (EDA) principles.

## Core Principles

### Hexagonal Architecture
- **Driving Adapters** (Primary/Left side): HTTP handlers, CLI commands, test harnesses
- **Core Business Logic**: Domain models, application services, ports
- **Driven Adapters** (Secondary/Right side): Databases, message queues, external APIs

### Domain-Driven Design
- **Entities**: Objects with identity that change over time (e.g., `Friend`)
- **Value Objects**: Immutable objects that describe things (e.g., `FriendData`)
- **Aggregates**: Clusters of domain objects that are treated as a single unit
- **Domain Events**: Important business events that have occurred

### Event-Driven Architecture
- Domain events are published when significant business events occur
- Events are named in past tense (e.g., `FriendCreated`, `FriendDataUpdated`, `FriendDeleted`)
- Events help decouple components and enable async processing

## Project Structure

```
├── cmd/                    # Application entry point
│   └── main.go            # Main application bootstrap with dependency wiring
├── internal/
│   ├── core/              # Business logic (no external dependencies)
│   │   ├── domain/        # Domain models organized by bounded context
│   │   │   ├── authorization/ # Authorization domain (roles, permissions, context)
│   │   │   ├── {sub-context-1}/   # Sub-context 1 domain logic (entity, events, value objects)
│   │   │   ├── {sub-context-2}/   # Sub-context 2 domain logic (entity, events, value objects)
│   │   │   ├── ...        # Sub-context N domain logic (entity, events, value objects)
│   │   │   └── common/    # Shared domain infrastructure (base models, errors)
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
```

## Dependency Flow

```
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

### Public Endpoints
- `GET /health` - Health check endpoint

### Protected Endpoints (require authentication)
- `POST /friends` - Create a new friend (admin role required)
- `GET /friends` - List all friends (any authenticated user)
- `GET /friends/{id}` - Get specific friend details (any authenticated user)
- `PUT /friends/{id}` - Update friend information (admin role or owner)
- `DELETE /friends/{id}` - Delete a friend (admin role required)
- `GET /greet` - Greet endpoint with optional authentication
- `GET /auth/me` - Get current user information

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

## Domain Model

The application implements a simple friend management system with the following key domain concepts:

### Entities and Value Objects
- **Friend**: The main entity representing a person with an identity (`FriendID`) and data (`FriendData`)
- **FriendData**: A value object containing name and optional title
- **FriendID**: A value object wrapping a UUID for type safety
- **AuthorizationContext**: Contains user identity, role, and permissions for access control

### Domain Events
- **FriendCreatedEvent**: Published when a new friend is added
- **FriendDataUpdatedEvent**: Published when friend information is modified
- **FriendDeletedEvent**: Published when a friend is removed

### Authorization Model
- **Role-based permissions**: Admin, User, and ReadOnly roles with specific capabilities
- **Resource ownership**: Users can modify their own resources
- **Permission-based access control**: Fine-grained permissions (AddFriend, ViewFriend, etc.)

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

## Extension Points

To add new features:

1. **New Entity**: Add to `internal/core/domain/`
2. **New Use Case**: Add to `internal/core/application/`
3. **New Port**: Add interface to `internal/core/ports/`
4. **New Adapter**: Add implementation to `internal/adapters/`
5. **New Event**: Add to `internal/core/domain/` (e.g., `friend_events.go`)
