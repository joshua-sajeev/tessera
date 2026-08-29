// Package handler provides HTTP request handlers for Tessera.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/adapters/http/middleware"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// UserHandler handles HTTP request operations for users.
type UserHandler struct {
	service ports.UserService
}

// NewUserHandler creates and returns a new UserHandler.
func NewUserHandler(service ports.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// CreateUserRequest represents the request body to create a user.
type CreateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	StorageQuota int64  `json:"storage_quota,omitempty"`
}

// CreateUserResponse represents the response body returned after creating a user.
type CreateUserResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	APIKey       string    `json:"api_key"`
	StorageQuota int64     `json:"storage_quota"`
	StorageUsed  int64     `json:"storage_used"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserResponse represents the response body returned when retrieving a user.
type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	StorageQuota int64     `json:"storage_quota"`
	StorageUsed  int64     `json:"storage_used"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateStatusRequest represents the request body to update a user's status.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// ErrorResponse represents a standard error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Create handles the HTTP request to create a new user.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Call service with cleaned input
	dto, err := h.service.Create(r.Context(), ports.CreateUserInput{
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.TrimSpace(req.Email),
		StorageQuota: req.StorageQuota,
	})
	if err != nil {
		// Map domain/service errors to HTTP responses
		if strings.Contains(err.Error(), "username and email are required") {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.Contains(err.Error(), "users_username_key") || (strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "username")) {
			respondError(w, http.StatusConflict, "username already exists")
			return
		}
		if strings.Contains(err.Error(), "users_email_key") || (strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email")) {
			respondError(w, http.StatusConflict, "email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, CreateUserResponse{
		ID:           dto.ID,
		Username:     dto.Username,
		Email:        dto.Email,
		APIKey:       dto.APIKey,
		StorageQuota: dto.StorageQuota,
		StorageUsed:  dto.StorageUsed,
		Status:       dto.Status,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	})
}

// Get handles the HTTP request to retrieve a user by ID.
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "missing user id")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id format")
		return
	}

	authUser := middleware.UserFromContext(r.Context())
	if authUser == nil || authUser.ID != id {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	dto, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		ID:           dto.ID,
		Username:     dto.Username,
		Email:        dto.Email,
		StorageQuota: dto.StorageQuota,
		StorageUsed:  dto.StorageUsed,
		Status:       dto.Status,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	})
}

// UpdateStatus handles the HTTP request to update a user's status.
func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "missing user id")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id format")
		return
	}

	authUser := middleware.UserFromContext(r.Context())
	if authUser == nil || authUser.ID != id {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, strings.ToLower(strings.TrimSpace(req.Status))); err != nil {
		// Check for specific error types
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		// Service validation errors (invalid status)
		if strings.Contains(err.Error(), "invalid status") {
			respondError(w, http.StatusBadRequest, "invalid status value")
			return
		}
		// Generic repository/database errors
		respondError(w, http.StatusInternalServerError, "failed to update user status")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
