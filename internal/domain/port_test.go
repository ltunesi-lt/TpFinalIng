// internal/domain/ports_test.go
package domain

import (
	"testing"
	"time"
)

// TestTaskStatusConstants verifica que las constantes de estado de las tareas sean correctas
func TestTaskStatusConstants(t *testing.T) {
	if TaskPending != "PENDING" {
		t.Errorf("Esperaba PENDING, obtuve %s", TaskPending)
	}
	if TaskProcessing != "PROCESSING" {
		t.Errorf("Esperaba PROCESSING, obtuve %s", TaskProcessing)
	}
	if TaskDone != "DONE" {
		t.Errorf("Esperaba DONE, obtuve %s", TaskDone)
	}
	if TaskFailed != "FAILED" {
		t.Errorf("Esperaba FAILED, obtuve %s", TaskFailed)
	}
}

// TestTripStruct verifica que la estructura Trip se inicialice correctamente con sus campos
func TestTripStruct(t *testing.T) {
	now := time.Now()
	trip := Trip{
		ID:               "trip-123",
		UserID:           "user-456",
		ItinerarySummary: []byte(`{"destination": "Tandil", "activities": ["trail running"]}`),
		UpdatedAt:        now,
	}

	if trip.ID != "trip-123" {
		t.Errorf("Esperaba ID 'trip-123', obtuve %s", trip.ID)
	}
	if trip.UserID != "user-456" {
		t.Errorf("Esperaba UserID 'user-456', obtuve %s", trip.UserID)
	}
	if len(trip.ItinerarySummary) == 0 {
		t.Error("Esperaba que ItinerarySummary contenga datos JSON")
	}
}