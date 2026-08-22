package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

type userBuilder struct {
	user *user.User
}

func newUser() *userBuilder {
	now := time.Now().UTC().Truncate(time.Microsecond)
	id := uuid.New()
	return &userBuilder{
		user: &user.User{
			ID:           id,
			Username:     "testuser-" + id.String()[:8],
			Email:        "user-" + id.String()[:8] + "@test.com",
			APIKeyID:     "api-key-id-" + id.String()[:8],
			APIKeyHash:   "api-key-hash-" + id.String()[:8],
			StorageQuota: 10000,
			StorageUsed:  0,
			Status:       string(user.Active),
			CreatedAt:    &now,
			UpdatedAt:    &now,
		},
	}
}

func (b *userBuilder) username(u string) *userBuilder {
	b.user.Username = u
	return b
}

func (b *userBuilder) email(e string) *userBuilder {
	b.user.Email = e
	return b
}

func (b *userBuilder) apiKeyID(id string) *userBuilder {
	b.user.APIKeyID = id
	return b
}

func (b *userBuilder) storageQuota(q int64) *userBuilder {
	b.user.StorageQuota = q
	return b
}

func (b *userBuilder) storageUsed(u int64) *userBuilder {
	b.user.StorageUsed = u
	return b
}

func (b *userBuilder) build() *user.User {
	return b.user
}

func setupUserTest(t *testing.T) (*postgres.UserRepository, context.Context) {
	t.Helper()
	cleanDB(t)
	return postgres.NewUserRepository(db), context.Background()
}

func mustCreateUser(t *testing.T, repo *postgres.UserRepository, ctx context.Context, u *user.User) {
	t.Helper()
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create(%s) failed: %v", u.Username, err)
	}
}

func mustGetUser(t *testing.T, repo *postgres.UserRepository, ctx context.Context, id uuid.UUID) *user.User {
	t.Helper()
	got, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get(%v) failed: %v", id, err)
	}
	return got
}

func assertUserEqual(t *testing.T, got, want *user.User) {
	t.Helper()
	if got.ID != want.ID {
		t.Errorf("ID: got %v, want %v", got.ID, want.ID)
	}
	if got.Username != want.Username {
		t.Errorf("Username: got %q, want %q", got.Username, want.Username)
	}
	if got.Email != want.Email {
		t.Errorf("Email: got %q, want %q", got.Email, want.Email)
	}
	if got.APIKeyID != want.APIKeyID {
		t.Errorf("APIKeyID: got %q, want %q", got.APIKeyID, want.APIKeyID)
	}
	if got.APIKeyHash != want.APIKeyHash {
		t.Errorf("APIKeyHash: got %q, want %q", got.APIKeyHash, want.APIKeyHash)
	}
	if got.StorageQuota != want.StorageQuota {
		t.Errorf("StorageQuota: got %d, want %d", got.StorageQuota, want.StorageQuota)
	}
	if got.StorageUsed != want.StorageUsed {
		t.Errorf("StorageUsed: got %d, want %d", got.StorageUsed, want.StorageUsed)
	}
	if got.Status != want.Status {
		t.Errorf("Status: got %q, want %q", got.Status, want.Status)
	}
	if (got.CreatedAt == nil) != (want.CreatedAt == nil) {
		t.Errorf("CreatedAt nil mismatch: got %v, want %v", got.CreatedAt, want.CreatedAt)
	} else if got.CreatedAt != nil && !got.CreatedAt.Equal(*want.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", *got.CreatedAt, *want.CreatedAt)
	}
	if (got.UpdatedAt == nil) != (want.UpdatedAt == nil) {
		t.Errorf("UpdatedAt nil mismatch: got %v, want %v", got.UpdatedAt, want.UpdatedAt)
	} else if got.UpdatedAt != nil && !got.UpdatedAt.Equal(*want.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", *got.UpdatedAt, *want.UpdatedAt)
	}
}

