// Servidor de desarrollo con repositorios EN MEMORIA (sin PostgreSQL).
// Sirve para probar la API con scripts/test_api.sh. Uso:
//
//	go run ./cmd/devserver
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/lt-tunesi/TpFinalIng/config/i18n"
	"github.com/lt-tunesi/TpFinalIng/internal/delivery"
	"github.com/lt-tunesi/TpFinalIng/internal/usecase"
)

// ---- Trip repo en memoria (con un viaje de ejemplo precargado)
type memTripRepo struct {
	mu sync.RWMutex
	m  map[string][]byte
}

func (r *memTripRepo) GetSummaryByID(ctx context.Context, id string) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.m[id]
	if !ok {
		return nil, errors.New("viaje no encontrado")
	}
	return s, nil
}

func (r *memTripRepo) UpdateSummary(ctx context.Context, id string, summary []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[id] = summary
	return nil
}

// ---- User repo en memoria
type memUserRepo struct {
	mu sync.Mutex
	m  map[string]string
}

func (r *memUserRepo) SaveUserMailToken(ctx context.Context, userID, tok string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[userID] = tok
	return nil
}

func (r *memUserRepo) RevokeUserMailToken(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.m, userID)
	return nil
}

// ---- Share repo en memoria
type shareLink struct {
	tripID  string
	expires time.Time
}

type memShareRepo struct {
	mu sync.Mutex
	m  map[string]shareLink
}

func (r *memShareRepo) CreateShareLink(ctx context.Context, token, tripID string, exp time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.m[token] = shareLink{tripID, exp}
	return nil
}

func (r *memShareRepo) ValidateAndGetTripID(ctx context.Context, token string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	l, ok := r.m[token]
	if !ok || time.Now().After(l.expires) {
		return "", errors.New("token inválido o expirado")
	}
	return l.tripID, nil
}

func (r *memShareRepo) RevokeShareLink(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.m, token)
	return nil
}

type dummyEncryptor struct{}

func (dummyEncryptor) Encrypt(p string) (string, error) { return "encrypted_" + p, nil }
func (dummyEncryptor) Decrypt(c string) (string, error) { return c, nil }

func main() {
	tripRepo := &memTripRepo{m: map[string][]byte{
		"trip-123": []byte(`{"destination":"Buenos Aires","flights":[]}`),
	}}
	userRepo := &memUserRepo{m: map[string]string{}}
	shareRepo := &memShareRepo{m: map[string]shareLink{}}

	// Cargar Módulo i18n (opcional, registrado por si se usa en otros servicios del devserver)
	_, err := i18n.NewTranslator("config/i18n", "es")
	if err != nil {
		log.Printf("Advertencia i18n en devserver: %v", err)
	}

	tripUC := usecase.NewTripUsecase(tripRepo)
	_ = usecase.NewEmailUsecase(userRepo, dummyEncryptor{})
	_ = usecase.NewShareUsecase(shareRepo)

	// Instanciar el enrutador HTTP
	mux := http.NewServeMux()

	// Instanciar el caso de uso del chat usando tripUC correctamente
	chatUsecase := usecase.NewChatUsecase(tripUC)
	chatHandler := delivery.NewChatHandler(chatUsecase)

	// Registrar ruta del chat
	mux.HandleFunc("/api/chat", chatHandler.HandleChat)

	// Leer puerto dinámico o usar 8080 por defecto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Dev server (repos en memoria) en %s — viaje de ejemplo: trip-123", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("error en servidor: %v", err)
	}
}