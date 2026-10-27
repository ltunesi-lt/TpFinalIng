package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresTripRepository struct {
	db DBTX
}

func NewPostgresTripRepository(db DBTX) *PostgresTripRepository {
	return &PostgresTripRepository{db: db}
}

func (r *PostgresTripRepository) GetSummaryByID(ctx context.Context, id string) ([]byte, error) {
	query := `SELECT itinerary_summary FROM trips WHERE id = $1`

	var summary []byte
	err := r.db.QueryRow(ctx, query, id).Scan(&summary)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("viaje no encontrado: %w", err)
		}
		return nil, fmt.Errorf("error al consultar el resumen del viaje: %w", err)
	}

	return summary, nil
}

func (r *PostgresTripRepository) UpdateSummary(ctx context.Context, id string, summary []byte) error {
	func (r *PostgresTripRepository) UpdateSummary(ctx context.Context, id string, summary []byte, currentVersion int) error {
    query := `
        UPDATE trips 
        SET itinerary_summary = $1::jsonb, 
            version = version + 1, 
            updated_at = NOW() 
        WHERE id = $2 AND version = $3`

    result, err := r.db.Exec(ctx, query, string(summary), id, currentVersion)
    if err != nil {
        return fmt.Errorf("error al actualizar el resumen del viaje: %w", err)
    }

    // Si RowsAffected() es 0, significa que otro proceso modificó el viaje 
    // al mismo tiempo (colisión de versión) o el ID no existe.
    if result.RowsAffected() == 0 {
        return fmt.Errorf("conflicto de concurrencia o viaje no encontrado para ID %s (versión %d)", id, currentVersion)
    }

    return nil
}
}
