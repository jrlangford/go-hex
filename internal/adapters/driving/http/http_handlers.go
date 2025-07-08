package http

import (
	"encoding/json"
	"fmt"
	"go_hex/internal/adapters/driving/http/middleware"
	bookingDomain "go_hex/internal/core/booking/domain"
	bookingPorts "go_hex/internal/core/booking/ports/primary"
	handlingDomain "go_hex/internal/core/handling/domain"
	handlingPrimary "go_hex/internal/core/handling/ports/primary"
	routingApp "go_hex/internal/core/routing/application"
	routingDomain "go_hex/internal/core/routing/domain"
	"go_hex/internal/support/validation"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Handler is the main HTTP handler for the cargo shipping application.
type Handler struct {
	authMiddleware        *middleware.AuthMiddleware
	bookingService        bookingPorts.BookingService
	routingService        *routingApp.RoutingApplicationService
	handlingReportService handlingPrimary.HandlingReportService
	handlingQueryService  handlingPrimary.HandlingEventQueryService
}

// NewHandler creates a new HTTP handler with the given services and middleware.
func NewHandler(
	authMiddleware *middleware.AuthMiddleware,
	bookingService bookingPorts.BookingService,
	routingService *routingApp.RoutingApplicationService,
	handlingReportService handlingPrimary.HandlingReportService,
	handlingQueryService handlingPrimary.HandlingEventQueryService,
) *Handler {
	return &Handler{
		authMiddleware:        authMiddleware,
		bookingService:        bookingService,
		routingService:        routingService,
		handlingReportService: handlingReportService,
		handlingQueryService:  handlingQueryService,
	}
}

// ErrorResponse represents a JSON error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a generic JSON success response.
type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

// writeErrorResponse is a helper to write standardized error responses
func (h *Handler) writeErrorResponse(w http.ResponseWriter, errorCode, message string, httpStatus int) {
	w.WriteHeader(httpStatus)
	errorResponse := ErrorResponse{
		Error:   errorCode,
		Message: message,
		Code:    httpStatus,
	}
	json.NewEncoder(w).Encode(errorResponse)
}

// extractResourceIDFromPath extracts resource ID from RESTful URL paths using net/http path utilities
// This is a more robust version that leverages Go's path manipulation
func (h *Handler) extractResourceIDFromPath(urlPath, basePattern string) (string, error) {
	// Clean the URL path using Go's path utilities
	cleanPath := path.Clean(urlPath)
	cleanBase := path.Clean(basePattern)
	
	// Remove query parameters
	if parsed, err := url.Parse(cleanPath); err == nil {
		cleanPath = parsed.Path
	}
	
	// Check if path starts with the base pattern
	if !strings.HasPrefix(cleanPath, cleanBase) {
		return "", fmt.Errorf("path %s does not match base pattern %s", cleanPath, cleanBase)
	}
	
	// Extract the ID part
	remainder := strings.TrimPrefix(cleanPath, cleanBase)
	remainder = strings.Trim(remainder, "/")
	
	// Split by "/" and take the first segment as the ID
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0], nil
	}
	
	return "", fmt.Errorf("no resource ID found in path %s", urlPath)
}

// getQueryParameter extracts a query parameter from the request URL
func (h *Handler) getQueryParameter(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// parseRequestBody parses JSON request body into the provided destination
func (h *Handler) parseRequestBody(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dest)
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks"`
}

// HealthHandler handles health check requests.
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthResponse := HealthResponse{
		Status:  "OK",
		Service: "Cargo Shipping System",
		Checks: map[string]string{
			"api":      "OK",
			"database": "OK", // Placeholder
		},
	}

	if err := json.NewEncoder(w).Encode(healthResponse); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DefaultHandler handles requests to undefined routes.
