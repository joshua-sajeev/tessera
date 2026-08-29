package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/application/auth"
	"github.com/joshua-sajeev/tessera/internal/domain/user"
)

func TestAuthenticator_Integration(t *testing.T) {
	cleanDB(t)

	userRepo := postgres.NewUserRepository(db)
	authenticator := auth.NewAuthenticator(userRepo, "tsr", "v1")
	generator := user.NewAPIKeyGenerator("tsr", "v1")
	hasher := user.NewKeyHasher()

	ctx := context.Background()

	// 1. Write integration test for user creation and API key storage
	rawKey, keyID, err := generator.Generate()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	hashedKey, err := hasher.Hash(rawKey)
	if err != nil {
		t.Fatalf("failed to hash key: %v", err)
	}

	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	activeUser := &user.User{
		ID:           userID,
		Username:     "integration-active",
		Email:        "integration-active@tessera.io",
		APIKeyID:     keyID,
		APIKeyHash:   hashedKey,
		StorageQuota: 1000,
		StorageUsed:  0,
		Status:       string(user.Active),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = userRepo.Create(ctx, activeUser)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 2. Write integration test for API key lookup and verification
	t.Run("successful lookup and verification", func(t *testing.T) {
		got, err := authenticator.GetUserByAPIKey(ctx, rawKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != userID {
			t.Errorf("got user ID %v, want %v", got.ID, userID)
		}
		if got.Username != "integration-active" {
			t.Errorf("got username %s, want %s", got.Username, "integration-active")
		}
	})

	// 3. Test edge cases: invalid keys, deleted users, suspended users
	t.Run("invalid key", func(t *testing.T) {
		prefix, version, randomPart, _ := generator.ParseKey(rawKey)
		badRandomPart := randomPart[:16] + "X" + randomPart[17:]
		badKey := prefix + "_" + version + "_" + badRandomPart

		_, err := authenticator.GetUserByAPIKey(ctx, badKey)
		if !errors.Is(err, user.ErrInvalidAPIKey) {
			t.Errorf("got error %v, want %v", err, user.ErrInvalidAPIKey)
		}
	})

	t.Run("unknown API key id", func(t *testing.T) {
		prefix, version, randomPart, _ := generator.ParseKey(rawKey)

		badRandomPart := "X" + randomPart[1:]
		badKey := prefix + "_" + version + "_" + badRandomPart

		_, err := authenticator.GetUserByAPIKey(ctx, badKey)

		if !errors.Is(err, user.ErrUserNotFound) {
			t.Errorf("got error %v, want %v", err, user.ErrUserNotFound)
		}
	})

	t.Run("invalid key format", func(t *testing.T) {
		_, err := authenticator.GetUserByAPIKey(ctx, "invalid-format")
		if !errors.Is(err, user.ErrInvalidAPIKey) {
			t.Errorf("got error %v, want %v", err, user.ErrInvalidAPIKey)
		}
	})

	t.Run("suspended user", func(t *testing.T) {
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
			Username:     "integration-suspended",
			Email:        "integration-suspended@tessera.io",
			APIKeyID:     keyIDSuspended,
			APIKeyHash:   hashedKeySuspended,
			StorageQuota: 1000,
			StorageUsed:  0,
			Status:       string(user.Suspended),
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err = userRepo.Create(ctx, suspendedUser)
		if err != nil {
			t.Fatalf("failed to create suspended user: %v", err)
		}

		_, err = authenticator.GetUserByAPIKey(ctx, rawKeySuspended)
		if !errors.Is(err, user.ErrUserSuspended) {
			t.Errorf("got error %v, want %v", err, user.ErrUserSuspended)
		}
	})

	t.Run("deleted user", func(t *testing.T) {
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
			Username:     "integration-deleted",
			Email:        "integration-deleted@tessera.io",
			APIKeyID:     keyIDDeleted,
			APIKeyHash:   hashedKeyDeleted,
			StorageQuota: 1000,
			StorageUsed:  0,
			Status:       string(user.Deleted),
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err = userRepo.Create(ctx, deletedUser)
		if err != nil {
			t.Fatalf("failed to create deleted user: %v", err)
		}

		_, err = authenticator.GetUserByAPIKey(ctx, rawKeyDeleted)
		if !errors.Is(err, user.ErrUserDeleted) {
			t.Errorf("got error %v, want %v", err, user.ErrUserDeleted)
		}
	})
}
