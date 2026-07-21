package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/domain/processing"
)

func TestProcessingRepository_Create(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewProcessingRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	want := &processing.Job{
		ID:        uuid.New(),
		AssetID:   uuid.New(),
		Status:    processing.StatusQueued,
		CreatedAt: now,
	}

	createTestAsset(t, want.AssetID)

	if err := repo.Create(ctx, want); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := repo.Get(ctx, want.ID)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if got.ID != want.ID {
		t.Errorf("ID = %v, want %v", got.ID, want.ID)
	}

	if got.AssetID != want.AssetID {
		t.Errorf("AssetID = %v, want %v", got.AssetID, want.AssetID)
	}

	if got.Status != want.Status {
		t.Errorf("Status = %q, want %q", got.Status, want.Status)
	}

	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, want.CreatedAt)
	}

	if got.StartedAt != nil {
		t.Errorf("StartedAt = %v, want nil", got.StartedAt)
	}

	if got.CompletedAt != nil {
		t.Errorf("CompletedAt = %v, want nil", got.CompletedAt)
	}
}

func TestProcessingRepository_Get(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewProcessingRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	assetID := uuid.New()

	createTestAsset(t, assetID)

	job := &processing.Job{
		ID:        uuid.New(),
		AssetID:   assetID,
		Status:    processing.StatusQueued,
		CreatedAt: now,
	}

	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	got, err := repo.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	if got.ID != job.ID {
		t.Fatalf("got ID %v, want %v", got.ID, job.ID)
	}
}

func TestProcessingRepository_Get_NotFound(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewProcessingRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, uuid.New())
	if !errors.Is(err, processing.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessingRepository_Update(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewProcessingRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	assetID := uuid.New()

	createTestAsset(t, assetID)

	job := &processing.Job{
		ID:        uuid.New(),
		AssetID:   assetID,
		Status:    processing.StatusQueued,
		CreatedAt: now,
	}

	if err := repo.Create(ctx, job); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	started := now.Add(time.Minute)
	completed := now.Add(2 * time.Minute)

	job.Status = processing.StatusCompleted
	job.StartedAt = &started
	job.CompletedAt = &completed

	if err := repo.Update(ctx, job); err != nil {
		t.Fatalf("Update(): %v", err)
	}

	got, err := repo.Get(ctx, job.ID)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	if got.Status != processing.StatusCompleted {
		t.Errorf("Status = %q, want %q", got.Status, processing.StatusCompleted)
	}

	if got.StartedAt == nil || !got.StartedAt.Equal(started) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, started)
	}

	if got.CompletedAt == nil || !got.CompletedAt.Equal(completed) {
		t.Errorf("CompletedAt = %v, want %v", got.CompletedAt, completed)
	}
}

func TestProcessingRepository_Update_NotFound(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewProcessingRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	job := &processing.Job{
		ID:        uuid.New(),
		AssetID:   uuid.New(),
		Status:    processing.StatusQueued,
		CreatedAt: now,
	}

	err := repo.Update(ctx, job)
	if !errors.Is(err, processing.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