func (h *Handler) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	errorResponse := ErrorResponse{
		Error:   "Not Found",
		Message: "The requested resource was not found",
		Code:    http.StatusNotFound,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// InfoHandler provides information about the cargo shipping system.
func (h *Handler) InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := SuccessResponse{
		Status: "success",
		Data: map[string]interface{}{
			"application": "Cargo Shipping System",
			"version":     "1.0.0",
			"description": "A DDD-based cargo shipping system with Booking, Routing, and Handling contexts",
			"contexts": []string{
				"booking",
				"routing",
				"handling",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// BookCargoHandler handles cargo booking requests.
func (h *Handler) BookCargoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req BookCargoRequest
	if err := h.parseRequestBody(r, &req); err != nil {
		h.writeErrorResponse(w, "invalid_request", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.Validate(req); err != nil {
		h.writeErrorResponse(w, "validation_error", err.Error(), http.StatusBadRequest)
		return
	}

	// Book cargo
	cargo, err := h.bookingService.BookNewCargo(r.Context(), req.Origin, req.Destination, req.ArrivalDeadline)
	if err != nil {
		h.writeErrorResponse(w, "booking_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	response := BookCargoToResponse(cargo)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   response,
	})
}

// TrackCargoHandler handles cargo tracking requests.
func (h *Handler) TrackCargoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract tracking ID from URL path using improved extraction
	trackingIdStr, err := h.extractResourceIDFromPath(r.URL.Path, "/api/v1/cargos")
	if err != nil {
		h.writeErrorResponse(w, "invalid_request", "Tracking ID is required", http.StatusBadRequest)
		return
	}

	// Parse tracking ID
	trackingId, err := bookingDomain.TrackingIdFromString(trackingIdStr)
	if err != nil {
		h.writeErrorResponse(w, "invalid_tracking_id", "Invalid tracking ID format", http.StatusBadRequest)
		return
	}

	// Get cargo details
	cargo, err := h.bookingService.GetCargoDetails(r.Context(), trackingId)
	if err != nil {
		h.writeErrorResponse(w, "cargo_not_found", "Cargo not found", http.StatusNotFound)
		return
	}

	// Return response
	response := CargoToResponse(cargo)
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   response,
	})
}

// RequestRouteCandidatesHandler handles route candidate requests.
func (h *Handler) RequestRouteCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req struct {
		TrackingId string `json:"trackingId" validate:"required"`
	}
	if err := h.parseRequestBody(r, &req); err != nil {
		h.writeErrorResponse(w, "invalid_request", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.Validate(req); err != nil {
		h.writeErrorResponse(w, "validation_error", err.Error(), http.StatusBadRequest)
		return
	}

	// Parse tracking ID
	trackingId, err := bookingDomain.TrackingIdFromString(req.TrackingId)
	if err != nil {
		h.writeErrorResponse(w, "invalid_tracking_id", "Invalid tracking ID format", http.StatusBadRequest)
		return
	}

	// Get route candidates
	candidates, err := h.bookingService.RequestRouteCandidates(r.Context(), trackingId)
	if err != nil {
		h.writeErrorResponse(w, "route_search_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to DTOs
	routes := make([]ItineraryDTO, len(candidates))
	for i, candidate := range candidates {
		routes[i] = *ItineraryToDTO(candidate)
	}

	// Return response as array of itineraries (according to API spec)
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   routes,
	})
}

// SubmitHandlingReportHandler handles handling report submissions.
func (h *Handler) SubmitHandlingReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req HandlingEventRequest
	if err := h.parseRequestBody(r, &req); err != nil {
		h.writeErrorResponse(w, "invalid_request", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.Validate(req); err != nil {
		h.writeErrorResponse(w, "validation_error", err.Error(), http.StatusBadRequest)
		return
	}

	// Validate tracking ID format
	_, err := bookingDomain.TrackingIdFromString(req.TrackingId)
	if err != nil {
		h.writeErrorResponse(w, "invalid_tracking_id", "Invalid tracking ID format", http.StatusBadRequest)
		return
	}

	// Validate completion time format
	_, err = time.Parse(time.RFC3339, req.CompletionTime)
	if err != nil {
		h.writeErrorResponse(w, "invalid_time", "Invalid completion time format, expected RFC3339", http.StatusBadRequest)
		return
	}

	// Submit handling report
	report := handlingPrimary.HandlingReport{
		TrackingId:     req.TrackingId,
		EventType:      req.EventType,
		Location:       req.Location,
		VoyageNumber:   req.VoyageNumber,
		CompletionTime: req.CompletionTime,
	}

	err = h.handlingReportService.SubmitHandlingReport(r.Context(), report)
	if err != nil {
		h.writeErrorResponse(w, "handling_report_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response
	response := HandlingEventResponse{
		EventId:        "", // We don't get the event ID back from this interface
		TrackingId:     req.TrackingId,
		EventType:      req.EventType,
		Location:       req.Location,
		VoyageNumber:   req.VoyageNumber,
		CompletionTime: req.CompletionTime,
		RegisteredAt:   time.Now().Format(time.RFC3339),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   response,
	})
}

// AssignRouteHandler handles route assignment to cargo.
func (h *Handler) AssignRouteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Handle the special case where URL has "/route" suffix
	urlPath := strings.TrimSuffix(r.URL.Path, "/route")

	// Extract tracking ID from URL path using improved extraction
	trackingIdStr, err := h.extractResourceIDFromPath(urlPath, "/api/v1/cargos")
	if err != nil {
		h.writeErrorResponse(w, "invalid_request", "Tracking ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req AssignRouteRequest
	if err := h.parseRequestBody(r, &req); err != nil {
		h.writeErrorResponse(w, "invalid_request", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.Validate(req); err != nil {
		h.writeErrorResponse(w, "validation_error", err.Error(), http.StatusBadRequest)
		return
	}

	// Parse tracking ID
	trackingId, err := bookingDomain.TrackingIdFromString(trackingIdStr)
	if err != nil {
		h.writeErrorResponse(w, "invalid_tracking_id", "Invalid tracking ID format", http.StatusBadRequest)
		return
	}

	// Convert request legs to domain itinerary
	var legs []bookingDomain.Leg
	for _, legReq := range req.Legs {
		// Parse times
		departureTime, err := time.Parse(time.RFC3339, legReq.LoadTime)
		if err != nil {
			h.writeErrorResponse(w, "invalid_time", "Invalid load time format", http.StatusBadRequest)
			return
		}
		arrivalTime, err := time.Parse(time.RFC3339, legReq.UnloadTime)
		if err != nil {
			h.writeErrorResponse(w, "invalid_time", "Invalid unload time format", http.StatusBadRequest)
			return
		}

		// Create leg
		leg, err := bookingDomain.NewLeg(legReq.VoyageNumber, legReq.LoadLocation, legReq.UnloadLocation, departureTime, arrivalTime)
		if err != nil {
			h.writeErrorResponse(w, "invalid_leg", err.Error(), http.StatusBadRequest)
			return
		}
		legs = append(legs, leg)
	}

	// Create itinerary
	itinerary, err := bookingDomain.NewItinerary(legs)
	if err != nil {
		h.writeErrorResponse(w, "invalid_itinerary", err.Error(), http.StatusBadRequest)
		return
	}

	// Assign route to cargo
	err = h.bookingService.AssignRouteToCargo(r.Context(), trackingId, itinerary)
	if err != nil {
		h.writeErrorResponse(w, "route_assignment_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   map[string]string{"message": "Route assigned successfully"},
	})
}

// ListCargoHandler handles listing all cargo.
func (h *Handler) ListCargoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get unrouted cargo (for simplicity, we'll just return unrouted cargo)
	cargoList, err := h.bookingService.ListUnroutedCargo(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "list_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to responses
	responses := make([]CargoDetailsResponse, len(cargoList))
	for i, cargo := range cargoList {
		responses[i] = CargoToResponse(cargo)
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   responses,
	})
}

// ListVoyagesHandler handles GET /api/v1/voyages
func (h *Handler) ListVoyagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get all voyages from the routing service
	voyages, err := h.routingService.ListAllVoyages(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "INTERNAL_ERROR", "Failed to list voyages", http.StatusInternalServerError)
		return
	}

	// Convert voyages to response format
	voyageResponses := make([]VoyageResponse, len(voyages))
	for i, voyage := range voyages {
		voyageResponses[i] = VoyageToResponse(voyage)
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   voyageResponses,
	})
}

// ListLocationsHandler handles GET /api/v1/locations
func (h *Handler) ListLocationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get all locations from the routing service
	locations, err := h.routingService.ListAllLocations(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "INTERNAL_ERROR", "Failed to list locations", http.StatusInternalServerError)
		return
	}

	// Convert locations to response format
	locationResponses := make([]LocationResponse, len(locations))
	for i, location := range locations {
		locationResponses[i] = LocationToResponse(location)
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   locationResponses,
	})
}

// ListHandlingEventsHandler handles GET /api/v1/handling-events
func (h *Handler) ListHandlingEventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check for tracking_id query parameter
	trackingID := h.getQueryParameter(r, "tracking_id")

	var events []HandlingEventResponse

	if trackingID != "" {
		// Get events for specific cargo
		handlingHistory, queryErr := h.handlingQueryService.GetHandlingHistory(r.Context(), trackingID)
		if queryErr != nil {
			h.writeErrorResponse(w, "INTERNAL_ERROR", "Failed to get handling events", http.StatusInternalServerError)
			return
		}

		// Convert handling history to response format
		events = make([]HandlingEventResponse, len(handlingHistory.Events))
		for i, event := range handlingHistory.Events {
			events[i] = HandlingEventToResponse(event)
		}
	} else {
		// Get all events using the new method
		allEvents, queryErr := h.handlingQueryService.ListAllHandlingEvents(r.Context())
		if queryErr != nil {
			h.writeErrorResponse(w, "INTERNAL_ERROR", "Failed to list all handling events", http.StatusInternalServerError)
			return
		}

		events = make([]HandlingEventResponse, len(allEvents))
		for i, event := range allEvents {
			events[i] = HandlingEventToResponse(event)
		}
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   events,
	})
}

