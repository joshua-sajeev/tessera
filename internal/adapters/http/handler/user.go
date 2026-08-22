// Package handler provides HTTP request handlers for Tessera.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

type UserHandler struct {
	repo          ports.UserRepository
	apiKeyPrefix  string
	apiKeyVersion string
}

func NewUserHandler(repo ports.UserRepository, apiKeyPrefix, apiKeyVersion string) *UserHandler {
	return &UserHandler{
		repo:          repo,
		apiKeyPrefix:  apiKeyPrefix,
		apiKeyVersion: apiKeyVersion,
	}
}

type CreateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	StorageQuota int64  `json:"storage_quota,omitempty"`
}

type CreateUserResponse struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	APIKey       string     `json:"api_key"`
	StorageQuota int64      `json:"storage_quota"`
	StorageUsed  int64      `json:"storage_used"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type UserResponse struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	StorageQuota int64      `json:"storage_quota"`
	StorageUsed  int64      `json:"storage_used"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

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

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" || req.Email == "" {
		respondError(w, http.StatusBadRequest, "username and email are required")
		return
	}

	generator := user.NewAPIKeyGenerator(h.apiKeyPrefix, h.apiKeyVersion)
	fullKey, keyID, err := generator.Generate()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate api key")
		return
	}

	hasher := user.NewKeyHasher()
	hash, err := hasher.Hash(fullKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash api key")
		return
	}

	quota := req.StorageQuota
	if quota <= 0 {
		quota = 10737418240 // 10GB default
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		APIKeyID:     keyID,
		APIKeyHash:   hash,
		StorageQuota: quota,
		StorageUsed:  0,
		Status:       string(user.Active),
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	if err := h.repo.Create(r.Context(), u); err != nil {
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
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		APIKey:       fullKey,
		StorageQuota: u.StorageQuota,
		StorageUsed:  u.StorageUsed,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	})
}

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

	u, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		StorageQuota: u.StorageQuota,
		StorageUsed:  u.StorageUsed,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	})
}

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

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	status := user.UserStatus(strings.ToLower(strings.TrimSpace(req.Status)))
	if status != user.Active && status != user.Suspended && status != user.Deleted {
		respondError(w, http.StatusBadRequest, "invalid status value")
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, status); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
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
