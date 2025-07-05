# Go Hex

A template for building applications in Go with the assistance of an AI Agent, following Domain-Driven Design (DDD) principles, Hexagonal Architecture, and Event-Driven Architecture.

The sample implementation is a direct reflection of the design principles and architecture described in the [Design Guidelines](docs/DESIGN_GUIDELINES.md). These should be loaded into the AI Agent's context to ensure it can follow the design principles effectively.

The recommended LLM (Large Language Model) for this project is Claude Sonnet 4. It has not been tested with other LLMs.

## Goals

- Provide a robust foundation for building maintainable, testable, and scalable golang applications
- Provide a clear and relevant context that reduces the time required to get from bounded context specifications to a mock application

## Template Architecture

This template implements several key architectural patterns:

- **Domain-Driven Design**: Clear bounded contexts with rich domain models
- **Hexagonal Architecture**: Clean separation of business logic from infrastructure
- **Event-Driven Architecture**: Asynchronous inter-context communication
- **CQRS**: Separate command and query responsibilities
- **Anti-Corruption Layers**: Protected integration boundaries
- **REST API**: Modern HTTP API following REST principles
- **JWT Security**: Token-based authentication with role-based access

## Sample Implementation

This template includes a complete reference implementation of a **cargo shipping system** that demonstrates all the architectural patterns and design principles. The sample application includes three bounded contexts (Booking, Routing, and Handling) with real-world integration patterns.

**[View Sample Implementation Documentation →](docs/SAMPLE_IMPLEMENTATION.md)**

## Test the sample application

### Prerequisites

- devbox >= 0.14.2

### Quick Start

Start the sample application in mock mode with pre-populated test data:

```bash
devbox shell
just run mock
```

The server will start on `http://localhost:8080` with a fully functional cargo shipping API.

### Application Modes

The template supports two operational modes:

- **mock**: Pre-populated with realistic generated test data (ideal for demos and development)
- **live**: Clean repositories for real usage

```bash
# Mock mode (default for development)
just run mock

# Live mode
just run live
```

### Authentication

The template includes JWT-based authentication with role-based access control:

```bash
# Generate tokens for testing
go run tools/generate_test_token.go admin    # Full access
go run tools/generate_test_token.go user     # Standard operations  
go run tools/generate_test_token.go readonly # Read-only access
```

### Basic API Testing

```bash
# Check system health
curl http://localhost:8080/health

# Test authenticated endpoint (example from cargo shipping sample)
curl -H "Authorization: Bearer $JWT_TOKEN" \
     http://localhost:8080/api/v1/locations
```

## Configuration

The template uses environment variables for configuration:

- `APP_MODE`: Application mode (mock, live) - default: live
- `PORT`: HTTP server port (default: 8080)
- `JWT_SECRET`: JWT signing secret (required)
- `JWT_ISSUER`: JWT issuer (default: "go-hex-service")
- `JWT_AUDIENCE`: JWT audience (default: "go-hex-api")
- `LOG_LEVEL`: Logging level (debug, info, warn, error) - default: info

## Development Commands

Using the provided `justfile`:

```bash
# Build the application
just build

# Run in different modes
just run mock    # With test data
just run live    # Clean repositories

# Run tests
just test

# Run tests with coverage
just test-coverage

# Format code
just fmt

# Run linter
just lint

# Clean build artifacts
just clean

# Tidy dependencies
just tidy
```

## Documentation

- **[Sample Implementation](docs/SAMPLE_IMPLEMENTATION.md)** - Complete cargo shipping reference implementation
- **[Design Guidelines](docs/DESIGN_GUIDELINES.md)** - Architectural principles and patterns
- **[API Documentation](docs/API.md)** - Complete API reference

## Testing

Run all tests:

```bash
just test
```

Run tests with coverage:

```bash
just test-coverage
```

The template includes comprehensive integration tests that demonstrate multi-context workflows and proper domain separation.

## License

This project is licensed under the MIT License - see the [LICENSE.txt](LICENSE.txt) file for details.

## References

- [Domain-Driven Design by Eric Evans](https://domainlanguage.com/ddd/)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)

## Getting Started with Your Own Domain

**Define Your Bounded Contexts**: Identify the core business domains for your application. Create a technical specification for each context, detailing the domain models, operations, permissions and integration points
**Implement Primary Ports & Domain Models**: Use the AI Agent to create the primary ports and domain models for each bounded context based on the technical specifications.
**Create REST APIs**: Use the AI Agent to generate REST APIs for each context, including endpoints for all required operations
**Mock Application**: Use the AI Agent to implement a mock application that demonstrates the complete workflow of your business processes using the provided domain models and API endpoints. Populate it with realistic test data for development purposes.

Example prompts for the AI Agent:

- "Implement the primary ports and domain models for the bounded context described in {technical specification}."
- "Create a REST API for the {context name} bounded context with endpoints for {list of operations}."
- "Implement a mock application that demonstrates the complete workflow of {specific business process} using the provided domain models and API endpoints. Add realistic test data for development purposes."

## Improve your mock application

Once you have the basic mock application set up, you can enhance it by:

**Generate Secondary Ports and Repositories**: Use the AI Agent to generate secondary ports for each bounded context, which will allow the application to interact with external systems or services. Also, generate in-memory repositories for testing purposes.
**Update Mock Application**: Use the AI Agent to update the mock application to use the generated secondary ports and repositories. This will allow the mock application to simulate real-world interactions with external systems, such as databases or message queues.

## Transitioning to a Real Application

**Implement Real Repositories**: Use the AI Agent to implement real repositories for each bounded context. Use mock repositories for mock mode and real repositories for live mode.
**Implement Domain Logic**: Add domain logic to the primary ports and secondary ports, ensuring that the business rules are enforced correctly. The AI Agent can assist in generating this logic based on the domain models and requirements. However, resulting code should be reviewed and refined by developers to ensure it meets the specific needs of the application.
**Add Unit Tests**: Write unit tests for the domain logic and repositories to ensure correctness and maintainability. The AI Agent can help generate initial test cases, but developers should review and enhance them to cover edge cases and specific scenarios.
**Implement Integration Tests**: Write integration tests that cover the complete workflows across multiple bounded contexts, ensuring that the system behaves as expected in real-world scenarios. The AI Agent can assist in generating these tests based on the defined workflows and interactions between contexts, but developers should ensure that the tests cover all necessary scenarios and edge cases.