// AuthMeResponse represents the response for /auth/me endpoint.
type AuthMeResponse struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	Roles    []string `json:"roles"`
}

// AuthMeHandler handles /auth/me requests to introspect tokens.
func (h *Handler) AuthMeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract claims from context
	claims := middleware.GetTokenClaims(r.Context())
	if claims == nil {
		h.writeErrorResponse(w, "unauthorized", "Authentication required", http.StatusUnauthorized)
		return
	}

	// Create response
	response := AuthMeResponse{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   response,
	})
}

// Conversion functions for response types

func VoyageToResponse(voyage interface{}) VoyageResponse {
	// Type assert to proper domain type
	if v, ok := voyage.(routingDomain.Voyage); ok {
		schedule := v.GetSchedule()
		legs := make([]LegDTO, len(schedule.Movements))

		for i, movement := range schedule.Movements {
			legs[i] = LegDTO{
				VoyageNumber:   v.GetVoyageNumber().String(),
				LoadLocation:   movement.DepartureLocation.String(),
				UnloadLocation: movement.ArrivalLocation.String(),
				LoadTime:       movement.DepartureTime.Format(time.RFC3339),
				UnloadTime:     movement.ArrivalTime.Format(time.RFC3339),
			}
		}

		return VoyageResponse{
			VoyageNumber: v.GetVoyageNumber().String(),
			Schedule:     legs,
		}
	}

	// Fallback for interface{} parameter
	return VoyageResponse{
		VoyageNumber: "UNKNOWN",
		Schedule:     []LegDTO{},
	}
}

