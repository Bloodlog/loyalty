package repositories

import (
	"context"
	"fmt"
	"gophermart/internal/app/entities"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepositoryInterface interface {
	GetPendingJobs(ctx context.Context, tx pgx.Tx, limit int) ([]entities.Job, error)
	SaveJob(ctx context.Context, tx pgx.Tx, job *entities.Job) error
	UpdateJobPoolAt(ctx context.Context, tx pgx.Tx, jobID int64) error
	DeleteJobByID(ctx context.Context, tx pgx.Tx, jobID int64) error
}

type jobRepository struct {
	Pool *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) JobRepositoryInterface {
	return &jobRepository{
		Pool: db,
	}
}

func (r *jobRepository) GetPendingJobs(ctx context.Context, tx pgx.Tx, limit int) ([]entities.Job, error) {
	query := `
		SELECT id, order_id, created_at, pool_at
		FROM jobs
		ORDER BY created_at ASC, pool_at ASC
		LIMIT $1
	`

	var rows pgx.Rows
	var err error

	if tx != nil {
		rows, err = tx.Query(ctx, query, limit)
	} else {
		rows, err = r.Pool.Query(ctx, query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []entities.Job
	for rows.Next() {
		var job entities.Job
		err = rows.Scan(
			&job.ID,
			&job.OrderID,
			&job.CreatedAt,
			&job.PoolAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse job result %d: %w", job.ID, err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}

	return jobs, nil
}

func (r *jobRepository) SaveJob(ctx context.Context, tx pgx.Tx, job *entities.Job) error {
	query := `
		INSERT INTO jobs (order_id, created_at, pool_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, job.OrderID, job.CreatedAt, job.PoolAt).Scan(&job.ID)
	} else {
		err = r.Pool.QueryRow(ctx, query, job.OrderID, job.CreatedAt, job.PoolAt).Scan(&job.ID)
	}

	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	return nil
}

func (r *jobRepository) UpdateJobPoolAt(ctx context.Context, tx pgx.Tx, jobID int64) error {
	query := `
		UPDATE jobs
		SET pool_at = $1
		WHERE id = $2
	`

	if tx != nil {
		_, err := tx.Exec(ctx, query, time.Now(), jobID)
		if err != nil {
			return fmt.Errorf("failed to update pool_at for job %d: %w", jobID, err)
		}
	} else {
		_, err := r.Pool.Exec(ctx, query, time.Now(), jobID)
		if err != nil {
			return fmt.Errorf("failed to update pool_at for job %d: %w", jobID, err)
		}
	}

	return nil
}

func (r *jobRepository) DeleteJobByID(ctx context.Context, tx pgx.Tx, jobID int64) error {
	query := `
		DELETE FROM jobs
		WHERE id = $1
	`

	if tx != nil {
		_, err := tx.Exec(ctx, query, jobID)
		if err != nil {
			return fmt.Errorf("failed to delete job %d: %w", jobID, err)
		}
	} else {
		_, err := r.Pool.Exec(ctx, query, jobID)
		if err != nil {
			return fmt.Errorf("failed to delete job %d: %w", jobID, err)
		}
	}

	return nil
}
