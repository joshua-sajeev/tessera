package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
)

var (
	testAssetTimeMu sync.Mutex
	testAssetTime   = time.Now().UTC().Truncate(time.Microsecond).Add(-10 * time.Hour)
)

func makeTestAssetStruct(userID uuid.UUID, filename string) *asset.Asset {
	testAssetTimeMu.Lock()
	testAssetTime = testAssetTime.Add(time.Second)
	now := testAssetTime
	testAssetTimeMu.Unlock()
	return &asset.Asset{
		ID:               uuid.New(),
		UserID:           userID,
		OriginalFilename: filename,
		ContentType:      "text/plain",
		Size:             1024,
		StoragePath:      "uploads/" + filename,
		Status:           asset.StatusUploaded,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createLocalTestUserWithStatus(t *testing.T, userID uuid.UUID, status string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now().UTC().Truncate(time.Microsecond)
	query := `
		INSERT INTO users (
			id,
			username,
			email,
			api_key_id,
			api_key_hash,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (id) DO UPDATE SET status = $6, updated_at = $7
	`

	_, err := db.Exec(
		ctx,
		query,
		userID,
		"test-"+userID.String(),
		userID.String()+"@test.com",
		"test-key-id-"+userID.String(),
		"test-hash-"+userID.String(),
		status,
		now,
	)
	if err != nil {
		t.Fatalf("createLocalTestUserWithStatus: failed to create user with status %q: %v", status, err)
	}
}

func TestAssetRepository_Create(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	userID := uuid.New()

	createTestUser(t, userID)

	want := &asset.Asset{
		ID:               uuid.New(),
		UserID:           userID,
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

	got, err := repo.Get(ctx, want.ID, want.UserID)
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

	want := createTestAsset(t, uuid.New(), uuid.New())

	got, err := repo.Get(ctx, want.ID, want.UserID)
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

	_, err := repo.Get(ctx, uuid.New(), uuid.New())
	if !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAssetRepository_UpdateStatus(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	a := createTestAsset(t, uuid.New(), userID)

	a.Status = asset.StatusProcessed

	if err := repo.UpdateStatus(ctx, a.ID, userID, a.Status); err != nil {
		t.Fatalf("UpdateStatus(): %v", err)
	}

	got, err := repo.Get(ctx, a.ID, userID)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	if got.Status != a.Status {
		t.Errorf("Status = %q, want %q",
			got.Status, a.Status)
	}
}

func TestAssetRepository_UpdateStatus_NotFound(t *testing.T) {
	cleanDB(t)

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	assetID := uuid.New()

	err := repo.UpdateStatus(
		ctx,
		assetID,
		userID,
		asset.StatusProcessed,
	)

	if !errors.Is(err, asset.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAssetRepository_ListByUser(t *testing.T) {
	tests := []struct {
		name           string
		limit          int
		offset         int
		setupAssets    func(t *testing.T, userID uuid.UUID) []*asset.Asset
		wantCount      int
		wantFilenames  []string
		wantNotContain []string
	}{
		{
			name:   "list all assets for user without pagination",
			limit:  100,
			offset: 0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:     3,
			wantFilenames: []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
		},
		{
			name:   "list assets with limit pagination",
			limit:  2,
			offset: 0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:      2,
			wantFilenames:  []string{"test-file-2.txt", "test-file-3.txt"},
			wantNotContain: []string{"test-file-1.txt"},
		},
		{
			name:   "list assets with offset pagination",
			limit:  2,
			offset: 1,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:      2,
			wantFilenames:  []string{"test-file-1.txt", "test-file-2.txt"},
			wantNotContain: []string{"test-file-3.txt"},
		},
		{
			name:   "list assets with offset beyond available",
			limit:  10,
			offset: 100,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
				}
			},
			wantCount:     0,
			wantFilenames: []string{},
		},
		{
			name:   "list assets for user with no assets",
			limit:  10,
			offset: 0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{}
			},
			wantCount:     0,
			wantFilenames: []string{},
		},
		{
			name:   "list returns only requested user's assets",
			limit:  100,
			offset: 0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				otherUserID := uuid.New()
				createTestUser(t, otherUserID)
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(otherUserID, "test-file-3.txt"),
				}
			},
			wantCount:      2,
			wantFilenames:  []string{"test-file-1.txt", "test-file-2.txt"},
			wantNotContain: []string{"test-file-3.txt"},
		},
		{
			name:   "list includes all statuses regardless of lifecycle state",
			limit:  100,
			offset: 0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				now := time.Now().UTC().Truncate(time.Microsecond)
				return []*asset.Asset{
					{
						ID:               uuid.New(),
						UserID:           userID,
						OriginalFilename: "processed.pdf",
						ContentType:      "application/pdf",
						Size:             1024,
						StoragePath:      "uploads/processed.pdf",
						Status:           asset.StatusProcessed,
						CreatedAt:        now.Add(-2 * time.Hour),
						UpdatedAt:        now.Add(-1 * time.Hour),
					},
					{
						ID:               uuid.New(),
						UserID:           userID,
						OriginalFilename: "failed.jpg",
						ContentType:      "image/jpeg",
						Size:             2048,
						StoragePath:      "uploads/failed.jpg",
						Status:           asset.StatusFailed,
						CreatedAt:        now.Add(-1 * time.Hour),
						UpdatedAt:        now,
					},
					{
						ID:               uuid.New(),
						UserID:           userID,
						OriginalFilename: "processing.png",
						ContentType:      "image/png",
						Size:             4096,
						StoragePath:      "uploads/processing.png",
						Status:           asset.StatusProcessing,
						CreatedAt:        now,
						UpdatedAt:        now,
					},
				}
			},
			wantCount:     3,
			wantFilenames: []string{"processed.pdf", "failed.jpg", "processing.png"},
		},
		{
			name:   "list with limit of 1",
			limit:  1,
			offset: 0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:      1,
			wantFilenames:  []string{"test-file-3.txt"},
			wantNotContain: []string{"test-file-1.txt", "test-file-2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t)
			repo := postgres.NewAssetRepository(db)
			ctx := context.Background()
			userID := uuid.New()
			createTestUser(t, userID)

			// Setup assets
			assets := tt.setupAssets(t, userID)
			for _, a := range assets {
				if err := repo.Create(ctx, a); err != nil {
					t.Fatalf("Create() error: %v", err)
				}
			}

			// Act
			results, err := repo.ListByUser(ctx, userID, tt.limit, tt.offset)
			// Assert
			if err != nil {
				t.Fatalf("ListByUser() error: %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("ListByUser() returned %d assets, want %d", len(results), tt.wantCount)
			}

			if len(tt.wantFilenames) > 0 {
				resultFilenames := make(map[string]bool)
				for _, r := range results {
					resultFilenames[r.OriginalFilename] = true
				}
				for _, filename := range tt.wantFilenames {
					if !resultFilenames[filename] {
						t.Errorf("expected filename %q not found in results", filename)
					}
				}
			}

			if len(tt.wantNotContain) > 0 {
				resultFilenames := make(map[string]bool)
				for _, r := range results {
					resultFilenames[r.OriginalFilename] = true
				}
				for _, filename := range tt.wantNotContain {
					if resultFilenames[filename] {
						t.Errorf("unexpected filename %q found in results", filename)
					}
				}
			}
		})
	}
}

func TestAssetRepository_ListByUser_OrderDesc(t *testing.T) {
	cleanDB(t)
	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createTestUser(t, userID)

	now := time.Now().UTC().Truncate(time.Microsecond)
	assets := []*asset.Asset{
		{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: "oldest.txt",
			ContentType:      "text/plain",
			Size:             1024,
			StoragePath:      "uploads/oldest.txt",
			Status:           asset.StatusUploaded,
			CreatedAt:        now.Add(-3 * time.Hour),
			UpdatedAt:        now.Add(-3 * time.Hour),
		},
		{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: "middle.txt",
			ContentType:      "text/plain",
			Size:             2048,
			StoragePath:      "uploads/middle.txt",
			Status:           asset.StatusUploaded,
			CreatedAt:        now.Add(-1 * time.Hour),
			UpdatedAt:        now.Add(-1 * time.Hour),
		},
		{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: "newest.txt",
			ContentType:      "text/plain",
			Size:             4096,
			StoragePath:      "uploads/newest.txt",
			Status:           asset.StatusUploaded,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	for _, a := range assets {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create(): %v", err)
		}
	}

	results, err := repo.ListByUser(ctx, userID, 100, 0)
	if err != nil {
		t.Fatalf("ListByUser(): %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].OriginalFilename != "newest.txt" {
		t.Errorf("first result filename = %q, want newest.txt", results[0].OriginalFilename)
	}

	if results[1].OriginalFilename != "middle.txt" {
		t.Errorf("second result filename = %q, want middle.txt", results[1].OriginalFilename)
	}

	if results[2].OriginalFilename != "oldest.txt" {
		t.Errorf("third result filename = %q, want oldest.txt", results[2].OriginalFilename)
	}
}

func TestAssetRepository_ListActiveByUser(t *testing.T) {
	tests := []struct {
		name           string
		userStatus     string // "active" or "suspended"
		limit          int
		offset         int
		setupAssets    func(t *testing.T, userID uuid.UUID) []*asset.Asset
		wantCount      int
		wantFilenames  []string
		wantNotContain []string
	}{
		{
			name:       "list active user's assets without pagination",
			userStatus: "active",
			limit:      100,
			offset:     0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:     3,
			wantFilenames: []string{"test-file-1.txt", "test-file-2.txt", "test-file-3.txt"},
		},
		{
			name:       "list excludes suspended user's assets",
			userStatus: "suspended",
			limit:      100,
			offset:     0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
				}
			},
			wantCount:      0,
			wantFilenames:  []string{},
			wantNotContain: []string{"test-file-1.txt", "test-file-2.txt"},
		},
		{
			name:       "list respects limit for active user",
			userStatus: "active",
			limit:      2,
			offset:     0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:      2,
			wantFilenames:  []string{"test-file-2.txt", "test-file-3.txt"},
			wantNotContain: []string{"test-file-1.txt"},
		},
		{
			name:       "list respects offset for active user",
			userStatus: "active",
			limit:      2,
			offset:     1,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:      2,
			wantFilenames:  []string{"test-file-1.txt", "test-file-2.txt"},
			wantNotContain: []string{"test-file-3.txt"},
		},
		{
			name:       "list returns empty for active user with no assets",
			userStatus: "active",
			limit:      100,
			offset:     0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{}
			},
			wantCount:     0,
			wantFilenames: []string{},
		},
		{
			name:       "list includes all statuses for active user",
			userStatus: "active",
			limit:      100,
			offset:     0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				now := time.Now().UTC().Truncate(time.Microsecond)
				return []*asset.Asset{
					{
						ID:               uuid.New(),
						UserID:           userID,
						OriginalFilename: "processed.pdf",
						ContentType:      "application/pdf",
						Size:             1024,
						StoragePath:      "uploads/processed.pdf",
						Status:           asset.StatusProcessed,
						CreatedAt:        now.Add(-2 * time.Hour),
						UpdatedAt:        now.Add(-1 * time.Hour),
					},
					{
						ID:               uuid.New(),
						UserID:           userID,
						OriginalFilename: "failed.jpg",
						ContentType:      "image/jpeg",
						Size:             2048,
						StoragePath:      "uploads/failed.jpg",
						Status:           asset.StatusFailed,
						CreatedAt:        now.Add(-1 * time.Hour),
						UpdatedAt:        now,
					},
					{
						ID:               uuid.New(),
						UserID:           userID,
						OriginalFilename: "processing.png",
						ContentType:      "image/png",
						Size:             4096,
						StoragePath:      "uploads/processing.png",
						Status:           asset.StatusProcessing,
						CreatedAt:        now,
						UpdatedAt:        now,
					},
				}
			},
			wantCount:     3,
			wantFilenames: []string{"processed.pdf", "failed.jpg", "processing.png"},
		},
		{
			name:       "list offset beyond available for active user",
			userStatus: "active",
			limit:      10,
			offset:     100,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
				}
			},
			wantCount:     0,
			wantFilenames: []string{},
		},
		{
			name:       "list with limit of 1 for active user",
			userStatus: "active",
			limit:      1,
			offset:     0,
			setupAssets: func(t *testing.T, userID uuid.UUID) []*asset.Asset {
				return []*asset.Asset{
					makeTestAssetStruct(userID, "test-file-1.txt"),
					makeTestAssetStruct(userID, "test-file-2.txt"),
					makeTestAssetStruct(userID, "test-file-3.txt"),
				}
			},
			wantCount:      1,
			wantFilenames:  []string{"test-file-3.txt"},
			wantNotContain: []string{"test-file-1.txt", "test-file-2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanDB(t)
			repo := postgres.NewAssetRepository(db)
			ctx := context.Background()
			userID := uuid.New()
			createLocalTestUserWithStatus(t, userID, tt.userStatus)

			// Setup assets
			assets := tt.setupAssets(t, userID)
			for _, a := range assets {
				if err := repo.Create(ctx, a); err != nil {
					t.Fatalf("Create() error: %v", err)
				}
			}

			// Act
			results, err := repo.ListActiveByUser(ctx, userID, tt.limit, tt.offset)
			// Assert
			if err != nil {
				t.Fatalf("ListActiveByUser() error: %v", err)
			}

			if len(results) != tt.wantCount {
				t.Errorf("ListActiveByUser() returned %d assets, want %d", len(results), tt.wantCount)
			}

			if len(tt.wantFilenames) > 0 {
				resultFilenames := make(map[string]bool)
				for _, r := range results {
					resultFilenames[r.OriginalFilename] = true
				}
				for _, filename := range tt.wantFilenames {
					if !resultFilenames[filename] {
						t.Errorf("expected filename %q not found in results", filename)
					}
				}
			}

			if len(tt.wantNotContain) > 0 {
				resultFilenames := make(map[string]bool)
				for _, r := range results {
					resultFilenames[r.OriginalFilename] = true
				}
				for _, filename := range tt.wantNotContain {
					if resultFilenames[filename] {
						t.Errorf("unexpected filename %q found in results", filename)
					}
				}
			}
		})
	}
}