func LocationToResponse(location interface{}) LocationResponse {
	// Type assert to proper domain type
	if l, ok := location.(routingDomain.Location); ok {
		return LocationResponse{
			Code: l.GetUnLocode().String(),
			Name: l.GetName(),
		}
	}

	// Fallback for interface{} parameter
	return LocationResponse{
		Code: "UNKNOWN",
		Name: "Unknown Location",
	}
}

func HandlingEventToResponse(event interface{}) HandlingEventResponse {
	// Type assert to proper domain type
	if e, ok := event.(handlingDomain.HandlingEvent); ok {
		return HandlingEventResponse{
			EventId:        e.GetEventId().String(),
			TrackingId:     e.GetTrackingId(),
			EventType:      string(e.GetEventType()),
			Location:       e.GetLocation(),
			VoyageNumber:   e.GetVoyageNumber(),
			CompletionTime: e.GetCompletionTime().Format(time.RFC3339),
			RegisteredAt:   e.GetRegistrationTime().Format(time.RFC3339),
		}
	}

	// Fallback for interface{} parameter
	return HandlingEventResponse{
		EventId:        "unknown",
		TrackingId:     "unknown",
		EventType:      "UNKNOWN",
		Location:       "unknown",
		VoyageNumber:   "",
		CompletionTime: time.Now().Format(time.RFC3339),
		RegisteredAt:   time.Now().Format(time.RFC3339),
	}
}
