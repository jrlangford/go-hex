# Cargo Shipping System - Sample Implementation

This document describes the reference implementation of the Go Hex template, which demonstrates a cargo shipping system following Domain-Driven Design (DDD) principles, Hexagonal Architecture, and Event-Driven Architecture.

This implementation serves as a practical example of the design principles and architectural patterns described in the [Design Guidelines](DESIGN_GUIDELINES.md).

## Overview

The cargo shipping system is a multi-context domain that demonstrates real-world application of DDD principles. It consists of three bounded contexts that work together to manage the complete cargo shipping lifecycle.

## Domain Contexts

The implementation includes three bounded contexts:

- **Booking Context**: Book cargo for shipment with origin, destination, and delivery requirements
- **Routing Context**: Find optimal routes for cargo based on voyages and schedules
- **Handling Context**: Track cargo handling events throughout the shipping process

## Getting Started with the Sample

### Prerequisites

- devbox >= 0.14.2

### Quick Start

Start the application in mock mode with pre-populated test data:

```bash
devbox shell
just run mock
```

The server will start on `http://localhost:8080` in mock mode with realistic generated test data for demonstration purposes.

### Application Modes

The application supports two operational modes:

- **mock**: Pre-populated with realistic generated test data (ideal for demos and development)
- **live**: Clean repositories for real usage

```bash
# Mock mode (default for development)
just run mock

# Live mode
just run live
```

### Authentication

Generate a JWT token for API access using predefined roles:

```bash
# Full access (all operations)
go run tools/generate_test_token.go admin

# Standard operations (booking, viewing, tracking)
go run tools/generate_test_token.go user

# Read-only access (viewing only)
go run tools/generate_test_token.go readonly
```

Or create custom tokens:

```bash
go run tools/generate_test_token.go user-123 test.user admin,user test@example.com 24
```

Copy the generated token and use it in API requests.

## Sample API Usage

### 1. Check System Health

```bash
curl http://localhost:8080/health
```

### 2. List Available Voyages (Mock Mode)

```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8080/api/v1/voyages
```

### 3. List Available Locations

```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8080/api/v1/locations
```

### 4. List Pre-populated Cargo (Mock Mode)

```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8080/api/v1/cargos
```

### 5. Book New Cargo

```bash
curl -X POST http://localhost:8080/api/v1/cargos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "origin": "SESTO",
    "destination": "NLRTM",
    "arrivalDeadline": "2025-12-31T23:59:59Z"
  }'
```

## REST API Endpoints

### Public Endpoints

- `GET /health` - System health check
- `GET /info` - System information

### Cargo Management (Booking Context)

- `POST /api/v1/cargos` - Book new cargo
- `GET /api/v1/cargos` - List all cargo
- `GET /api/v1/cargos/{trackingId}` - Get specific cargo details
- `PUT /api/v1/cargos/{trackingId}/route` - Assign route to cargo

### Route Planning (Routing Context)

- `POST /api/v1/route-candidates` - Find route candidates
- `GET /api/v1/voyages` - List available voyages
- `GET /api/v1/locations` - List shipping locations

### Cargo Tracking (Handling Context)

- `POST /api/v1/handling-events` - Submit handling event
- `GET /api/v1/handling-events` - List handling events
- `GET /api/v1/handling-events?tracking_id={id}` - Get events for specific cargo

*Note: All API endpoints except `/health` and `/info` require JWT authentication.*

## Configuration

The application uses environment variables for configuration:

- `APP_MODE`: Application mode (mock, live) - default: live
- `PORT`: HTTP server port (default: 8080)
- `JWT_SECRET`: JWT signing secret (required)
- `JWT_ISSUER`: JWT issuer (default: "go-hex-service")
- `JWT_AUDIENCE`: JWT audience (default: "go-hex-api")
- `LOG_LEVEL`: Logging level (debug, info, warn, error) - default: info

## Inter-Context Integration Patterns

The system demonstrates two key DDD integration patterns:

### 1. Synchronous Integration (Customer-Supplier)

**Booking ↔ Routing**: Booking context acts as customer, Routing context as supplier

- When booking cargo, the system automatically finds available routes
- Implemented via direct service calls with Anti-Corruption Layer
- Ensures Booking context gets routes in its own domain language

### 2. Asynchronous Integration (Event-Driven)  

**Handling → Booking**: Loosely coupled via domain events

- Handling events (load, unload, customs) automatically update cargo delivery status
- Implemented via event bus with subscriber pattern
- Allows independent evolution of contexts

### Event-Driven Architecture

The system uses an in-memory event bus for inter-context communication:

- `HandlingEventRegistered` events update cargo delivery status
- Events are published in the Handling context's language
- Anti-Corruption Layer translates events for Booking context consumption

## Implementation Status

**Fully Implemented:**

- Core domain models for all three bounded contexts (Booking, Routing, Handling)
- Application services with complete business logic
- Hexagonal architecture with ports and adapters
- In-memory repository implementations for all aggregates
- Event-driven integration (Handling → Booking)
- Synchronous integration (Booking ↔ Routing)  
- Anti-Corruption Layers for context boundaries
- REST-compliant HTTP API endpoints
- JWT authentication and authorization (validated at application level)
- Multi-mode operation (mock/live)
- Realistic test data generation and integration
- Comprehensive integration test covering multi-context flows

**Future Enhancements:**

- Persistent storage adapters (database repositories)
- Comprehensive unit test coverage
- Docker containerization
- Monitoring and observability
- Rate limiting and advanced security features

## Testing the Sample

Run all tests:

```bash
just test
```

Run tests with coverage:

```bash
just test-coverage
```

### Integration Test

The system includes a comprehensive integration test that demonstrates the complete multi-context flow:

```bash
go test ./test/integration/ -v
```

This test verifies:

- Cargo booking in the Booking context
- Route finding via Booking→Routing integration  
- Route assignment to cargo
- Handling event registration in the Handling context
- Automatic delivery status updates via Handling→Booking events

## Architecture Highlights

The cargo shipping implementation showcases:

- **Domain-Driven Design**: Clear bounded contexts with rich domain models
- **Hexagonal Architecture**: Clean separation of business logic from infrastructure
- **Domain-Specific Authorization**: Each bounded context defines its own permissions
- **CQRS**: Separate command and query responsibilities
- **Event Sourcing**: Domain events for inter-context communication
- **Anti-Corruption Layers**: Protected integration boundaries
- **REST API**: Modern HTTP API following REST principles
- **JWT Security**: Token-based authentication with role-based access

## Domain Model Details

### Booking Context

- **Cargo**: Core aggregate representing shipments
- **TrackingId**: Unique identifier for cargo
- **RouteSpecification**: Origin, destination, and delivery requirements
- **Delivery**: Current status and progress tracking

### Routing Context

- **Voyage**: Shipping routes with schedules
- **Location**: Shipping ports and terminals (UN/LOCODE standard)
- **Itinerary**: Planned route for cargo

### Handling Context

- **HandlingEvent**: Physical cargo handling operations
- **HandlingHistory**: Complete timeline of cargo events
- **EventType**: Load, unload, customs, receive, claim

## Development Notes

This implementation demonstrates how to:

1. **Structure DDD Applications**: Clear separation of contexts, domains, and infrastructure
2. **Implement Hexagonal Architecture**: Ports and adapters pattern with dependency inversion
3. **Handle Cross-Context Integration**: Both synchronous and asynchronous patterns
4. **Build REST APIs**: Resource-oriented design with proper HTTP semantics
5. **Implement Authentication**: JWT with domain-specific permissions
6. **Generate Test Data**: Realistic scenarios for development and demos
7. **Write Integration Tests**: Multi-context workflow verification

## References

- [API Documentation](API.md) - Complete API reference for the cargo system
- [Design Guidelines](DESIGN_GUIDELINES.md) - Architectural principles and patterns
