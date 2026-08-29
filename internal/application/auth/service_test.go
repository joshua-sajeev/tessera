package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/application/auth"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

type mockUserRepository struct {
	users map[string]*user.User // apiKeyID -> user
}

func (m *mockUserRepository) Create(ctx context.Context, u *user.User) error {
	m.users[u.APIKeyID] = u
	return nil
}

func (m *mockUserRepository) Get(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByAPIKeyID(ctx context.Context, apiKeyID string) (*user.User, error) {
	u, ok := m.users[apiKeyID]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status user.UserStatus) error {
	return nil
}

func (m *mockUserRepository) AddStorageUsed(ctx context.Context, id uuid.UUID, size int64) error {
	return nil
}

func (m *mockUserRepository) SubtractStorageUsed(ctx context.Context, id uuid.UUID, size int64) error {
	return nil
}

func TestAuthenticator_GetUserByAPIKey(t *testing.T) {
	repo := &mockUserRepository{users: make(map[string]*user.User)}
	authenticator := auth.NewAuthenticator(repo, "tsr", "v1")
	generator := user.NewAPIKeyGenerator("tsr", "v1")
	hasher := user.NewKeyHasher()

	ctx := context.Background()

	// 1. Success path: Active User
	rawKey, keyID, err := generator.Generate()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	hashedKey, err := hasher.Hash(rawKey)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	activeUserID := uuid.New()
	now := time.Now().UTC()
	activeUser := &user.User{
		ID:           activeUserID,
		Username:     "active-user",
		Email:        "active@tessera.io",
		APIKeyID:     keyID,
		APIKeyHash:   hashedKey,
		StorageQuota: 1000,
		StorageUsed:  0,
		Status:       string(user.Active),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_ = repo.Create(ctx, activeUser)

	got, err := authenticator.GetUserByAPIKey(ctx, rawKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != activeUserID {
		t.Errorf("got user ID %v, want %v", got.ID, activeUserID)
	}

	// 2. Suspended User path
	rawKeySuspended, keyIDSuspended, err := generator.Generate()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	hashedKeySuspended, err := hasher.Hash(rawKeySuspended)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	suspendedUserID := uuid.New()
	suspendedUser := &user.User{
		ID:           suspendedUserID,
		Username:     "suspended-user",
		Email:        "suspended@tessera.io",
		APIKeyID:     keyIDSuspended,
		APIKeyHash:   hashedKeySuspended,
		StorageQuota: 1000,
		StorageUsed:  0,
		Status:       string(user.Suspended),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_ = repo.Create(ctx, suspendedUser)

	_, err = authenticator.GetUserByAPIKey(ctx, rawKeySuspended)
	if !errors.Is(err, user.ErrUserSuspended) {
		t.Errorf("got error %v, want %v", err, user.ErrUserSuspended)
	}

	// 3. Deleted User path
	rawKeyDeleted, keyIDDeleted, err := generator.Generate()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	hashedKeyDeleted, err := hasher.Hash(rawKeyDeleted)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	deletedUserID := uuid.New()
	deletedUser := &user.User{
		ID:           deletedUserID,
		Username:     "deleted-user",
		Email:        "deleted@tessera.io",
		APIKeyID:     keyIDDeleted,
		APIKeyHash:   hashedKeyDeleted,
		StorageQuota: 1000,
		StorageUsed:  0,
		Status:       string(user.Deleted),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_ = repo.Create(ctx, deletedUser)

	_, err = authenticator.GetUserByAPIKey(ctx, rawKeyDeleted)
	if !errors.Is(err, user.ErrUserDeleted) {
		t.Errorf("got error %v, want %v", err, user.ErrUserDeleted)
	}

	// 4. Invalid key secret (wrong API key secret but correct prefix/format)
	prefix, version, randomPart, _ := generator.ParseKey(rawKey)
	apiKeyID := randomPart[:16]
	badRandomPart := apiKeyID + "A" + randomPart[17:]
	badKey := prefix + "_" + version + "_" + badRandomPart

	_, err = authenticator.GetUserByAPIKey(ctx, badKey)
	if !errors.Is(err, user.ErrInvalidAPIKey) {
		t.Errorf("got error %v, want %v", err, user.ErrInvalidAPIKey)
	}

	// 5. Invalid format key
	_, err = authenticator.GetUserByAPIKey(ctx, "invalid_format")
	if !errors.Is(err, user.ErrInvalidAPIKey) {
		t.Errorf("got error %v, want %v", err, user.ErrInvalidAPIKey)
	}
}
