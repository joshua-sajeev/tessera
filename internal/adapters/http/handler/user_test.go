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
	"github.com/joshua-sajeev/tessera/internal/adapters/http/middleware"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

// Mock service for handler tests
type mockUserService struct {
	createFn       func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error)
	getFn          func(ctx context.Context, id uuid.UUID) (*ports.UserDTO, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status string) error
}

func (m *mockUserService) Create(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockUserService) Get(ctx context.Context, id uuid.UUID) (*ports.UserDTO, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		requestBody        string
		setupService       func() ports.UserService
		expectedStatus     int
		expectedError      string
		expectedUsername   string
		expectedEmail      string
		expectedQuota      int64
		validateAPIKey     bool
		validateTimestamps bool
	}{
		{
			name:        "success with default quota",
			method:      http.MethodPost,
			requestBody: `{"username": "caveman", "email": "caveman@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						now := time.Now().UTC()
						return &ports.UserDTO{
							ID:           uuid.New(),
							Username:     input.Username,
							Email:        input.Email,
							APIKey:       "tsr_v1_abcdef123456",
							StorageQuota: 10737418240,
							StorageUsed:  0,
							Status:       string(user.Active),
							CreatedAt:    now,
							UpdatedAt:    now,
						}, nil
					},
				}
			},
			expectedStatus:     http.StatusCreated,
			expectedUsername:   "caveman",
			expectedEmail:      "caveman@tessera.io",
			expectedQuota:      10737418240,
			validateAPIKey:     true,
			validateTimestamps: true,
		},
		{
			name:        "success with custom quota",
			method:      http.MethodPost,
			requestBody: `{"username": "explorer", "email": "explorer@tessera.io", "storage_quota": 5368709120}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						now := time.Now().UTC()
						return &ports.UserDTO{
							ID:           uuid.New(),
							Username:     input.Username,
							Email:        input.Email,
							APIKey:       "tsr_v1_xyz789",
							StorageQuota: input.StorageQuota,
							StorageUsed:  0,
							Status:       string(user.Active),
							CreatedAt:    now,
							UpdatedAt:    now,
						}, nil
					},
				}
			},
			expectedStatus:   http.StatusCreated,
			expectedUsername: "explorer",
			expectedEmail:    "explorer@tessera.io",
			expectedQuota:    5368709120,
			validateAPIKey:   true,
		},
		{
			name:        "method not allowed",
			method:      http.MethodGet,
			requestBody: `{"username": "test", "email": "test@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
		{
			name:        "invalid json",
			method:      http.MethodPost,
			requestBody: "invalid json",
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:        "missing username",
			method:      http.MethodPost,
			requestBody: `{"email": "test@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("username and email are required")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:        "missing email",
			method:      http.MethodPost,
			requestBody: `{"username": "sailor"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("username and email are required")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:        "empty username after trim",
			method:      http.MethodPost,
			requestBody: `{"username": "   ", "email": "test@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("username and email are required")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:        "empty email after trim",
			method:      http.MethodPost,
			requestBody: `{"username": "test", "email": "   "}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("username and email are required")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "username and email are required",
		},
		{
			name:        "whitespace trimmed",
			method:      http.MethodPost,
			requestBody: `{"username": "  sailor  ", "email": "  sailor@tessera.io  "}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						now := time.Now().UTC()
						return &ports.UserDTO{
							ID:           uuid.New(),
							Username:     input.Username,
							Email:        input.Email,
							APIKey:       "tsr_v1_test",
							StorageQuota: 10737418240,
							StorageUsed:  0,
							Status:       string(user.Active),
							CreatedAt:    now,
							UpdatedAt:    now,
						}, nil
					},
				}
			},
			expectedStatus:   http.StatusCreated,
			expectedUsername: "sailor",
			expectedEmail:    "sailor@tessera.io",
			validateAPIKey:   true,
		},
		{
			name:        "duplicate username",
			method:      http.MethodPost,
			requestBody: `{"username": "knight", "email": "knight2@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("users_username_key")
					},
				}
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "username already exists",
		},
		{
			name:        "duplicate email",
			method:      http.MethodPost,
			requestBody: `{"username": "pirate2", "email": "pirate@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("users_email_key")
					},
				}
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "email already exists",
		},
		{
			name:        "service error",
			method:      http.MethodPost,
			requestBody: `{"username": "merchant", "email": "merchant@tessera.io"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to create user",
		},
		{
			name:        "zero quota defaults to 10GB",
			method:      http.MethodPost,
			requestBody: `{"username": "nomad", "email": "nomad@tessera.io", "storage_quota": 0}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						now := time.Now().UTC()
						return &ports.UserDTO{
							ID:           uuid.New(),
							Username:     input.Username,
							Email:        input.Email,
							APIKey:       "tsr_v1_test",
							StorageQuota: 10737418240,
							StorageUsed:  0,
							Status:       string(user.Active),
							CreatedAt:    now,
							UpdatedAt:    now,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
			expectedQuota:  10737418240,
		},
		{
			name:        "negative quota defaults to 10GB",
			method:      http.MethodPost,
			requestBody: `{"username": "wanderer", "email": "wanderer@tessera.io", "storage_quota": -1000}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					createFn: func(ctx context.Context, input ports.CreateUserInput) (*ports.UserDTO, error) {
						now := time.Now().UTC()
						return &ports.UserDTO{
							ID:           uuid.New(),
							Username:     input.Username,
							Email:        input.Email,
							APIKey:       "tsr_v1_test",
							StorageQuota: 10737418240,
							StorageUsed:  0,
							Status:       string(user.Active),
							CreatedAt:    now,
							UpdatedAt:    now,
						}, nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
			expectedQuota:  10737418240,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			h := handler.NewUserHandler(service)

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
					if resp.CreatedAt.IsZero() || resp.UpdatedAt.IsZero() {
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
	existingUserDTO := &ports.UserDTO{
		ID:           userID,
		Username:     "hunter",
		Email:        "hunter@tessera.io",
		StorageQuota: 5000,
		StorageUsed:  1500,
		Status:       string(user.Active),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tests := []struct {
		name              string
		method            string
		pathID            string
		setupService      func() ports.UserService
		expectedStatus    int
		expectedError     string
		validate          func(t *testing.T, resp *handler.UserResponse)
		bypassContext     bool
		mismatchedContext bool
	}{
		{
			name:   "success",
			method: http.MethodGet,
			pathID: userID.String(),
			setupService: func() ports.UserService {
				return &mockUserService{
					getFn: func(ctx context.Context, id uuid.UUID) (*ports.UserDTO, error) {
						return existingUserDTO, nil
					},
				}
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
			name:   "mismatched user context forbidden",
			method: http.MethodGet,
			pathID: userID.String(),
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus:    http.StatusForbidden,
			expectedError:     "forbidden",
			mismatchedContext: true,
		},
		{
			name:   "missing user context forbidden",
			method: http.MethodGet,
			pathID: userID.String(),
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			bypassContext:  true,
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			pathID: userID.String(),
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
		{
			name:   "missing id",
			method: http.MethodGet,
			pathID: "",
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing user id",
		},
		{
			name:   "invalid uuid format",
			method: http.MethodGet,
			pathID: "not-a-uuid",
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user id format",
		},
		{
			name:   "user not found",
			method: http.MethodGet,
			pathID: uuid.New().String(),
			setupService: func() ports.UserService {
				return &mockUserService{
					getFn: func(ctx context.Context, id uuid.UUID) (*ports.UserDTO, error) {
						return nil, user.ErrUserNotFound
					},
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
		{
			name:   "service error",
			method: http.MethodGet,
			pathID: userID.String(),
			setupService: func() ports.UserService {
				return &mockUserService{
					getFn: func(ctx context.Context, id uuid.UUID) (*ports.UserDTO, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get user",
		},
		{
			name:   "invalid uuid - malformed",
			method: http.MethodGet,
			pathID: "550e8400-e29b-41d4-a716",
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user id format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			h := handler.NewUserHandler(service)

			req := httptest.NewRequest(tt.method, "/users/"+tt.pathID, nil)
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
				if !tt.bypassContext {
					var contextUID uuid.UUID
					if tt.mismatchedContext {
						contextUID = uuid.New()
					} else {
						if uID, err := uuid.Parse(tt.pathID); err == nil {
							contextUID = uID
						}
					}
					if contextUID != uuid.Nil {
						req = req.WithContext(middleware.WithUser(req.Context(), &user.User{ID: contextUID}))
					}
				}
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
		name              string
		method            string
		pathID            string
		requestBody       string
		setupService      func() ports.UserService
		expectedStatus    int
		expectedError     string
		bypassContext     bool
		mismatchedContext bool
	}{
		{
			name:        "suspend user",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "mismatched user context forbidden",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus:    http.StatusForbidden,
			expectedError:     "forbidden",
			mismatchedContext: true,
		},
		{
			name:        "missing user context forbidden",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "forbidden",
			bypassContext:  true,
		},
		{
			name:        "delete user",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "deleted"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "reactivate user",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "active"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "put method allowed",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "patch method allowed",
			method:      http.MethodPatch,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "delete method not allowed",
			method:      http.MethodDelete,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "method not allowed",
		},
		{
			name:        "missing id",
			method:      http.MethodPut,
			pathID:      "",
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing user id",
		},
		{
			name:        "invalid uuid",
			method:      http.MethodPut,
			pathID:      "bad-id",
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid user id format",
		},
		{
			name:        "invalid json",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: "not json",
			setupService: func() ports.UserService {
				return &mockUserService{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:        "invalid status",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "extinct"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return errors.New("invalid status: extinct")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid status value",
		},
		{
			name:        "status uppercase converted to lowercase",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "SUSPENDED"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						if status != "suspended" {
							t.Errorf("expected lowercase 'suspended', got %q", status)
						}
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "status with whitespace trimmed",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "  suspended  "}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						if status != "suspended" {
							t.Errorf("expected trimmed 'suspended', got %q", status)
						}
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "user not found",
			method:      http.MethodPut,
			pathID:      uuid.New().String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return user.ErrUserNotFound
					},
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "user not found",
		},
		{
			name:        "service error",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "suspended"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to update user status",
		},
		{
			name:        "empty status",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": ""}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return errors.New("invalid status: ")
					},
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid status value",
		},
		{
			name:        "status mixed case",
			method:      http.MethodPut,
			pathID:      userID.String(),
			requestBody: `{"status": "SuSpEnDeD"}`,
			setupService: func() ports.UserService {
				return &mockUserService{
					updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
						return nil
					},
				}
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()
			h := handler.NewUserHandler(service)

			req := httptest.NewRequest(tt.method, "/users/"+tt.pathID+"/status", bytes.NewBufferString(tt.requestBody))
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
				if !tt.bypassContext {
					var contextUID uuid.UUID
					if tt.mismatchedContext {
						contextUID = uuid.New()
					} else {
						if uID, err := uuid.Parse(tt.pathID); err == nil {
							contextUID = uID
						}
					}
					if contextUID != uuid.Nil {
						req = req.WithContext(middleware.WithUser(req.Context(), &user.User{ID: contextUID}))
					}
				}
			}
			rec := httptest.NewRecorder()

			h.UpdateStatus(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedStatus != http.StatusNoContent && tt.expectedError != "" {
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
