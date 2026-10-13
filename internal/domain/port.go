// internal/domain/ports.go
package domain

import (
	"context"
	"time"
)

type Trip struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ItinerarySummary []byte    `json:"itinerary_summary"` // Columna JSONB cruda para máxima velocidad
	UpdatedAt        time.Time `json:"updated_at"`
}

type TaskStatus string

const (
	TaskPending    TaskStatus = "PENDING"
	TaskProcessing TaskStatus = "PROCESSING"
	TaskDone       TaskStatus = "DONE"
	TaskFailed     TaskStatus = "FAILED"
)

type Task struct {
	ID        string
	Type      string
	Payload   []byte
	Status    TaskStatus
	Attempts  int
	CreatedAt time.Time
}

type ProviderPort interface {
	FetchUpdates(ctx context.Context, externalID string) ([]byte, error)
}

type TripRepository interface {
	GetSummaryByID(ctx context.Context, id string) ([]byte, error)
	UpdateSummary(ctx context.Context, id string, summary []byte) error
}

type TaskRepository interface {
	ClaimNextPending(ctx context.Context) (*Task, error)
	MarkDone(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, reason string) error
}