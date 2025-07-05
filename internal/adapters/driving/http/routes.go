package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Public endpoints (no authentication required)
	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/info", handler.InfoHandler)

	// Cargo Booking Context endpoints - REST compliant
	// GET/POST /api/v1/cargos - list/create cargo
	mux.HandleFunc("/api/v1/cargos", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.authMiddleware.RequireAuth(handler.BookCargoHandler)(w, r)
		case http.MethodGet:
			handler.authMiddleware.RequireAuth(handler.ListCargoHandler)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// GET /api/v1/cargos/{trackingId} - get specific cargo
	// PUT /api/v1/cargos/{trackingId}/route - assign route to cargo
	mux.HandleFunc("/api/v1/cargos/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if it's a route assignment request
		if strings.HasSuffix(path, "/route") && r.Method == http.MethodPut {
			trackingID := extractIDFromPath(strings.TrimSuffix(path, "/route"), "/api/v1/cargos/")
			if trackingID == "" {
				writeBadRequestError(w, "Tracking ID is required")
				return
			}
			handler.authMiddleware.RequireAuth(handler.AssignRouteHandler)(w, r)
			return
		}

		// Otherwise, it's a GET for specific cargo
		trackingID := extractIDFromPath(path, "/api/v1/cargos/")
		if trackingID == "" {
			writeBadRequestError(w, "Tracking ID is required")
			return
		}

		switch r.Method {
		case http.MethodGet:
			handler.authMiddleware.RequireAuth(handler.TrackCargoHandler)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// Routing Context endpoints - REST compliant
	// POST /api/v1/route-candidates - request route candidates
	mux.HandleFunc("/api/v1/route-candidates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.authMiddleware.RequireAuth(handler.RequestRouteCandidatesHandler)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// GET /api/v1/voyages - list voyages
	mux.HandleFunc("/api/v1/voyages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.authMiddleware.RequireAuth(handler.ListVoyagesHandler)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// GET /api/v1/locations - list locations
	mux.HandleFunc("/api/v1/locations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.authMiddleware.RequireAuth(handler.ListLocationsHandler)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// Handling Context endpoints - REST compliant
	// POST /api/v1/handling-events - submit handling event
	// GET /api/v1/handling-events - list handling events with optional filtering
	// GET /api/v1/handling-events?tracking_id={id} - list handling events for cargo
	mux.HandleFunc("/api/v1/handling-events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.authMiddleware.RequireAuth(handler.SubmitHandlingReportHandler)(w, r)
		case http.MethodGet:
			handler.authMiddleware.RequireAuth(handler.ListHandlingEventsHandler)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// Default handler for undefined routes
	mux.HandleFunc("/", handler.DefaultHandler)
}

// extractIDFromPath extracts resource ID from REST URL path
func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	id = strings.Split(id, "/")[0] // Take only the first segment after prefix
	return id
}

func writeMethodNotAllowedError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)

	errorResponse := ErrorResponse{
		Error:   "Method Not Allowed",
		Message: "The requested HTTP method is not allowed for this resource",
		Code:    http.StatusMethodNotAllowed,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func writeInternalServerError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	errorResponse := ErrorResponse{
		Error:   "Internal Server Error",
		Message: message,
		Code:    http.StatusInternalServerError,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func writeBadRequestError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	errorResponse := ErrorResponse{
		Error:   "Bad Request",
		Message: message,
		Code:    http.StatusBadRequest,
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func writeNotFoundError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	errorResponse := ErrorResponse{
		Error:   "Not Found",
		Message: message,
		Code:    http.StatusNotFound,
	}

	json.NewEncoder(w).Encode(errorResponse)
}
