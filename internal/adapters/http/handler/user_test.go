package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/adapters/http/handler"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

type mockUserRepository struct {
	users map[uuid.UUID]*user.User
	err   error
}

func (m *mockUserRepository) Create(ctx context.Context, u *user.User) error {
	if m.err != nil {
		return m.err
	}
	for _, existing := range m.users {
		if existing.Username == u.Username {
			return errors.New("users_username_key")
		}
		if existing.Email == u.Email {
			return errors.New("users_email_key")
		}
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepository) Get(ctx context.Context, id uuid.UUID) (*user.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByAPIKeyID(ctx context.Context, apiKeyID string) (*user.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.UserStatus) error {
	if m.err != nil {
		return m.err
	}
	u, ok := m.users[id]
	if !ok {
		return user.ErrUserNotFound
	}
	u.Status = string(status)
	return nil
}

func (m *mockUserRepository) AddStorageUsed(ctx context.Context, id uuid.UUID, size int64) error {
	return nil
}

func (m *mockUserRepository) SubtractStorageUsed(ctx context.Context, id uuid.UUID, size int64) error {
	return nil
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		requestBody        string
		setupRepo          func() *mockUserRepository
		expectedStatus     int
		expectedError      string
		expectedUsername   string
		expectedEmail      string
		expectedQuota      int64
		validateAPIKey     bool
		validateTimestamps bool
	}{
		{
			name:               "success with default quota",
			method:             http.MethodPost,
			requestBody:        `{"username": "caveman", "email": "caveman@tessera.io"}`,
			setupRepo:          func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus:     http.StatusCreated,
			expectedUsername:   "caveman",
			expectedEmail:      "caveman@tessera.io",
			expectedQuota:      10737418240,
			validateAPIKey:     true,
			validateTimestamps: true,
		},
		{
			name:               "success with custom quota",
			method:             http.MethodPost,
			requestBody:        `{"username": "explorer", "email": "explorer@tessera.io", "storage_quota": 5368709120}`,
			setupRepo:          func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus:     http.StatusCreated,
			expectedUsername:   "explorer",
			expectedEmail:      "explorer@tessera.io",
			expectedQuota:      5368709120,
			validateAPIKey:     true,
			validateTimestamps: true,
		},
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			requestBody:    `{"username": "test", "email": "test@tessera.io"}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
		{
			name:           "invalid json",
			method:         http.MethodPost,
			requestBody:    "invalid json",
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "missing username",
			method:         http.MethodPost,
			requestBody:    `{"email": "test@tessera.io"}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:           "missing email",
			method:         http.MethodPost,
			requestBody:    `{"username": "sailor"}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:           "empty username after trim",
			method:         http.MethodPost,
			requestBody:    `{"username": "   ", "email": "test@tessera.io"}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:           "empty email after trim",
			method:         http.MethodPost,
			requestBody:    `{"username": "test", "email": "   "}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:             "whitespace trimmed",
			method:           http.MethodPost,
			requestBody:      `{"username": "  sailor  ", "email": "  sailor@tessera.io  "}`,
			setupRepo:        func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus:   http.StatusCreated,
			expectedUsername: "sailor",
			expectedEmail:    "sailor@tessera.io",
			validateAPIKey:   true,
		},
		{
			name:        "duplicate username",
			method:      http.MethodPost,
			requestBody: `{"username": "knight", "email": "knight2@tessera.io"}`,
			setupRepo: func() *mockUserRepository {
				repo := &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
				repo.users[uuid.New()] = &user.User{Username: "knight", Email: "knight1@tessera.io"}
				return repo
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "username already exists",
		},
		{
			name:        "duplicate email",
			method:      http.MethodPost,
			requestBody: `{"username": "pirate2", "email": "pirate@tessera.io"}`,
			setupRepo: func() *mockUserRepository {
				repo := &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
				repo.users[uuid.New()] = &user.User{Username: "pirate1", Email: "pirate@tessera.io"}
				return repo
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "email already exists",
		},
		{
			name:        "repository error",
			method:      http.MethodPost,
			requestBody: `{"username": "merchant", "email": "merchant@tessera.io"}`,
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User), err: errors.New("db error")}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create user",
		},
		{
			name:           "zero quota defaults to 10GB",
			method:         http.MethodPost,
			requestBody:    `{"username": "nomad", "email": "nomad@tessera.io", "storage_quota": 0}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusCreated,
			expectedQuota:  10737418240,
		},
		{
			name:           "negative quota defaults to 10GB",
			method:         http.MethodPost,
			requestBody:    `{"username": "wanderer", "email": "wanderer@tessera.io", "storage_quota": -1000}`,
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusCreated,
			expectedQuota:  10737418240,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			h := handler.NewUserHandler(repo, "tsr", "v1")

			req := httptest.NewRequest(tt.method, "/users", bytes.NewBufferString(tt.requestBody))
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp handler.CreateUserResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if tt.expectedUsername != "" && resp.Username != tt.expectedUsername {
					t.Errorf("expected username %q, got %q", tt.expectedUsername, resp.Username)
				}

				if tt.expectedEmail != "" && resp.Email != tt.expectedEmail {
					t.Errorf("expected email %q, got %q", tt.expectedEmail, resp.Email)
				}

				if tt.expectedQuota > 0 && resp.StorageQuota != tt.expectedQuota {
					t.Errorf("expected quota %d, got %d", tt.expectedQuota, resp.StorageQuota)
				}

				if tt.validateAPIKey && !strings.HasPrefix(resp.APIKey, "tsr_v1_") {
					t.Errorf("expected API key prefix 'tsr_v1_', got %q", resp.APIKey)
				}

				if tt.validateTimestamps {
					if resp.CreatedAt == nil || resp.UpdatedAt == nil {
						t.Error("expected CreatedAt and UpdatedAt to be set")
					}
				}

				if resp.Status != string(user.Active) {
					t.Errorf("expected status 'active', got %q", resp.Status)
				}

				if resp.StorageUsed != 0 {
					t.Errorf("expected storage used 0, got %d", resp.StorageUsed)
				}
			} else if tt.expectedError != "" {
				var errResp handler.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if !strings.Contains(errResp.Error, tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, errResp.Error)
				}
			}
		})
	}
}

