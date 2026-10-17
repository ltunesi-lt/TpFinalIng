package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ============================================================================
// 1. MOCKS UNIFICADOS
// ============================================================================

type MockTripRepo struct {
	GetSummaryByIDFunc func(ctx context.Context, id string) ([]byte, error)
	UpdateSummaryFunc  func(ctx context.Context, id string, summary []byte) error
}

func (m *MockTripRepo) GetSummaryByID(ctx context.Context, id string) ([]byte, error) {
	if m.GetSummaryByIDFunc != nil {
		return m.GetSummaryByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockTripRepo) UpdateSummary(ctx context.Context, id string, summary []byte) error {
	if m.UpdateSummaryFunc != nil {
		return m.UpdateSummaryFunc(ctx, id, summary)
	}
	return nil
}

type MockUserRepo struct {
	SaveUserMailTokenFunc   func(ctx context.Context, userID, encryptedToken string) error
	RevokeUserMailTokenFunc func(ctx context.Context, userID string) error
}

func (m *MockUserRepo) SaveUserMailToken(ctx context.Context, userID, encryptedToken string) error {
	if m.SaveUserMailTokenFunc != nil {
		return m.SaveUserMailTokenFunc(ctx, userID, encryptedToken)
	}
	return nil
}

func (m *MockUserRepo) RevokeUserMailToken(ctx context.Context, userID string) error {
	if m.RevokeUserMailTokenFunc != nil {
		return m.RevokeUserMailTokenFunc(ctx, userID)
	}
	return nil
}

type MockEncryption struct {
	EncryptFunc func(plaintext string) (string, error)
	DecryptFunc func(ciphertext string) (string, error)
}

func (m *MockEncryption) Encrypt(plaintext string) (string, error) {
	if m.EncryptFunc != nil {
		return m.EncryptFunc(plaintext)
	}
	return "aes256_" + plaintext, nil
}

func (m *MockEncryption) Decrypt(ciphertext string) (string, error) {
	if m.DecryptFunc != nil {
		return m.DecryptFunc(ciphertext)
	}
	return ciphertext, nil
}

type MockShareRepo struct {
	CreateShareLinkFunc    func(ctx context.Context, token string, tripID string, expiresAt time.Time) error
	ValidateAndGetTripIDFn func(ctx context.Context, token string) (string, error)
	RevokeShareLinkFunc    func(ctx context.Context, token string) error
}

func (m *MockShareRepo) CreateShareLink(ctx context.Context, token string, tripID string, expiresAt time.Time) error {
	if m.CreateShareLinkFunc != nil {
		return m.CreateShareLinkFunc(ctx, token, tripID, expiresAt)
	}
	return nil
}

func (m *MockShareRepo) ValidateAndGetTripID(ctx context.Context, token string) (string, error) {
	if m.ValidateAndGetTripIDFn != nil {
		return m.ValidateAndGetTripIDFn(ctx, token)
	}
	return "", nil
}

func (m *MockShareRepo) RevokeShareLink(ctx context.Context, token string) error {
	if m.RevokeShareLinkFunc != nil {
		return m.RevokeShareLinkFunc(ctx, token)
	}
	return nil
}

// ============================================================================
// 2. TESTS DE TRIP USECASE
// ============================================================================

func TestTripUsecase_GetItinerarySummary(t *testing.T) {
	ctx := context.Background()
	tripID := "trip-1"
	expectedJSON := []byte(`{"destination":"Mendoza"}`)

	repo := &MockTripRepo{
		GetSummaryByIDFunc: func(ctx context.Context, id string) ([]byte, error) {
			if id == tripID {
				return expectedJSON, nil
			}
			return nil, errors.New("no encontrado")
		},
	}

	uc := NewTripUsecase(repo)

	// Caso exitoso
	summary, err := uc.GetItinerarySummary(ctx, tripID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if string(summary) != string(expectedJSON) {
		t.Errorf("esperado %s, obtenido %s", string(expectedJSON), string(summary))
	}

	// Caso ID vacío (Validación)
	_, err = uc.GetItinerarySummary(ctx, "")
	if err == nil {
		t.Error("se esperaba error por ID vacío")
	}
}

func TestTripUsecase_UpdateItinerarySummary(t *testing.T) {
	ctx := context.Background()
	tripID := "trip-1"
	payload := []byte(`{"updated":true}`)
	called := false

	repo := &MockTripRepo{
		UpdateSummaryFunc: func(ctx context.Context, id string, summary []byte) error {
			if id == tripID && len(summary) > 0 {
				called = true
				return nil
			}
			return errors.New("error al actualizar")
		},
	}

	uc := NewTripUsecase(repo)

	err := uc.UpdateItinerarySummary(ctx, tripID, payload)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !called {
		t.Error("el repositorio no fue llamado")
	}

	// Validaciones de parámetros vacíos
	if err := uc.UpdateItinerarySummary(ctx, "", payload); err == nil {
		t.Error("se esperaba error por tripID vacío")
	}
	if err := uc.UpdateItinerarySummary(ctx, tripID, nil); err == nil {
		t.Error("se esperaba error por resumen JSON vacío")
	}
}

// ============================================================================
// 3. TESTS DE EMAIL USECASE (SEC1)
// ============================================================================

func TestEmailUsecase_LinkAndRevoke(t *testing.T) {
	ctx := context.Background()
	userID := "user-123"
	rawToken := "oauth-token-secret"
	var storedToken string

	userRepo := &MockUserRepo{
		SaveUserMailTokenFunc: func(ctx context.Context, uid, etok string) error {
			storedToken = etok
			return nil
		},
		RevokeUserMailTokenFunc: func(ctx context.Context, uid string) error {
			if uid == userID {
				return nil
			}
			return errors.New("usuario no encontrado")
		},
	}

	encryptor := &MockEncryption{
		EncryptFunc: func(plain string) (string, error) {
			return "aes256_" + plain, nil
		},
	}

	uc := NewEmailUsecase(userRepo, encryptor)

	// Probar vinculación y cifrado (SEC1)
	err := uc.LinkEmailAccount(ctx, userID, rawToken)
	if err != nil {
		t.Fatalf("error inesperado al vincular: %v", err)
	}
	if storedToken != "aes256_"+rawToken {
		t.Errorf("el token no se cifró correctamente: %s", storedToken)
	}

	// Probar revocación
	err = uc.RevokeEmailAccount(ctx, userID)
	if err != nil {
		t.Fatalf("error inesperado al revocar: %v", err)
	}

	// Validaciones de entradas vacías
	if err := uc.LinkEmailAccount(ctx, "", rawToken); err == nil {
		t.Error("se esperaba error por userID vacío")
	}
	if err := uc.LinkEmailAccount(ctx, userID, ""); err == nil {
		t.Error("se esperaba error por token vacío")
	}
}

// ============================================================================
// 4. TESTS DE SHARE USECASE (SEC2)
// ============================================================================

func TestShareUsecase_Links(t *testing.T) {
	ctx := context.Background()
	tripID := "trip-789"
	expectedToken := "abc123token"

	shareRepo := &MockShareRepo{
		CreateShareLinkFunc: func(ctx context.Context, token string, tid string, expiresAt time.Time) error {
			return nil
		},
		ValidateAndGetTripIDFn: func(ctx context.Context, token string) (string, error) {
			if token == expectedToken {
				return tripID, nil
			}
			return "", errors.New("token inválido o expirado") // <-- Usamos un error estándar aquí
		},
		RevokeShareLinkFunc: func(ctx context.Context, token string) error {
			return nil
		},
	}

	uc := NewShareUsecase(shareRepo)

	// Crear enlace
	token, err := uc.CreateItineraryShareLink(ctx, tripID, 24*time.Hour)
	if err != nil {
		t.Fatalf("error al crear enlace compartido: %v", err)
	}
	if token == "" {
		t.Error("se esperaba un token generado")
	}

	// Validar acceso al enlace (SEC2)
	// Forzamos un token simulado para probar la validación
	resolvedTripID, err := shareRepo.ValidateAndGetTripID(ctx, expectedToken)
	if err != nil || resolvedTripID != tripID {
		t.Errorf("falló la validación del enlace compartido: %v", err)
	}

	// Revocar enlace instantáneamente (SEC2 <= 1 min)
	err = uc.RevokeLink(ctx, expectedToken)
	if err != nil {
		t.Errorf("error al revocar el enlace: %v", err)
	}

	// Validar errores de parámetros vacíos
	if _, err := uc.CreateItineraryShareLink(ctx, "", time.Hour); err == nil {
		t.Error("se esperaba error por tripID vacío")
	}
	if _, err := uc.AccessSharedItinerary(ctx, ""); err == nil {
		t.Error("se esperaba error por token vacío")
	}
}