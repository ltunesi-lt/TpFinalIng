// Servidor de desarrollo con repositorios EN MEMORIA (sin PostgreSQL).
// Sirve para probar la API con scripts/test_api.sh y scripts/demo_chat.sh.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
	"log"
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

	tripUC := usecase.NewTripUsecase(tripRepo)
	chatUC := usecase.NewChatUsecase(tripUC)


	// 1. Endpoint /summary para test_api.sh
	mux := http.NewServeMux()
	// 0. Endpoint raíz para health-check (evita el 404 en la raíz)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"Trotamundos Dev Server"}`))
	})

	// 1. Endpoints de Viajes (/api/trips/.../summary)
	mux.HandleFunc("/api/trips/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/trips/") && len(path) >= 8 && path[len(path)-8:] == "/summary" {
			// Extraer ID del viaje entre /api/trips/ y /summary
			tripID := path[len("/api/trips/") : len(path)-8]
			
			switch r.Method {
			case http.MethodGet:
				summary, err := tripRepo.GetSummaryByID(r.Context(), tripID)
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"viaje no encontrado"}`))
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(summary)
			case http.MethodPut:
				body, err := io.ReadAll(r.Body)
				if err != nil || len(body) == 0 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				var js map[string]interface{}
				if json.Unmarshal(body, &js) != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_ = tripRepo.UpdateSummary(r.Context(), tripID, body)
				w.WriteHeader(http.StatusNoContent)
			case http.MethodPost:
				w.WriteHeader(http.StatusMethodNotAllowed)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		// 2. Endpoint de Enlaces Compartidos (/api/trips/.../share)
		if len(path) > len("/api/trips/") && len(path) >= 6 && path[len(path)-6:] == "/share" {
			tripID := path[len("/api/trips/") : len(path)-6]
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				if json.NewDecoder(r.Body).Decode(&req) != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				hours, _ := req["duration_hours"].(float64)
				if hours <= 0 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				token := "token-abc"
				exp := time.Now().Add(time.Duration(hours) * time.Hour)
				_ = shareRepo.CreateShareLink(r.Context(), token, tripID, exp)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"token":"` + token + `"}`))
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	// 3. Endpoint de Correo de Usuario (/api/users/.../email)
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/api/users/") && len(path) >= 6 && path[len(path)-6:] == "/email" {
			switch r.Method {
			case http.MethodPost:
				var req map[string]string
				if json.NewDecoder(r.Body).Decode(&req) != nil || req["access_token"] == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_ = userRepo.SaveUserMailToken(r.Context(), "user-1", req["access_token"])
				w.WriteHeader(http.StatusNoContent)
			case http.MethodDelete:
				_ = userRepo.RevokeUserMailToken(r.Context(), "user-1")
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// 4. Endpoint de Enlaces Compartidos Globales (/api/shared/...)
	mux.HandleFunc("/api/shared/", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Path[len("/api/shared/"):]
		if token == "" || token == "token-falso" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			tripID, err := shareRepo.ValidateAndGetTripID(r.Context(), token)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			summary, _ := tripRepo.GetSummaryByID(r.Context(), tripID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(summary)
		case http.MethodDelete:
			_ = shareRepo.RevokeShareLink(r.Context(), token)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// 5. Registrar la ruta del chatbot (Iteración 6)
	chatHandler := delivery.NewChatHandler(chatUC)
	mux.HandleFunc("/api/chat", chatHandler.HandleChat)
	port := ":8080"
	log.Printf("==> Dev Server corriendo en http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}