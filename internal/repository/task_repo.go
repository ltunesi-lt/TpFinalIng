package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

type PostgresTaskRepository struct {
	db DBTX
}

func NewPostgresTaskRepository(db DBTX) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) ClaimNextPending(ctx context.Context) (*domain.Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción para tarea: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT id, type, payload, status, attempts, created_at
		FROM tasks
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED;`

	var task domain.Task
	var statusStr string

	err = tx.QueryRow(ctx, query).Scan(
		&task.ID,
		&task.Type,
		&task.Payload,
		&statusStr,
		&task.Attempts,
		&task.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al reclamar tarea pendiente: %w", err)
	}

	task.Status = domain.TaskStatus(statusStr)

	updateQuery := `
		UPDATE tasks 
		SET status = 'PROCESSING', attempts = attempts + 1 
		WHERE id = $1;`

	_, err = tx.Exec(ctx, updateQuery, task.ID)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar estado de tarea a PROCESSING: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción de tarea: %w", err)
	}

	task.Status = domain.TaskProcessing
	task.Attempts++
	return &task, nil
}

func (r *PostgresTaskRepository) MarkDone(ctx context.Context, id string) error {
	query := `UPDATE tasks SET status = 'DONE' WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error al marcar tarea como DONE: %w", err)
	}
	return nil
}

func (r *PostgresTaskRepository) MarkFailed(ctx context.Context, id string, reason string) error {
	query := `UPDATE tasks SET status = 'FAILED' WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error al marcar tarea como FAILED: %w", err)
	}
	return nil
}

func (r *PostgresTaskRepository) MarkRetry(ctx context.Context, id string) error {
	query := `UPDATE tasks SET status = 'PENDING' WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error al devolver tarea %s a PENDING para reintento: %w", id, err)
	}
	return nil
}