func TestUserRepository_Create(t *testing.T) {
	repo, ctx := setupUserTest(t)

	want := newUser().build()
	mustCreateUser(t, repo, ctx, want)

	got := mustGetUser(t, repo, ctx, want.ID)
	assertUserEqual(t, got, want)
}

func TestUserRepository_Create_UniqueConstraints(t *testing.T) {
	tests := []struct {
		name      string
		setupUser func(base *user.User) *user.User
	}{
		{
			name: "duplicate username",
			setupUser: func(base *user.User) *user.User {
				return newUser().username(base.Username).build()
			},
		},
		{
			name: "duplicate email",
			setupUser: func(base *user.User) *user.User {
				return newUser().email(base.Email).build()
			},
		},
		{
			name: "duplicate api key id",
			setupUser: func(base *user.User) *user.User {
				return newUser().apiKeyID(base.APIKeyID).build()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)

			base := newUser().build()
			mustCreateUser(t, repo, ctx, base)

			dup := tt.setupUser(base)
			err := repo.Create(ctx, dup)
			if err == nil {
				t.Error("expected error due to unique constraint violation, got nil")
			}
		})
	}
}

func TestUserRepository_Get(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) uuid.UUID
		wantError error
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) uuid.UUID {
				u := newUser().build()
				mustCreateUser(t, repo, ctx, u)
				return u.ID
			},
			wantError: nil,
		},
		{
			name: "not found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) uuid.UUID {
				return uuid.New()
			},
			wantError: user.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)
			userID := tt.setup(t, repo, ctx)

			got, err := repo.Get(ctx, userID)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("Get() error: got %v, want %v", err, tt.wantError)
			}

			if tt.wantError == nil && got == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestUserRepository_GetByAPIKeyID(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string
		wantError error
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string {
				u := newUser().build()
				mustCreateUser(t, repo, ctx, u)
				return u.APIKeyID
			},
			wantError: nil,
		},
		{
			name: "not found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string {
				return "nonexistent-key"
			},
			wantError: user.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)
			apiKeyID := tt.setup(t, repo, ctx)

			got, err := repo.GetByAPIKeyID(ctx, apiKeyID)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("GetByAPIKeyID() error: got %v, want %v", err, tt.wantError)
			}

			if tt.wantError == nil && got == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string
		wantError error
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string {
				u := newUser().build()
				mustCreateUser(t, repo, ctx, u)
				return u.Email
			},
			wantError: nil,
		},
		{
			name: "not found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string {
				return "nonexistent@test.com"
			},
			wantError: user.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)
			email := tt.setup(t, repo, ctx)

			got, err := repo.GetByEmail(ctx, email)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("GetByEmail() error: got %v, want %v", err, tt.wantError)
			}

			if tt.wantError == nil && got == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string
		wantError error
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string {
				u := newUser().build()
				mustCreateUser(t, repo, ctx, u)
				return u.Username
			},
			wantError: nil,
		},
		{
			name: "not found",
			setup: func(t *testing.T, repo *postgres.UserRepository, ctx context.Context) string {
				return "nonexistentuser"
			},
			wantError: user.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)
			username := tt.setup(t, repo, ctx)

			got, err := repo.GetByUsername(ctx, username)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("GetByUsername() error: got %v, want %v", err, tt.wantError)
			}

			if tt.wantError == nil && got == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	tests := []struct {
		name       string
		newStatus  user.UserStatus
		userExists bool
		wantError  error
	}{
		{
			name:       "success - active to suspended",
			newStatus:  user.Suspended,
			userExists: true,
			wantError:  nil,
		},
		{
			name:       "success - active to deleted",
			newStatus:  user.Deleted,
			userExists: true,
			wantError:  nil,
		},
		{
			name:       "user not found",
			newStatus:  user.Suspended,
			userExists: false,
			wantError:  user.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)

			var userID uuid.UUID
			if tt.userExists {
				u := newUser().build()
				mustCreateUser(t, repo, ctx, u)
				userID = u.ID
			} else {
				userID = uuid.New()
			}

			err := repo.UpdateStatus(ctx, userID, tt.newStatus)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("UpdateStatus() error: got %v, want %v", err, tt.wantError)
				return
			}

			if tt.userExists && tt.wantError == nil {
				got := mustGetUser(t, repo, ctx, userID)
				if got.Status != string(tt.newStatus) {
					t.Errorf("Status: got %q, want %q", got.Status, tt.newStatus)
				}
			}
		})
	}
}

