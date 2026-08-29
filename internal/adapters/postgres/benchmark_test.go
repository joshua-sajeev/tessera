package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
)

func BenchmarkAssetRepository(b *testing.B) {
	ctx := context.Background()
	repo := postgres.NewAssetRepository(db)

	// Setup: create a dedicated benchmark user.
	userID := uuid.New()
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, username, email, api_key_id, api_key_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "bench-user-"+userID.String()[:8], "bench@test.com", "bench-key-id", "bench-hash", "active")
	if err != nil {
		b.Fatalf("failed to insert bench user: %v", err)
	}

	// Pre-populate assets so SELECT benchmarks operate on a realistic dataset.
	assetIDs := make([]uuid.UUID, 0, 100)

	for i := range 100 {
		now := time.Now().UTC().Truncate(time.Microsecond)

		ast := &asset.Asset{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: fmt.Sprintf("file-%d.png", i),
			ContentType:      "image/png",
			Size:             int64(5000 + i),
			StoragePath:      fmt.Sprintf("uploads/file-%d.png", i),
			Status:           asset.StatusUploaded,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		if err := repo.Create(ctx, ast); err != nil {
			b.Fatalf("failed to seed asset: %v", err)
		}

		assetIDs = append(assetIDs, ast.ID)
	}

	b.Run("GetAsset", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			_, err := repo.Get(ctx, assetIDs[0], userID)
			if err != nil {
				b.Fatalf("Get error: %v", err)
			}
		}
	})

	b.Run("ListAssets", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			_, err := repo.ListByUser(ctx, userID, 20, 0)
			if err != nil {
				b.Fatalf("List error: %v", err)
			}
		}
	})

	b.Run("ListActiveAssets", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			_, err := repo.ListActiveByUser(ctx, userID, 20, 0)
			if err != nil {
				b.Fatalf("ListActive error: %v", err)
			}
		}
	})

	b.Run("CreateAsset", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			ast := &asset.Asset{
				ID:               uuid.New(),
				UserID:           userID,
				OriginalFilename: "bench-upload.jpg",
				ContentType:      "image/jpeg",
				Size:             5000,
				StoragePath:      "uploads/bench.jpg",
				Status:           asset.StatusUploaded,
				CreatedAt:        time.Now().UTC().Truncate(time.Microsecond),
				UpdatedAt:        time.Now().UTC().Truncate(time.Microsecond),
			}

			if err := repo.Create(ctx, ast); err != nil {
				b.Fatalf("Create error: %v", err)
			}
		}
	})
}
