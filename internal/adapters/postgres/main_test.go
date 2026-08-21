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
	"github.com/joshua-sajeev/tessera/internal/config"
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
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(
		ctx,
		`
		TRUNCATE TABLE
			asset_variants,
			processing_jobs,
			assets,
			users
		CASCADE;
		`,
	)
	if err != nil {
		t.Fatalf("clean database: %v", err)
	}
}

func createTestUser(t *testing.T, id uuid.UUID) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(
		ctx,
		`
		INSERT INTO users (
			id,
			username,
			email,
			api_key_id,
			api_key_hash
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		id,
		"test-"+id.String(),
		id.String()+"@test.com",
		"test-key-id-"+id.String(),
		"test-hash-"+id.String(),
	)
	if err != nil {
		t.Fatalf("createTestUser: %v", err)
	}
}

// createTestAsset is a helper that creates an asset in the database for testing purposes.
// It should be used for setup in tests that need an existing asset (e.g., processing job tests).
func createTestAsset(t *testing.T, id uuid.UUID, userID uuid.UUID) *asset.Asset {
	t.Helper()
	createTestUser(t, userID)

	repo := postgres.NewAssetRepository(db)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	now := time.Now().UTC().Truncate(time.Microsecond)
	a := &asset.Asset{
		ID:               id,
		UserID:           userID,
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

func TestNewPool(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "tessera",
		Password: "tessera",
		Name:     "tessera",
		SSLMode:  "disable",
	}
	pool, err := postgres.NewPool(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()
	if pool == nil {
		t.Fatal("expected pool, got nil")
	}
}
