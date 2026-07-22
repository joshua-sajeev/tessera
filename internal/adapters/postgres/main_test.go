package postgres_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/domain/asset"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var db *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:17",
		tcpostgres.WithDatabase("tessera_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}

	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminate postgres container: %v", err)
		}
	}()

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("connection string: %v", err)
	}

	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("open sql db: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close sql db: %v", err)
		}
	}()

	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	db, err = pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("create pgx pool: %v", err)
	}
	defer db.Close()

	os.Exit(m.Run())
}

func cleanDB(t *testing.T) {
	t.Helper()

	_, err := db.Exec(
		t.Context(),
		`
		TRUNCATE TABLE
			asset_variants,
			processing_jobs,
			assets
		CASCADE;
		`,
	)
	if err != nil {
		t.Fatalf("clean database: %v", err)
	}
}

// createTestAsset is a helper that creates an asset in the database for testing purposes.
// It should be used for setup in tests that need an existing asset (e.g., processing job tests).
func createTestAsset(t *testing.T, id uuid.UUID) *asset.Asset {
	t.Helper()

	repo := postgres.NewAssetRepository(db)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)

	a := &asset.Asset{
		ID:               id,
		OriginalFilename: "test-image.png",
		ContentType:      "image/png",
		Size:             1024,
		StoragePath:      "uploads/test-image.png",
		Status:           asset.StatusUploaded,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("createTestAsset: failed to create asset: %v", err)
	}

	return a
}
