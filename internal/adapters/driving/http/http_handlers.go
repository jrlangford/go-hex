package http

import (
	"encoding/json"
	"go_hex/internal/adapters/driving/http/middleware"
	bookingApp "go_hex/internal/core/booking/application"
	bookingDomain "go_hex/internal/core/booking/domain"
	handlingDomain "go_hex/internal/core/handling/domain"
	handlingPrimary "go_hex/internal/core/handling/ports/primary"
	routingApp "go_hex/internal/core/routing/application"
	routingDomain "go_hex/internal/core/routing/domain"
	"go_hex/internal/support/validation"
	"net/http"
	"strings"
	"time"
)

// Handler is the main HTTP handler for the cargo shipping application.
type Handler struct {
	authMiddleware        *middleware.AuthMiddleware
	bookingService        *bookingApp.BookingApplicationService
	routingService        *routingApp.RoutingApplicationService
	handlingReportService handlingPrimary.HandlingReportService
	handlingQueryService  handlingPrimary.HandlingEventQueryService
}

// NewHandler creates a new HTTP handler with the given services and middleware.
func NewHandler(
	authMiddleware *middleware.AuthMiddleware,
	bookingService *bookingApp.BookingApplicationService,
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

// extractTrackingIdFromPath extracts tracking ID from URL path
func (h *Handler) extractTrackingIdFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// Extract tracking ID from URL path
	trackingIdStr := h.extractTrackingIdFromPath(r.URL.Path)
	if trackingIdStr == "" {
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

	// Extract tracking ID from URL path
	trackingIdStr := h.extractTrackingIdFromPath(r.URL.Path)
	if trackingIdStr == "" {
		h.writeErrorResponse(w, "invalid_request", "Tracking ID is required", http.StatusBadRequest)
		return
	}

	// Parse tracking ID
	trackingId, err := bookingDomain.TrackingIdFromString(trackingIdStr)
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

	// Return response
	response := RouteCandidatesResponse{
		TrackingId: trackingIdStr,
		Routes:     routes,
	}

	json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   response,
	})
}

// SubmitHandlingReportHandler handles handling report submissions.
func (h *Handler) SubmitHandlingReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req HandlingEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// Extract tracking ID from URL path
	trackingIdStr := h.extractTrackingIdFromPath(r.URL.Path)
	if trackingIdStr == "" {
		h.writeErrorResponse(w, "invalid_request", "Tracking ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req AssignRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "invalid_request", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validation.Validate(req); err != nil {
		h.writeErrorResponse(w, "validation_error", err.Error(), http.StatusBadRequest)
		return
	}

	// Parse tracking ID (validate format)
	_, err := bookingDomain.TrackingIdFromString(trackingIdStr)
	if err != nil {
		h.writeErrorResponse(w, "invalid_tracking_id", "Invalid tracking ID format", http.StatusBadRequest)
		return
	}

	// TODO: For now, we'll return success but the actual itinerary creation
	// would require getting the route candidate by ID and converting it to an Itinerary
	// This would typically involve the routing service
	h.writeErrorResponse(w, "not_implemented", "Route assignment not fully implemented", http.StatusNotImplemented)
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
	trackingID := r.URL.Query().Get("tracking_id")

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