func TestUserHandler_Get(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	existingUser := &user.User{
		ID:           userID,
		Username:     "hunter",
		Email:        "hunter@tessera.io",
		StorageQuota: 5000,
		StorageUsed:  1500,
		Status:       string(user.Active),
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	tests := []struct {
		name           string
		method         string
		pathID         string
		setupRepo      func() *mockUserRepository
		expectedStatus int
		expectedError  string
		validate       func(t *testing.T, resp *handler.UserResponse)
	}{
		{
			name:   "success",
			method: http.MethodGet,
			pathID: userID.String(),
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: map[uuid.UUID]*user.User{userID: existingUser}}
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, resp *handler.UserResponse) {
				if resp.ID != userID {
					t.Errorf("expected ID %v, got %v", userID, resp.ID)
				}
				if resp.Username != "hunter" {
					t.Errorf("expected username 'hunter', got %q", resp.Username)
				}
				if resp.Email != "hunter@tessera.io" {
					t.Errorf("expected email 'hunter@tessera.io', got %q", resp.Email)
				}
				if resp.StorageQuota != 5000 {
					t.Errorf("expected quota 5000, got %d", resp.StorageQuota)
				}
				if resp.StorageUsed != 1500 {
					t.Errorf("expected storage used 1500, got %d", resp.StorageUsed)
				}
				if resp.Status != string(user.Active) {
					t.Errorf("expected status 'active', got %q", resp.Status)
				}
			},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			pathID:         userID.String(),
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
		{
			name:           "missing id",
			method:         http.MethodGet,
			pathID:         "",
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing user id",
		},
		{
			name:           "invalid uuid format",
			method:         http.MethodGet,
			pathID:         "not-a-uuid",
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user id format",
		},
		{
			name:           "user not found",
			method:         http.MethodGet,
			pathID:         uuid.New().String(),
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
		{
			name:   "repository error",
			method: http.MethodGet,
			pathID: userID.String(),
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User), err: errors.New("db error")}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get user",
		},
		{
			name:           "invalid uuid - malformed",
			method:         http.MethodGet,
			pathID:         "550e8400-e29b-41d4-a716",
			setupRepo:      func() *mockUserRepository { return &mockUserRepository{users: make(map[uuid.UUID]*user.User)} },
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user id format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			h := handler.NewUserHandler(repo, "tsr", "v1")

			req := httptest.NewRequest(tt.method, "/users/"+tt.pathID, nil)
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			rec := httptest.NewRecorder()

			h.Get(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusOK && tt.validate != nil {
				var resp handler.UserResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				tt.validate(t, &resp)
			} else if tt.expectedError != "" {
				var errResp handler.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if !strings.Contains(errResp.Error, tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, errResp.Error)
				}
			}
		})
	}
}

