package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

type TripUsecase struct {
	tripRepo domain.TripRepository
}

// NewTripUsecase instancia el caso de uso inyectando el puerto del dominio
func NewTripUsecase(tripRepo domain.TripRepository) *TripUsecase {
	return &TripUsecase{
		tripRepo: tripRepo,
	}
}

// GetItinerarySummary obtiene el resumen JSON de un viaje optimizado para el panel (P2, P3 / Iteración 3)
func (uc *TripUsecase) GetItinerarySummary(ctx context.Context, tripID string) ([]byte, error) {
	if tripID == "" {
		return nil, errors.New("el ID de viaje no puede estar vacío")
	}

	summary, err := uc.tripRepo.GetSummaryByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener resumen del viaje: %w", err)
	}

	return summary, nil
}

// UpdateItinerarySummary actualiza el resumen JSON precalculado del viaje tras cambios o nuevas reservas
func (uc *TripUsecase) UpdateItinerarySummary(ctx context.Context, tripID string, summary []byte) error {
	if tripID == "" {
		return errors.New("el ID de viaje es requerido")
	}
	if len(summary) == 0 {
		return errors.New("el resumen JSON no puede estar vacío")
	}

	err := uc.tripRepo.UpdateSummary(ctx, tripID, summary)
	if err != nil {
		return fmt.Errorf("error al actualizar resumen en el caso de uso: %w", err)
	}

	return nil
}