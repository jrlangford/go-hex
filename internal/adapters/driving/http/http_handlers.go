package http

import (
	"encoding/json"
	"go_hex/internal/adapters/driving/http/middleware"
	"go_hex/internal/core/domain/friendship"
	"go_hex/internal/core/ports/primary"
	"net/http"

	"github.com/google/uuid"
)

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

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status  string            `json:"status"`
	Service string            `json:"service"`
	Checks  map[string]string `json:"checks"`
}

func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResp := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}

	json.NewEncoder(w).Encode(errorResp)
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	successResp := SuccessResponse{
		Status: "success",
		Data:   data,
	}

	json.NewEncoder(w).Encode(successResp)
}

type Handler struct {
	service        primary.Greeter
	healthChecker  primary.HealthChecker
	authMiddleware *middleware.AuthMiddleware
}

func NewHandler(service primary.Greeter, healthChecker primary.HealthChecker, authMiddleware *middleware.AuthMiddleware) *Handler {
	return &Handler{
		service:        service,
		healthChecker:  healthChecker,
		authMiddleware: authMiddleware,
	}
}

func (h *Handler) AddFriend(w http.ResponseWriter, r *http.Request) {
	var friendDTO FriendDTO
	if err := json.NewDecoder(r.Body).Decode(&friendDTO); err != nil {
		writeErrorResponse(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	friendData, err := friendDTO.ToDomain()
	if err != nil {
		writeErrorResponse(w, "Invalid friend data: "+err.Error(), http.StatusBadRequest)
		return
	}

	friend, err := h.service.AddFriend(r.Context(), friendData)
	if err != nil {
		writeErrorResponse(w, "Failed to add friend: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseDTO := NewFriendDTOFromDomain(friend)
	writeSuccessResponse(w, responseDTO, http.StatusCreated)
}

func (h *Handler) Greet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, "Invalid friend ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	greeting, err := h.service.Greet(r.Context(), friendship.NewFriendID(uuid))
	if err != nil {
		writeErrorResponse(w, "Friend not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Check if request is authenticated for enhanced greeting
	claims := middleware.GetTokenClaims(r.Context())
	response := map[string]interface{}{
		"greeting": greeting,
	}

	if claims != nil {
		response["authenticated"] = true
		response["greeted_by"] = claims.Username

		// Add personalized message if greeting own profile
		if claims.UserID == id {
			response["personal_note"] = "This is your own profile!"
		}
	} else {
		response["authenticated"] = false
	}

	writeSuccessResponse(w, response, http.StatusOK)
}

func (h *Handler) GetAllFriends(w http.ResponseWriter, r *http.Request) {
	friends, err := h.service.GetAllFriends(r.Context())
	if err != nil {
		writeErrorResponse(w, "Failed to retrieve friends: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert domain friends to DTOs
	friendDTOs := make([]FriendDTO, 0, len(friends))
	for _, friend := range friends {
		friendDTOs = append(friendDTOs, NewFriendDTOFromDomain(friend))
	}

	writeSuccessResponse(w, friendDTOs, http.StatusOK)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	healthData, err := h.healthChecker.CheckHealth(r.Context())
	if err != nil {
		writeErrorResponse(w, "Health check failed: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Determine overall status
	status := healthData["status"]
	code := http.StatusOK
	if status != "healthy" {
		code = http.StatusServiceUnavailable
	}

	// Remove status from checks to avoid duplication
	checks := make(map[string]string)
	for k, v := range healthData {
		if k != "status" && k != "service" {
			checks[k] = v
		}
	}

	healthResp := HealthResponse{
		Status:  status,
		Service: healthData["service"],
		Checks:  checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(healthResp)
}

func (h *Handler) GetFriend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, "Friend ID is required", http.StatusBadRequest)
		return
	}

	friendUUID, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, "Invalid friend ID format: "+err.Error(), http.StatusBadRequest)
		return
	}

	friend, err := h.service.GetFriend(r.Context(), friendship.NewFriendID(friendUUID))
	if err != nil {
		writeErrorResponse(w, "Friend not found: "+err.Error(), http.StatusNotFound)
		return
	}

	responseDTO := NewFriendDTOFromDomain(friend)
	writeSuccessResponse(w, responseDTO, http.StatusOK)
}

func (h *Handler) UpdateFriend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, "Friend ID is required", http.StatusBadRequest)
		return
	}

	friendUUID, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, "Invalid friend ID format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user can update this resource (admin or owner)
	if !h.isOwnerOrAdmin(r, id) {
		writeErrorResponse(w, "Insufficient permissions to update this friend", http.StatusForbidden)
		return
	}

	var friendDTO FriendDTO
	if err := json.NewDecoder(r.Body).Decode(&friendDTO); err != nil {
		writeErrorResponse(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	friendData, err := friendDTO.ToDomain()
	if err != nil {
		writeErrorResponse(w, "Invalid friend data: "+err.Error(), http.StatusBadRequest)
		return
	}

	friend, err := h.service.UpdateFriend(r.Context(), friendship.NewFriendID(friendUUID), friendData)
	if err != nil {
		writeErrorResponse(w, "Failed to update friend: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responseDTO := NewFriendDTOFromDomain(friend)
	writeSuccessResponse(w, responseDTO, http.StatusOK)
}

func (h *Handler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeErrorResponse(w, "Friend ID is required", http.StatusBadRequest)
		return
	}

	friendUUID, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, "Invalid friend ID format: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.DeleteFriend(r.Context(), friendship.NewFriendID(friendUUID))
	if err != nil {
		writeErrorResponse(w, "Failed to delete friend: "+err.Error(), http.StatusInternalServerError)
		return
	}

	successResp := map[string]string{"message": "Friend deleted successfully"}
	writeSuccessResponse(w, successResp, http.StatusOK)
}

// isOwnerOrAdmin checks if the authenticated user is the owner of the resource or has admin role.
func (h *Handler) isOwnerOrAdmin(r *http.Request, userID string) bool {
	claims := middleware.GetTokenClaims(r.Context())
	if claims == nil {
		return false
	}

	// Check if user is admin
	if claims.IsAuthorized("admin") {
		return true
	}

	// Check if user is the owner
	return claims.UserID == userID
}

// WhoAmI returns information about the currently authenticated user.
func (h *Handler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetTokenClaims(r.Context())
	if claims == nil {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	userInfo := map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"email":    claims.Email,
		"roles":    claims.Roles,
		"metadata": claims.Metadata,
	}

	writeSuccessResponse(w, userInfo, http.StatusOK)
}
