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

// -----------------------------------------------------------------
// Puertos de Integración y Proveedores externos (Iteración 1 y 2)
// -----------------------------------------------------------------

type ProviderPort interface {
	FetchUpdates(ctx context.Context, externalID string) ([]byte, error)
}

// -----------------------------------------------------------------
// Puertos de Repositorios (Dominio <-> Infraestructura)
// -----------------------------------------------------------------

type TripRepository interface {
	GetSummaryByID(ctx context.Context, id string) ([]byte, error)
	UpdateSummary(ctx context.Context, id string, summary []byte) error
}

type TaskRepository interface {
	ClaimNextPending(ctx context.Context) (*Task, error)
	MarkDone(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, reason string) error
	MarkRetry(ctx context.Context, id string) error // Deforma la tarea a PENDING para el próximo ciclo
}

// UserRepository gestiona el almacenamiento seguro de credenciales/tokens de correo (SEC1)
type UserRepository interface {
	SaveUserMailToken(ctx context.Context, userID, encryptedToken string) error
	RevokeUserMailToken(ctx context.Context, userID string) error
}

// ShareRepository gestiona los enlaces compartidos de itinerarios con revocación instantánea (SEC2)
type ShareRepository interface {
	CreateShareLink(ctx context.Context, token string, tripID string, expiresAt time.Time) error
	ValidateAndGetTripID(ctx context.Context, token string) (string, error)
	RevokeShareLink(ctx context.Context, token string) error
}

// -----------------------------------------------------------------
// Puertos de Servicios de Soporte / Seguridad (Iteración 2)
// -----------------------------------------------------------------

type EncryptionService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}
// -----------------------------------------------------------------
// Puerto de Consultas Analíticas (Iteración 4 - P4 / ADR-17)
// -----------------------------------------------------------------

type AnalyticsRepository interface {
	GetTravelTrends(ctx context.Context) ([]byte, error)
}
