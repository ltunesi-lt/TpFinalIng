package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

type EmailUsecase struct {
	userRepo  domain.UserRepository
	encryptor domain.EncryptionService
}

func NewEmailUsecase(userRepo domain.UserRepository, encryptor domain.EncryptionService) *EmailUsecase {
	return &EmailUsecase{
		userRepo:  userRepo,
		encryptor: encryptor,
	}
}

// US-01: Conectar cuenta de correo vinculando mediante OAuth 2.0 y cifrando tokens (SEC1)
func (uc *EmailUsecase) LinkEmailAccount(ctx context.Context, userID string, rawAccessToken string) error {
	if userID == "" {
		return errors.New("el ID de usuario no puede estar vacío")
	}
	if rawAccessToken == "" {
		return errors.New("el token de acceso no puede estar vacío")
	}

	// Cifrado simétrico AES-256 antes de guardarlo (Decisión C2 / SEC1)
	encryptedToken, err := uc.encryptor.Encrypt(rawAccessToken)
	if err != nil {
		return fmt.Errorf("falló el cifrado seguro del token de correo: %w", err)
	}

	return uc.userRepo.SaveUserMailToken(ctx, userID, encryptedToken)
}

// US-01: Revocar el acceso de manera inmediata
func (uc *EmailUsecase) RevokeEmailAccount(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("el ID de usuario no puede estar vacío")
	}
	return uc.userRepo.RevokeUserMailToken(ctx, userID)
}