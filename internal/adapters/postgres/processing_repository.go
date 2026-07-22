package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joshua-sajeev/tessera/internal/domain/processing"
	"github.com/joshua-sajeev/tessera/internal/ports"
)

const (
	insertProcessingJobQuery = `
		INSERT INTO processing_jobs (
			id,
			asset_id,
			status,
			created_at,
			started_at,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	getProcessingJobQuery = `
		SELECT
			id,
			asset_id,
			status,
			created_at,
			started_at,
			completed_at
		FROM processing_jobs
		WHERE id = $1
	`

	updateProcessingJobQuery = `
		UPDATE processing_jobs
		SET
			status = $2,
			started_at = $3,
			completed_at = $4
		WHERE id = $1
	`
)

// ProcessingRepository implements the ProcessingRepository port using PostgreSQL.
type ProcessingRepository struct {
	db *pgxpool.Pool
}

// NewProcessingRepository creates a new PostgreSQL-backed processing repository.
func NewProcessingRepository(db *pgxpool.Pool) ports.ProcessingRepository {
	return &ProcessingRepository{
		db: db,
	}
}

// Ensure ProcessingRepository satisfies the ProcessingRepository port.
var _ ports.ProcessingRepository = (*ProcessingRepository)(nil)

func processingJobValues(job *processing.Job) []any {
	return []any{
		job.ID,
		job.AssetID,
		job.Status,
		job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
	}
}

func processingJobScanArgs(job *processing.Job) []any {
	return []any{
		&job.ID,
		&job.AssetID,
		&job.Status,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	}
}

// Create persists a new processing job.
func (r *ProcessingRepository) Create(ctx context.Context, job *processing.Job) error {
	_, err := r.db.Exec(ctx, insertProcessingJobQuery, processingJobValues(job)...)
	if err != nil {
		return fmt.Errorf("create processing job: %w", err)
	}

	return nil
}

// Get retrieves a processing job by its unique identifier.
func (r *ProcessingRepository) Get(ctx context.Context, id uuid.UUID) (*processing.Job, error) {
	var job processing.Job

	err := r.db.QueryRow(ctx, getProcessingJobQuery, id).Scan(processingJobScanArgs(&job)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, processing.ErrNotFound
		}
		return nil, fmt.Errorf("get processing job: %w", err)
	}

	return &job, nil
}

// Update persists changes to an existing processing job.
func (r *ProcessingRepository) Update(ctx context.Context, job *processing.Job) error {
	cmd, err := r.db.Exec(
		ctx,
		updateProcessingJobQuery,
		job.ID,
		job.Status,
		job.StartedAt,
		job.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("update processing job: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return processing.ErrNotFound
	}

	return nil
}