func TestUserRepository_AddStorageUsed(t *testing.T) {
	tests := []struct {
		name         string
		initialQuota int64
		initialUsed  int64
		addSize      int64
		userExists   bool
		wantError    error
		expectedUsed int64
	}{
		{
			name:         "success - within quota",
			initialQuota: 100,
			initialUsed:  20,
			addSize:      30,
			userExists:   true,
			wantError:    nil,
			expectedUsed: 50,
		},
		{
			name:         "success - exact quota limit",
			initialQuota: 100,
			initialUsed:  20,
			addSize:      80,
			userExists:   true,
			wantError:    nil,
			expectedUsed: 100,
		},
		{
			name:         "failure - exceeds quota limits",
			initialQuota: 100,
			initialUsed:  20,
			addSize:      81,
			userExists:   true,
			wantError:    user.ErrUserNotFound, // RowsAffected == 0 returns ErrUserNotFound
			expectedUsed: 20,
		},
		{
			name:         "failure - user not found",
			initialQuota: 100,
			initialUsed:  20,
			addSize:      10,
			userExists:   false,
			wantError:    user.ErrUserNotFound,
			expectedUsed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)

			var userID uuid.UUID
			if tt.userExists {
				u := newUser().storageQuota(tt.initialQuota).storageUsed(tt.initialUsed).build()
				mustCreateUser(t, repo, ctx, u)
				userID = u.ID
			} else {
				userID = uuid.New()
			}

			err := repo.AddStorageUsed(ctx, userID, tt.addSize)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("AddStorageUsed() error: got %v, want %v", err, tt.wantError)
				return
			}

			if tt.userExists {
				got := mustGetUser(t, repo, ctx, userID)
				if got.StorageUsed != tt.expectedUsed {
					t.Errorf("StorageUsed: got %d, want %d", got.StorageUsed, tt.expectedUsed)
				}
			}
		})
	}
}

func TestUserRepository_SubtractStorageUsed(t *testing.T) {
	tests := []struct {
		name         string
		initialUsed  int64
		subSize      int64
		userExists   bool
		wantError    error
		expectedUsed int64
	}{
		{
			name:         "success - valid subtraction",
			initialUsed:  50,
			subSize:      30,
			userExists:   true,
			wantError:    nil,
			expectedUsed: 20,
		},
		{
			name:         "success - subtract to exact zero",
			initialUsed:  50,
			subSize:      50,
			userExists:   true,
			wantError:    nil,
			expectedUsed: 0,
		},
		{
			name:         "failure - subtracts below zero",
			initialUsed:  50,
			subSize:      51,
			userExists:   true,
			wantError:    user.ErrUserNotFound, // RowsAffected == 0 returns ErrUserNotFound
			expectedUsed: 50,
		},
		{
			name:         "failure - user not found",
			initialUsed:  50,
			subSize:      10,
			userExists:   false,
			wantError:    user.ErrUserNotFound,
			expectedUsed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ctx := setupUserTest(t)

			var userID uuid.UUID
			if tt.userExists {
				u := newUser().storageUsed(tt.initialUsed).build()
				mustCreateUser(t, repo, ctx, u)
				userID = u.ID
			} else {
				userID = uuid.New()
			}

			err := repo.SubtractStorageUsed(ctx, userID, tt.subSize)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("SubtractStorageUsed() error: got %v, want %v", err, tt.wantError)
				return
			}

			if tt.userExists {
				got := mustGetUser(t, repo, ctx, userID)
				if got.StorageUsed != tt.expectedUsed {
					t.Errorf("StorageUsed: got %d, want %d", got.StorageUsed, tt.expectedUsed)
				}
			}
		})
	}
}
