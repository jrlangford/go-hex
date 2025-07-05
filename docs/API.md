# Cargo Shipping System API Documentation

This document provides comprehensive information about all available API endpoints for the cargo shipping system, organized by bounded contexts.

## Table of Contents

- [Authentication](#authentication)
- [General Endpoints](#general-endpoints)
- [Booking Context](#booking-context)
- [Routing Context](#routing-context)
- [Handling Context](#handling-context)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

**Authentication is validated at the application level** - all business logic operations check permissions before executing.

### Roles

- **admin**: Full access to all operations including route assignment and handling submission
- **user**: Standard cargo operations (booking, viewing, tracking)
- **readonly**: View-only access to cargo, voyages, and locations

## General Endpoints

### GET /health

Returns the health status of the application.

**Authentication:** Not required

**Response:**
```json
{
  "status": "OK",
  "service": "Cargo Shipping System",
  "checks": {
    "api": "OK",
    "database": "OK"
  }
}
```

### GET /info

Returns information about the cargo shipping system.

**Authentication:** Not required

**Response:**
```json
{
  "status": "success",
  "data": {
    "application": "Cargo Shipping System",
    "version": "1.0.0",
    "description": "A DDD-based cargo shipping system with Booking, Routing, and Handling contexts",
    "contexts": ["booking", "routing", "handling"]
  }
}
```

## Booking Context

### POST /api/v1/cargos

Books new cargo for shipment.

**Authentication:** Required (user, admin)
**Permission:** book_cargo

**Request Body:**
```json
{
  "origin": "SESTO",
  "destination": "USNYC",
  "arrivalDeadline": "2024-12-31T23:59:59Z"
}
```

**Response:** `201 Created`
```json
{
  "status": "success",
  "data": {
    "trackingId": "b6865953-1eb8-43c3-9cfa-9cb8ffa8e718",
    "origin": "SESTO",
    "destination": "USNYC",
    "arrivalDeadline": "2024-12-31T23:59:59Z",
    "routingStatus": "NOT_ROUTED",
    "transportStatus": "NOT_RECEIVED",
    "isOnTrack": false,
    "isMisdirected": false,
    "isUnloadedAtDestination": false
  }
}
```

### GET /api/v1/cargos

Retrieves all booked cargo.

**Authentication:** Required (user, admin, readonly)
**Permission:** view_cargo

**Query Parameters:**
- `status` (optional): Filter by routing status (`NOT_ROUTED`, `ROUTED`, `MISROUTED`)
- `limit` (optional): Number of results (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": [
    {
      "trackingId": "b6865953-1eb8-43c3-9cfa-9cb8ffa8e718",
      "origin": "SESTO",
      "destination": "USNYC",
      "arrivalDeadline": "2024-12-31T23:59:59Z",
      "routingStatus": "NOT_ROUTED",
      "transportStatus": "NOT_RECEIVED",
      "isOnTrack": false,
      "isMisdirected": false,
      "isUnloadedAtDestination": false
    }
  ]
}
```

### GET /api/v1/cargos/{trackingId}

Retrieves specific cargo details.

**Authentication:** Required (user, admin, readonly)
**Permission:** view_cargo

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": {
    "trackingId": "b6865953-1eb8-43c3-9cfa-9cb8ffa8e718",
    "origin": "SESTO",
    "destination": "USNYC",
    "arrivalDeadline": "2024-12-31T23:59:59Z",
    "routingStatus": "NOT_ROUTED",
    "transportStatus": "NOT_RECEIVED",
    "isOnTrack": false,
    "isMisdirected": false,
    "isUnloadedAtDestination": false,
    "itinerary": null,
    "delivery": {
      "transportStatus": "NOT_RECEIVED",
      "currentVoyage": "",
      "currentLocation": "",
      "isMisdirected": false,
      "estimatedArrival": "0001-01-01T00:00:00Z",
      "isOnTrack": false,
      "isUnloadedAtDestination": false
    }
  }
}
```

### PUT /api/v1/cargos/{trackingId}/route

Assigns a route to cargo.

**Authentication:** Required (user, admin)
**Permission:** assign_route

**Request Body:**
```json
{
  "legs": [
    {
      "voyageNumber": "V001",
      "from": "SESTO",
      "to": "DEHAM", 
      "departureTime": "2024-01-20T08:00:00Z",
      "arrivalTime": "2024-01-21T16:00:00Z"
    },
    {
      "voyageNumber": "V002",
      "from": "DEHAM", 
      "to": "USNYC",
      "departureTime": "2024-01-22T10:00:00Z",
      "arrivalTime": "2024-01-25T14:00:00Z"
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "message": "Route assigned successfully"
}
```

## Routing Context

### POST /api/v1/route-candidates

Finds possible routes for cargo.

**Authentication:** Required (user, admin)
**Permission:** plan_routes

**Request Body:**
```json
{
  "trackingId": "b6865953-1eb8-43c3-9cfa-9cb8ffa8e718"
}
```

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": [
    {
      "legs": [
        {
          "voyageNumber": "V001",
          "from": "SESTO",
          "to": "DEHAM",
          "departureTime": "2024-01-20T08:00:00Z",
          "arrivalTime": "2024-01-21T16:00:00Z"
        },
        {
          "voyageNumber": "V002", 
          "from": "DEHAM",
          "to": "USNYC",
          "departureTime": "2024-01-22T10:00:00Z",
          "arrivalTime": "2024-01-25T14:00:00Z"
        }
      ]
    }
  ]
}
```

### GET /api/v1/voyages

Lists available voyages.

**Authentication:** Required (user, admin, readonly)
**Permission:** view_voyages

**Query Parameters:**
- `from` (optional): Filter by departure location
- `to` (optional): Filter by arrival location  
- `departure_date` (optional): Filter by departure date (YYYY-MM-DD)

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": [
    {
      "voyageNumber": "V001",
      "carrierMovements": [
        {
          "from": "SESTO",
          "to": "DEHAM",
          "departureTime": "2024-01-20T08:00:00Z",
          "arrivalTime": "2024-01-21T16:00:00Z"
        }
      ]
    }
  ]
}
```

### GET /api/v1/locations

Lists all shipping locations.

**Authentication:** Required (user, admin, readonly)
**Permission:** view_locations

**Response:** `200 OK`
```json
{
  "status": "success", 
  "data": [
    {
      "unLocode": "SESTO",
      "name": "Stockholm"
    },
    {
      "unLocode": "DEHAM", 
      "name": "Hamburg"
    },
    {
      "unLocode": "USNYC",
      "name": "New York"
    }
  ]
}
```

## Handling Context

### POST /api/v1/handling-events

Registers a new handling event.

**Authentication:** Required (user, admin)
**Permission:** submit_handling

**Request Body:**
```json
{
  "trackingId": "b6865953-1eb8-43c3-9cfa-9cb8ffa8e718",
  "eventType": "LOAD",
  "location": "SESTO",
  "voyageNumber": "V001",
  "completionTime": "2024-01-20T09:30:00Z"
}
```

**Event Types:**
- `RECEIVE`: Cargo received at port
- `LOAD`: Cargo loaded onto vessel
- `UNLOAD`: Cargo unloaded from vessel  
- `CLAIM`: Cargo claimed by consignee
- `CUSTOMS`: Cargo processed through customs

**Response:** `201 Created`
```json
{
  "status": "success",
  "message": "Handling event submitted successfully"
}
```

### GET /api/v1/handling-events

Retrieves handling events with optional filtering.

**Authentication:** Required (user, admin, readonly)
**Permission:** view_handling

**Query Parameters:**
- `tracking_id` (optional): Filter by cargo tracking ID

**Response:** `200 OK`
```json
{
  "status": "success",
  "data": [
    {
      "id": "evt_123456",
      "trackingId": "b6865953-1eb8-43c3-9cfa-9cb8ffa8e718",
      "eventType": "LOAD",
      "location": "SESTO",
      "voyageNumber": "V001",
      "completionTime": "2024-01-20T09:30:00Z",
      "registrationTime": "2024-01-20T09:32:15Z"
    }
  ]
}
```

## Error Handling

All endpoints return consistent error responses:

### 400 Bad Request
```json
{
  "error": "Bad Request",
  "message": "Invalid request data",
  "code": 400
}
```

### 401 Unauthorized  
```json
{
  "error": "Unauthorized",
  "message": "Authentication required",
  "code": 401
}
```

### 403 Forbidden
```json
{
  "error": "Forbidden", 
  "message": "Insufficient permissions",
  "code": 403
}
```

### 404 Not Found
```json
{
  "error": "Not Found",
  "message": "The requested resource was not found",
  "code": 404
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal Server Error",
  "message": "An unexpected error occurred",
  "code": 500
}
```

## Examples

### Generate JWT Token for Testing

Use the provided token generator with predefined roles:

```bash
# For full access (all operations)
go run tools/generate_test_token.go admin

# For standard operations (booking, viewing, tracking)
go run tools/generate_test_token.go user

# For read-only access (viewing only)
go run tools/generate_test_token.go readonly

# For testing with all roles combined
go run tools/generate_test_token.go super
```

Or create custom tokens:

```bash
go run tools/generate_test_token.go user-123 test.user admin,user test@example.com 24
```

### Complete Cargo Booking Flow

1. **Book cargo:**
```bash
curl -X POST http://localhost:8080/api/v1/cargos \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "SESTO",
    "destination": "USNYC",
    "arrivalDeadline": "2024-12-31T23:59:59Z"
  }'
```

2. **Find route candidates:**
```bash
curl -X POST http://localhost:8080/api/v1/route-candidates \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "trackingId": "{trackingId}"
  }'
```

3. **Assign route to cargo:**
```bash
curl -X PUT http://localhost:8080/api/v1/cargos/{trackingId}/route \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "legs": [
      {
        "voyageNumber": "V001",
        "from": "SESTO",
        "to": "DEHAM",
        "departureTime": "2024-01-20T08:00:00Z",
        "arrivalTime": "2024-01-21T16:00:00Z"
      }
    ]
  }'
```

4. **Register handling events:**
```bash
curl -X POST http://localhost:8080/api/v1/handling-events \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "trackingId": "{trackingId}",
    "eventType": "RECEIVE",
    "location": "SESTO",
    "completionTime": "2024-01-19T14:00:00Z"
  }'
```

5. **Track cargo:**
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/cargos/{trackingId}
```

### Using the justfile for Development

All development commands should be run using the provided `justfile`:

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

# Clean build artifacts
just clean
```

## Rate Limiting

Currently not implemented, but planned for future versions:
- 100 requests per minute for authenticated users
- 10 requests per minute for unauthenticated endpoints

## Pagination

All list endpoints support pagination:
- `limit`: Number of items per page (max 100, default 50)
- `offset`: Number of items to skip (default 0)

## API Versioning

The current API is version 1. Future versions will be accessible via:
- Header: `Accept: application/vnd.cargo-api.v2+json`
- URL: `http://localhost:8080/v2/api/v1/cargos`

## Application Modes

The system supports multiple operational modes:

- **mock**: Pre-populated with realistic generated test data (ideal for demos)
- **live**: Clean repositories for real usage

Start the application in different modes using the `justfile`:

```bash
just run mock    # Default for development
just run live    # Live mode
```

---

**Note:** This cargo shipping system implements Domain-Driven Design principles with clear bounded contexts for Booking, Routing, and Handling operations. Authentication is validated at the application level, ensuring business logic is properly protected.
