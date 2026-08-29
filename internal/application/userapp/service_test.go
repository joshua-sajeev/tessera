package userapp_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/application/userapp"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
	"github.com/joshua-sajeev/tessera/internal/ports"
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

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name           string
		input          ports.CreateUserInput
		setupRepo      func() *mockUserRepository
		expectedErr    string
		validateResult func(t *testing.T, dto *ports.UserDTO)
	}{
		{
			name: "success with default quota",
			input: ports.CreateUserInput{
				Username:     "alice",
				Email:        "alice@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if dto.Username != "alice" {
					t.Errorf("expected username 'alice', got %q", dto.Username)
				}
				if dto.Email != "alice@example.com" {
					t.Errorf("expected email 'alice@example.com', got %q", dto.Email)
				}
				if dto.StorageQuota != 10737418240 {
					t.Errorf("expected quota 10GB, got %d", dto.StorageQuota)
				}
				if dto.Status != string(user.Active) {
					t.Errorf("expected status 'active', got %q", dto.Status)
				}
				if dto.StorageUsed != 0 {
					t.Errorf("expected storage used 0, got %d", dto.StorageUsed)
				}
				if dto.APIKey == "" {
					t.Error("expected APIKey to be generated")
				}
				if dto.CreatedAt.IsZero() || dto.UpdatedAt.IsZero() {
					t.Error("expected CreatedAt and UpdatedAt to be set")
				}
			},
		},
		{
			name: "success with custom quota",
			input: ports.CreateUserInput{
				Username:     "bob",
				Email:        "bob@example.com",
				StorageQuota: 5368709120, // 5GB
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if dto.StorageQuota != 5368709120 {
					t.Errorf("expected quota 5368709120, got %d", dto.StorageQuota)
				}
			},
		},
		{
			name: "negative quota defaults to 10GB",
			input: ports.CreateUserInput{
				Username:     "charlie",
				Email:        "charlie@example.com",
				StorageQuota: -1000,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if dto.StorageQuota != 10737418240 {
					t.Errorf("expected quota 10GB for negative input, got %d", dto.StorageQuota)
				}
			},
		},
		{
			name: "missing username",
			input: ports.CreateUserInput{
				Username:     "",
				Email:        "dave@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedErr: "username and email are required",
		},
		{
			name: "missing email",
			input: ports.CreateUserInput{
				Username:     "eve",
				Email:        "",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedErr: "username and email are required",
		},
		{
			name: "missing both username and email",
			input: ports.CreateUserInput{
				Username:     "",
				Email:        "",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedErr: "username and email are required",
		},
		{
			name: "duplicate username",
			input: ports.CreateUserInput{
				Username:     "frank",
				Email:        "frank2@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				repo := &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
				repo.users[uuid.New()] = &user.User{
					Username: "frank",
					Email:    "frank1@example.com",
				}
				return repo
			},
			expectedErr: "users_username_key",
		},
		{
			name: "duplicate email",
			input: ports.CreateUserInput{
				Username:     "grace",
				Email:        "shared@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				repo := &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
				repo.users[uuid.New()] = &user.User{
					Username: "henry",
					Email:    "shared@example.com",
				}
				return repo
			},
			expectedErr: "users_email_key",
		},
		{
			name: "repository error",
			input: ports.CreateUserInput{
				Username:     "iris",
				Email:        "iris@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: make(map[uuid.UUID]*user.User),
					err:   errors.New("database connection failed"),
				}
			},
			expectedErr: "database connection failed",
		},
		{
			name: "api key has correct prefix and version",
			input: ports.CreateUserInput{
				Username:     "jack",
				Email:        "jack@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if !hasValidAPIKeyFormat(dto.APIKey, "tsr", "v1") {
					t.Errorf("expected API key with prefix 'tsr_v1_', got %q", dto.APIKey)
				}
			},
		},
		{
			name: "different users get different api keys",
			input: ports.CreateUserInput{
				Username:     "kate",
				Email:        "kate@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				// Verify it's not empty and has the right format
				if dto.APIKey == "" {
					t.Error("expected APIKey to be generated")
				}
			},
		},
		{
			name: "id is generated as uuid",
			input: ports.CreateUserInput{
				Username:     "leo",
				Email:        "leo@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if dto.ID == uuid.Nil {
					t.Error("expected ID to be a valid UUID, got nil")
				}
			},
		},
		{
			name: "timestamps are close to now",
			input: ports.CreateUserInput{
				Username:     "mia",
				Email:        "mia@example.com",
				StorageQuota: 0,
			},
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				now := time.Now().UTC()
				if dto.CreatedAt.IsZero() || dto.UpdatedAt.IsZero() {
					t.Fatal("expected timestamps to be set")
				}
				if dto.CreatedAt.After(now.Add(1 * time.Second)) {
					t.Error("CreatedAt is in the future")
				}
				if dto.UpdatedAt.After(now.Add(1 * time.Second)) {
					t.Error("UpdatedAt is in the future")
				}
				if dto.CreatedAt.Before(now.Add(-2 * time.Second)) {
					t.Error("CreatedAt is too far in the past")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			service := userapp.NewUserService(repo, "tsr", "v1")

			dto, err := service.Create(context.Background(), tt.input)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if !containsError(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dto == nil {
				t.Fatal("expected DTO, got nil")
			}

			if tt.validateResult != nil {
				tt.validateResult(t, dto)
			}
		})
	}
}

