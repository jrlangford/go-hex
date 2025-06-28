# Go Hexagonal Architecture Template

This template launches a simple 'Greeter' service, which allows users to add friends and greet them. It demonstrates a basic implementation of the hexagonal architecture pattern in Go, focusing on clean separation of concerns between business logic and external interfaces.

## Requirements

- devbox >= 0.14.2

## Getting Started

Enter the devbox shell. This will install system dependencies.

```bash
devbox shell
```

Run the server.

```bash
just run
```

### Generate an Authentication Token

Most endpoints require authentication. Generate a test JWT token first:

```bash
# Generate an admin token (expires in 24 hours)
go run tools/generate_test_token.go admin-123 admin "admin,user" admin@example.com 24

# Copy the token from the output and use it in subsequent requests
```

For convenience, you can store the token in a variable:

```bash
TOKEN=$(go run tools/generate_test_token.go admin-123 admin "admin,user" admin@example.com 1 | grep -A1 "Token:" | tail -1)
```

### API Operations

Add a friend to the list (requires admin role).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     -X POST localhost:8080/friends \
     -d '{"name": "Alice", "title": "Ms."}'
```
The endpoint's response will include the newly created friend.

Get a custom greeting (authentication optional but provides enhanced response).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/greet?id={uuid}
```

Get all friends (requires authentication).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/friends
```

Get a specific friend (requires authentication).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/friends/{uuid}
```

Update a friend (requires admin role).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     -X PUT localhost:8080/friends/{uuid} \
     -d '{"name": "Alice Updated", "title": "Dr."}'
```

Delete a friend (requires admin role).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     -X DELETE localhost:8080/friends/{uuid}
```

Get current user information (requires authentication).

```bash
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/auth/me
```

Check application health (no authentication required).

```bash
curl localhost:8080/health
```

## Authentication

The service uses JWT (JSON Web Token) based authentication. Tokens must be provided in the Authorization header using the Bearer scheme.

### JWT Configuration

The authentication middleware is configured with:
- **Secret Key**: `your-secret-key` (configurable)
- **Issuer**: `go-hex-service` 
- **Audience**: `go-hex-api`

### JWT Token Structure

Tokens must contain the following claims:
```json
{
  "sub": "user-123",           // UserID
  "username": "john.doe",
  "email": "john.doe@example.com", 
  "roles": ["user", "admin"],
  "metadata": {"dept": "engineering"},
  "iss": "go-hex-service",     // Must match configured issuer
  "aud": "go-hex-api",         // Must match configured audience
  "iat": 1640991600,           // Issued at (Unix timestamp)
  "exp": 1640995200            // Expires at (Unix timestamp)
}
```

### Generating Test Tokens

Use the provided token generator for testing:

```bash
# Generate admin token (expires in 24 hours)
go run tools/generate_test_token.go admin-123 admin "admin,user" admin@example.com 24

# Generate regular user token (expires in 1 hour)
go run tools/generate_test_token.go user-456 john.doe "user" john@example.com 1
```

### Protected Endpoints

Most endpoints require JWT authentication. Use the Authorization header:

```bash
# Generate a test token first
TOKEN=$(go run tools/generate_test_token.go admin-123 admin "admin,user" admin@example.com 1 | grep -A1 "Token:" | tail -1)

# Create a friend (admin only)
curl -H "Authorization: Bearer $TOKEN" \
     -X POST localhost:8080/friends \
     -d '{"name": "Alice", "title": "Ms."}'

# Get all friends (any authenticated user)
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/friends

# Get specific friend (any authenticated user)
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/friends/{uuid}

# Update friend (admin only)
curl -H "Authorization: Bearer $TOKEN" \
     -X PUT localhost:8080/friends/{uuid} \
     -d '{"name": "Alice Updated", "title": "Dr."}'

# Delete friend (admin only)
curl -H "Authorization: Bearer $TOKEN" \
     -X DELETE localhost:8080/friends/{uuid}

# Get current user info
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/auth/me

# Enhanced greeting (with optional auth)
curl -H "Authorization: Bearer $TOKEN" \
     localhost:8080/greet?id={uuid}
```

### Public Endpoints

These endpoints don't require authentication:

- `GET /health` - Health check

### External Authentication Integration

The service validates JWT tokens from external authentication providers. The JWT validation includes:

- **Signature Verification**: Validates token integrity (simplified for demo)
- **Issuer Validation**: Must match configured issuer (`go-hex-service`)
- **Audience Validation**: Must match configured audience (`go-hex-api`)
- **Expiration Check**: Tokens must not be expired
- **Claims Extraction**: Maps JWT claims to internal user model

For production use, integrate with:
- **Auth0**: Configure with your Auth0 domain and API identifier
- **AWS Cognito**: Use Cognito User Pool tokens
- **Firebase Auth**: Validate Firebase ID tokens
- **Custom JWT Provider**: Any service that issues RFC 7519 compliant JWTs

The middleware expects standard JWT claims (`sub`, `iss`, `aud`, `exp`, `iat`) plus custom claims for `username`, `email`, `roles`, and `metadata`.

## API Response Format

All API endpoints return JSON responses with a consistent structure:

### Success Response
```json
{
  "status": "success",
  "data": {
    // Response data here
  }
}
```

### Error Response
```json
{
  "error": "Bad Request",
  "message": "Detailed error message",
  "code": 400
}
```

### Health Check Response
```json
{
  "status": "healthy",
  "service": "go_hex",
  "checks": {
    "repository": "healthy"
  }
}
```

### Authentication Error Response
```json
{
  "error": "Unauthorized",
  "message": "Authentication token required",
  "code": 401
}
```

### Authorization Error Response
```json
{
  "error": "Forbidden", 
  "message": "Insufficient permissions",
  "code": 403
}
```

### User Info Response (/auth/me)
```json
{
  "status": "success",
  "data": {
    "user_id": "admin-123",
    "username": "admin",
    "email": "admin@example.com",
    "roles": ["admin", "user"],
    "metadata": {
      "department": "IT"
    }
  }
}
```

## Architecture

This project is structured using the hexagonal (ports and adapters) architecture pattern for Go applications.

## Structure

- `cmd/` - Application entry point (main.go)
- `internal/core/` - Contains business logic
- `internal/core/domain/` - Domain models organized by bounded context
  - `internal/core/domain/authorization/` - Authorization and permission domain logic
  - `internal/core/domain/{sub-context-1}/` - Sub-context 1 domain logic
  - `internal/core/domain/{sub-context-2}/` - Sub-context 2 domain logic
  - ...
  - `internal/core/domain/{sub-context-N}/` - Sub-context N domain logic
  - `internal/core/domain/common/` - Shared domain infrastructure (base models, errors)
- `internal/core/ports/` - Interface definition for primary and secondary ports
- `internal/core/ports/primary` - The primary ports
- `internal/core/ports/secondary` - The secondary ports
- `internal/core/application/` - Application services/use cases
- `internal/adapters` - Contains adapters
- `internal/adapters/driving` - Driving Adapters: HTTP, CLI, etc.
- `internal/adapters/driven` - Driven Adapters: Frameworks, DB, external APIs
- `examples/` - Example usage of the service
- `build/` - Artifacts from the build process


---

For more details on the architecture and design decisions in this project, see [ARCHITECTURE.md](docs/ARCHITECTURE.md).