func TestUserHandler_UpdateStatus(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		method         string
		pathID         string
		requestBody    string
		setupRepo      func(id uuid.UUID) *mockUserRepository
		expectedStatus int
		expectedError  string
		validateStatus func(t *testing.T, repo *mockUserRepository, id uuid.UUID)
	}{
		{
			name:           "suspend user",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "suspended"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
			validateStatus: func(t *testing.T, repo *mockUserRepository, id uuid.UUID) {
				if u, ok := repo.users[id]; ok && u.Status != string(user.Suspended) {
					t.Errorf("expected status 'suspended', got %q", u.Status)
				}
			},
		},
		{
			name:           "delete user",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "deleted"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
			validateStatus: func(t *testing.T, repo *mockUserRepository, id uuid.UUID) {
				if u, ok := repo.users[id]; ok && u.Status != string(user.Deleted) {
					t.Errorf("expected status 'deleted', got %q", u.Status)
				}
			},
		},
		{
			name:        "reactivate user",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "active"}`,
			setupRepo: func(id uuid.UUID) *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						id: {ID: id, Username: "test", Email: "test@tessera.io", Status: string(user.Suspended)},
					},
				}
			},
			expectedStatus: http.StatusNoContent,
			validateStatus: func(t *testing.T, repo *mockUserRepository, id uuid.UUID) {
				if u, ok := repo.users[id]; ok && u.Status != string(user.Active) {
					t.Errorf("expected status 'active', got %q", u.Status)
				}
			},
		},
		{
			name:           "put method allowed",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "suspended"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "patch method allowed",
			method:         http.MethodPatch,
			pathID:         userID.String(),
			requestBody:    `{"status": "suspended"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete method not allowed",
			method:         http.MethodDelete,
			pathID:         userID.String(),
			requestBody:    `{"status": "suspended"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
		{
			name:        "missing id",
			method:      http.MethodPut,
			pathID:      "",
			requestBody: `{"status": "suspended"}`,
			setupRepo: func(id uuid.UUID) *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing user id",
		},
		{
			name:        "invalid uuid",
			method:      http.MethodPut,
			pathID:      "bad-id",
			requestBody: `{"status": "suspended"}`,
			setupRepo: func(id uuid.UUID) *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user id format",
		},
		{
			name:           "invalid json",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    "not json",
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "invalid status",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "extinct"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid status value",
		},
		{
			name:           "status uppercase converted to lowercase",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "SUSPENDED"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
			validateStatus: func(t *testing.T, repo *mockUserRepository, id uuid.UUID) {
				if u, ok := repo.users[id]; ok && u.Status != string(user.Suspended) {
					t.Errorf("expected status 'suspended' from uppercase input, got %q", u.Status)
				}
			},
		},
		{
			name:           "status with whitespace trimmed",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "  suspended  "}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
			validateStatus: func(t *testing.T, repo *mockUserRepository, id uuid.UUID) {
				if u, ok := repo.users[id]; ok && u.Status != string(user.Suspended) {
					t.Errorf("expected status 'suspended' after trimming, got %q", u.Status)
				}
			},
		},
		{
			name:        "user not found",
			method:      http.MethodPut,
			pathID:      uuid.New().String(),
			requestBody: `{"status": "suspended"}`,
			setupRepo: func(id uuid.UUID) *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
		{
			name:        "repository error",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupRepo: func(id uuid.UUID) *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{id: setupActiveUser(id).users[id]},
					err:   errors.New("db error"),
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update user status",
		},
		{
			name:           "empty status",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": ""}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid status value",
		},
		{
			name:           "status mixed case",
			method:         http.MethodPut,
			pathID:         userID.String(),
			requestBody:    `{"status": "SuSpEnDeD"}`,
			setupRepo:      setupActiveUser,
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(userID)
			h := handler.NewUserHandler(repo, "tsr", "v1")

			req := httptest.NewRequest(tt.method, "/users/"+tt.pathID+"/status", bytes.NewBufferString(tt.requestBody))
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			rec := httptest.NewRecorder()

			h.UpdateStatus(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus == http.StatusNoContent && tt.validateStatus != nil {
				tt.validateStatus(t, repo, userID)
			} else if tt.expectedError != "" {
				var errResp handler.ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if !strings.Contains(errResp.Error, tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, errResp.Error)
				}
			}
		})
	}
}

func setupActiveUser(id uuid.UUID) *mockUserRepository {
	return &mockUserRepository{
		users: map[uuid.UUID]*user.User{
			id: {
				ID:           id,
				Username:     "testuser",
				Email:        "testuser@tessera.io",
				StorageQuota: 5000,
				Status:       string(user.Active),
			},
		},
	}
}
