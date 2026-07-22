package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
)

func TestAssetRepository_Create(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	want := &asset.Asset{
		ID:               uuid.New(),
		OriginalFilename: "photo.jpg",
		ContentType:      "image/jpeg",
		Size:             1024,
		StoragePath:      "uploads/photo.jpg",
		Status:           asset.StatusUploaded,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

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

	if got.OriginalFilename != want.OriginalFilename {
		t.Errorf("OriginalFilename = %q, want %q",
			got.OriginalFilename, want.OriginalFilename)
	}

	if got.ContentType != want.ContentType {
		t.Errorf("ContentType = %q, want %q",
			got.ContentType, want.ContentType)
	}

	if got.Size != want.Size {
		t.Errorf("Size = %d, want %d",
			got.Size, want.Size)
	}

	if got.StoragePath != want.StoragePath {
		t.Errorf("StoragePath = %q, want %q",
			got.StoragePath, want.StoragePath)
	}

	if got.Status != want.Status {
		t.Errorf("Status = %q, want %q",
			got.Status, want.Status)
	}

	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v",
			got.CreatedAt, want.CreatedAt)
	}

	if !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v",
			got.UpdatedAt, want.UpdatedAt)
	}
}

func TestAssetRepository_Get(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	want := createTestAsset(t, uuid.New())

	got, err := repo.Get(ctx, want.ID)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	if got.ID != want.ID {
		t.Fatalf("got ID %v, want %v", got.ID, want.ID)
	}
}

func TestAssetRepository_Get_NotFound(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, uuid.New())
	if !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAssetRepository_Update(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	a := createTestAsset(t, uuid.New())

	updatedAt := a.UpdatedAt.Add(time.Minute).Truncate(time.Microsecond)

	a.OriginalFilename = "updated.jpg"
	a.ContentType = "image/png"
	a.Size = 4096
	a.StoragePath = "processed/updated.png"
	a.Status = asset.StatusProcessed
	a.UpdatedAt = updatedAt

	if err := repo.Update(ctx, a); err != nil {
		t.Fatalf("Update(): %v", err)
	}

	got, err := repo.Get(ctx, a.ID)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	if got.OriginalFilename != a.OriginalFilename {
		t.Errorf("OriginalFilename = %q, want %q",
			got.OriginalFilename, a.OriginalFilename)
	}

	if got.ContentType != a.ContentType {
		t.Errorf("ContentType = %q, want %q",
			got.ContentType, a.ContentType)
	}

	if got.Size != a.Size {
		t.Errorf("Size = %d, want %d",
			got.Size, a.Size)
	}

	if got.StoragePath != a.StoragePath {
		t.Errorf("StoragePath = %q, want %q",
			got.StoragePath, a.StoragePath)
	}

	if got.Status != a.Status {
		t.Errorf("Status = %q, want %q",
			got.Status, a.Status)
	}

	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v",
			got.UpdatedAt, updatedAt)
	}
}

func TestAssetRepository_Update_NotFound(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	a := &asset.Asset{
		ID:               uuid.New(),
		OriginalFilename: "missing.jpg",
		ContentType:      "image/jpeg",
		Size:             100,
		StoragePath:      "uploads/missing.jpg",
		Status:           asset.StatusUploaded,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := repo.Update(ctx, a)
	if !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