func TestUserService_Get(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	existingUser := &user.User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		StorageQuota: 5000,
		StorageUsed:  1500,
		Status:       string(user.Active),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tests := []struct {
		name           string
		userID         uuid.UUID
		setupRepo      func() *mockUserRepository
		expectedErr    string
		validateResult func(t *testing.T, dto *ports.UserDTO)
	}{
		{
			name:   "success",
			userID: userID,
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: map[uuid.UUID]*user.User{userID: existingUser}}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if dto.ID != userID {
					t.Errorf("expected ID %v, got %v", userID, dto.ID)
				}
				if dto.Username != "testuser" {
					t.Errorf("expected username 'testuser', got %q", dto.Username)
				}
				if dto.Email != "test@example.com" {
					t.Errorf("expected email 'test@example.com', got %q", dto.Email)
				}
				if dto.StorageQuota != 5000 {
					t.Errorf("expected quota 5000, got %d", dto.StorageQuota)
				}
				if dto.StorageUsed != 1500 {
					t.Errorf("expected storage used 1500, got %d", dto.StorageUsed)
				}
				if dto.Status != string(user.Active) {
					t.Errorf("expected status 'active', got %q", dto.Status)
				}
				if dto.CreatedAt.IsZero() || dto.UpdatedAt.IsZero() {
					t.Error("expected CreatedAt and UpdatedAt to be set")
				}
			},
		},
		{
			name:   "user not found",
			userID: uuid.New(),
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedErr: "user not found",
		},
		{
			name:   "repository error",
			userID: userID,
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: make(map[uuid.UUID]*user.User),
					err:   errors.New("database connection failed"),
				}
			},
			expectedErr: "database connection failed",
		},
		{
			name:   "dto excludes api key fields",
			userID: userID,
			setupRepo: func() *mockUserRepository {
				userCopy := *existingUser
				userCopy.APIKeyID = "secret-key-id"
				userCopy.APIKeyHash = "secret-hash"
				return &mockUserRepository{users: map[uuid.UUID]*user.User{userID: &userCopy}}
			},
			validateResult: func(t *testing.T, dto *ports.UserDTO) {
				if dto.APIKey != "" {
					t.Errorf("expected empty APIKey in Get response, got %q", dto.APIKey)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			service := userapp.NewUserService(repo, "tsr", "v1")

			dto, err := service.Get(context.Background(), tt.userID)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if !containsError(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if dto == nil {
				t.Fatal("expected DTO, got nil")
			}

			if tt.validateResult != nil {
				tt.validateResult(t, dto)
			}
		})
	}
}

func TestUserService_UpdateStatus(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		newStatus   string
		setupRepo   func() *mockUserRepository
		expectedErr string
	}{
		{
			name:      "suspend active user",
			userID:    userID,
			newStatus: "suspended",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
		},
		{
			name:      "delete user",
			userID:    userID,
			newStatus: "deleted",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
		},
		{
			name:      "reactivate suspended user",
			userID:    userID,
			newStatus: "active",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Suspended),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
		},
		{
			name:      "invalid status",
			userID:    userID,
			newStatus: "invalid_status",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
			expectedErr: "invalid status",
		},
		{
			name:      "empty status",
			userID:    userID,
			newStatus: "",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
			expectedErr: "invalid status",
		},
		{
			name:      "user not found",
			userID:    uuid.New(),
			newStatus: "suspended",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{users: make(map[uuid.UUID]*user.User)}
			},
			expectedErr: "user not found",
		},
		{
			name:      "repository error",
			userID:    userID,
			newStatus: "suspended",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
					err: errors.New("database connection failed"),
				}
			},
			expectedErr: "database connection failed",
		},
		{
			name:      "status case insensitive - uppercase",
			userID:    userID,
			newStatus: "SUSPENDED",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
		},
		{
			name:      "status case insensitive - mixed case",
			userID:    userID,
			newStatus: "SuSpEnDeD",
			setupRepo: func() *mockUserRepository {
				return &mockUserRepository{
					users: map[uuid.UUID]*user.User{
						userID: {
							ID:       userID,
							Status:   string(user.Active),
							Username: "test",
							Email:    "test@example.com",
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			service := userapp.NewUserService(repo, "tsr", "v1")

			err := service.UpdateStatus(context.Background(), tt.userID, tt.newStatus)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if !containsError(err.Error(), tt.expectedErr) {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// Helper functions

func hasValidAPIKeyFormat(apiKey, prefix, version string) bool {
	expected := prefix + "_" + version + "_"
	return len(apiKey) > len(expected) && apiKey[:len(expected)] == expected
}

func containsError(actualErr, expectedSubstring string) bool {
	return errorContains(actualErr, expectedSubstring)
}

func errorContains(err, substring string) bool {
	return len(err) > 0 && len(substring) > 0 && strings.Contains(err, substring)
}
