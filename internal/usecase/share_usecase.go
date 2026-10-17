package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

type ShareUsecase struct {
	shareRepo domain.ShareRepository
}

func NewShareUsecase(shareRepo domain.ShareRepository) *ShareUsecase {
	return &ShareUsecase{shareRepo: shareRepo}
}

// Genera un enlace seguro y único para compartir un itinerario
func (uc *ShareUsecase) CreateItineraryShareLink(ctx context.Context, tripID string, duration time.Duration) (string, error) {
	if tripID == "" {
		return "", errors.New("el ID de viaje no puede estar vacío")
	}

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error al generar token aleatorio: %w", err)
	}
	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(duration)

	err := uc.shareRepo.CreateShareLink(ctx, token, tripID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("error al guardar enlace compartido: %w", err)
	}

	return token, nil
}

// Valida si un enlace compartido sigue activo (SEC2: 100% accesos no autorizados bloqueados)
func (uc *ShareUsecase) AccessSharedItinerary(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("el token de acceso no puede estar vacío")
	}
	return uc.shareRepo.ValidateAndGetTripID(ctx, token)
}

// Revoca instantáneamente el enlace de acceso (Efectivo <= 1 min)
func (uc *ShareUsecase) RevokeLink(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("el token de acceso no puede estar vacío")
	}
	return uc.shareRepo.RevokeShareLink(ctx, token)
}