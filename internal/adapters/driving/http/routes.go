package http

import (
	"encoding/json"
	"net/http"
)

// RegisterRoutes registers HTTP routes for the Handler.
func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Public endpoints (no authentication required)
	mux.HandleFunc("/health", handler.HealthCheck)

	// Protected endpoints (authentication required)
	mux.HandleFunc("/friends", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// Creating friends requires admin role
			handler.authMiddleware.RequireRole("admin")(handler.AddFriend)(w, r)
		case http.MethodGet:
			// Reading all friends requires authentication (any role)
			handler.authMiddleware.RequireAuth(handler.GetAllFriends)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	mux.HandleFunc("/friends/{id}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Reading specific friend requires authentication
			handler.authMiddleware.RequireAuth(handler.GetFriend)(w, r)
		case http.MethodPut:
			// Updating friends requires admin role or ownership
			handler.authMiddleware.RequireAuth(handler.UpdateFriend)(w, r)
		case http.MethodDelete:
			// Deleting friends requires admin role
			handler.authMiddleware.RequireRole("admin")(handler.DeleteFriend)(w, r)
		default:
			writeMethodNotAllowedError(w)
		}
	})

	// Authentication info endpoint
	mux.HandleFunc("/auth/me", handler.authMiddleware.RequireAuth(handler.WhoAmI))

	// Greeting endpoint with optional auth (different greetings for authenticated vs anonymous)
	mux.HandleFunc("/greet", handler.authMiddleware.OptionalAuth(handler.Greet))
}

func writeMethodNotAllowedError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)

	errorResp := map[string]interface{}{
		"error":   "Method Not Allowed",
		"message": "The requested HTTP method is not allowed for this endpoint",
		"code":    http.StatusMethodNotAllowed,
	}

	json.NewEncoder(w).Encode(errorResp)
}
