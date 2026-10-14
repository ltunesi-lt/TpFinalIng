package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestGetSummaryByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresTripRepository(mock)
	ctx := context.Background()
	tripID := "trip-123"
	expectedJSON := []byte(`{"destination": "Buenos Aires", "flights": []}`)

	rows := pgxmock.NewRows([]string{"itinerary_summary"}).AddRow(expectedJSON)
	mock.ExpectQuery(`SELECT itinerary_summary FROM trips WHERE id = \$1`).
		WithArgs(tripID).
		WillReturnRows(rows)

	summary, err := repo.GetSummaryByID(ctx, tripID)

	assert.NoError(t, err)
	assert.Equal(t, expectedJSON, summary)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSummaryByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresTripRepository(mock)
	ctx := context.Background()
	tripID := "trip-999"

	mock.ExpectQuery(`SELECT itinerary_summary FROM trips WHERE id = \$1`).
		WithArgs(tripID).
		WillReturnError(pgx.ErrNoRows)

	summary, err := repo.GetSummaryByID(ctx, tripID)

	assert.Error(t, err)
	assert.Nil(t, summary)
	assert.Contains(t, err.Error(), "viaje no encontrado")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkDone_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := NewPostgresTaskRepository(mock)
	ctx := context.Background()
	taskID := "task-456"

	mock.ExpectExec(`UPDATE tasks SET status = 'DONE' WHERE id = \$1`).
		WithArgs(taskID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.MarkDone(ctx, taskID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