func TestAssetRepository_ListActiveByUser_OrderDesc(t *testing.T) {
	cleanDB(t)
	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	createLocalTestUserWithStatus(t, userID, "active")

	now := time.Now().UTC().Truncate(time.Microsecond)
	assets := []*asset.Asset{
		{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: "oldest.txt",
			ContentType:      "text/plain",
			Size:             1024,
			StoragePath:      "uploads/oldest.txt",
			Status:           asset.StatusUploaded,
			CreatedAt:        now.Add(-3 * time.Hour),
			UpdatedAt:        now.Add(-3 * time.Hour),
		},
		{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: "middle.txt",
			ContentType:      "text/plain",
			Size:             2048,
			StoragePath:      "uploads/middle.txt",
			Status:           asset.StatusUploaded,
			CreatedAt:        now.Add(-1 * time.Hour),
			UpdatedAt:        now.Add(-1 * time.Hour),
		},
		{
			ID:               uuid.New(),
			UserID:           userID,
			OriginalFilename: "newest.txt",
			ContentType:      "text/plain",
			Size:             4096,
			StoragePath:      "uploads/newest.txt",
			Status:           asset.StatusUploaded,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	for _, a := range assets {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create(): %v", err)
		}
	}

	results, err := repo.ListActiveByUser(ctx, userID, 100, 0)
	if err != nil {
		t.Fatalf("ListActiveByUser(): %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].OriginalFilename != "newest.txt" {
		t.Errorf("first result filename = %q, want newest.txt", results[0].OriginalFilename)
	}

	if results[1].OriginalFilename != "middle.txt" {
		t.Errorf("second result filename = %q, want middle.txt", results[1].OriginalFilename)
	}

	if results[2].OriginalFilename != "oldest.txt" {
		t.Errorf("third result filename = %q, want oldest.txt", results[2].OriginalFilename)
	}
}

func TestAssetRepository_ListActiveByUser_MultipleUsers(t *testing.T) {
	cleanDB(t)
	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	activeUserID := uuid.New()
	otherActiveUserID := uuid.New()
	suspendedUserID := uuid.New()

	createLocalTestUserWithStatus(t, activeUserID, "active")
	createLocalTestUserWithStatus(t, otherActiveUserID, "active")
	createLocalTestUserWithStatus(t, suspendedUserID, "suspended")

	// Create assets for all users
	activeUserAssets := []*asset.Asset{
		makeTestAssetStruct(activeUserID, "test-file-1.txt"),
		makeTestAssetStruct(activeUserID, "test-file-2.txt"),
	}
	otherUserAssets := []*asset.Asset{
		makeTestAssetStruct(otherActiveUserID, "test-file-3.txt"),
		makeTestAssetStruct(otherActiveUserID, "test-file-4.txt"),
	}
	suspendedUserAssets := []*asset.Asset{
		makeTestAssetStruct(suspendedUserID, "test-file-5.txt"),
		makeTestAssetStruct(suspendedUserID, "test-file-6.txt"),
	}

	for _, a := range append(activeUserAssets, append(otherUserAssets, suspendedUserAssets...)...) {
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create(): %v", err)
		}
	}

	// Test: active user sees only their assets
	results, err := repo.ListActiveByUser(ctx, activeUserID, 100, 0)
	if err != nil {
		t.Fatalf("ListActiveByUser(): %v", err)
	}

	if len(results) != 2 {
		t.Errorf("ListActiveByUser() for active user returned %d assets, want 2", len(results))
	}

	for _, a := range results {
		if a.UserID != activeUserID {
			t.Errorf("result has userID %v, want %v", a.UserID, activeUserID)
		}
	}

	// Test: suspended user sees no assets
	suspendedResults, err := repo.ListActiveByUser(ctx, suspendedUserID, 100, 0)
	if err != nil {
		t.Fatalf("ListActiveByUser(): %v", err)
	}

	if len(suspendedResults) != 0 {
		t.Errorf("ListActiveByUser() for suspended user returned %d assets, want 0", len(suspendedResults))
	}

	// Test: other active user sees only their assets
	otherResults, err := repo.ListActiveByUser(ctx, otherActiveUserID, 100, 0)
	if err != nil {
		t.Fatalf("ListActiveByUser(): %v", err)
	}

	if len(otherResults) != 2 {
		t.Errorf("ListActiveByUser() for other active user returned %d assets, want 2", len(otherResults))
	}

	for _, a := range otherResults {
		if a.UserID != otherActiveUserID {
			t.Errorf("result has userID %v, want %v", a.UserID, otherActiveUserID)
		}
	}
}
