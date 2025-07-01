# Go Hexagonal Architecture Template

This template launches a simple 'Greeter' service, which allows users to add friends and greet them. It follows hexagonal architecture, event-driven architecture, and DDD principles. Read the [architecture documentation](docs/ARCHITECTURE.md) for more information.

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

### External Authentication Integration

The service validates JWT tokens from external authentication providers. The JWT validation includes:

- **Signature Verification**: Validates token integrity (simplified for demo)
- **Issuer Validation**: Must match configured issuer (`go-hex-service`)
- **Audience Validation**: Must match configured audience (`go-hex-api`)
- **Expiration Check**: Tokens must not be expired
- **Claims Extraction**: Maps JWT claims to internal user model

The middleware expects standard JWT claims (`sub`, `iss`, `aud`, `exp`, `iat`) plus custom claims for `username`, `email`, `roles`, and `metadata`.
