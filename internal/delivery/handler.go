package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Interfaces mínimas que necesita la capa de entrega. Los casos de uso
// concretos (*usecase.TripUsecase, etc.) las satisfacen sin cambios.
type TripService interface {
	GetItinerarySummary(ctx context.Context, tripID string) ([]byte, error)
	UpdateItinerarySummary(ctx context.Context, tripID string, summary []byte) error
}

type EmailService interface {
	LinkEmailAccount(ctx context.Context, userID string, rawAccessToken string) error
	RevokeEmailAccount(ctx context.Context, userID string) error
}

type ShareService interface {
	CreateItineraryShareLink(ctx context.Context, tripID string, duration time.Duration) (string, error)
	AccessSharedItinerary(ctx context.Context, token string) (string, error)
	RevokeLink(ctx context.Context, token string) error
}

type HTTPHandler struct {
	trip  TripService
	email EmailService
	share ShareService
}

func NewHTTPHandler(trip TripService, email EmailService, share ShareService) *HTTPHandler {
	return &HTTPHandler{trip: trip, email: email, share: share}
}

// RegisterRoutes usa los patrones con método y {wildcards} de Go 1.22+.
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/trips/{id}/summary", h.getSummary)
	mux.HandleFunc("PUT /api/trips/{id}/summary", h.updateSummary)
	mux.HandleFunc("POST /api/trips/{id}/share", h.createShare)

	mux.HandleFunc("POST /api/users/{id}/email", h.linkEmail)
	mux.HandleFunc("DELETE /api/users/{id}/email", h.revokeEmail)

	mux.HandleFunc("GET /api/shared/{token}", h.accessShared)
	mux.HandleFunc("DELETE /api/shared/{token}", h.revokeShare)
}

// ---------------------------------------------------------------- helpers

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func serviceUnavailable(w http.ResponseWriter) {
	writeError(w, http.StatusServiceUnavailable, "servicio no disponible")
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo JSON inválido")
		return false
	}
	return true
}

func isNotFound(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "no encontrado") || strings.Contains(msg, "inválido") || strings.Contains(msg, "expirado")
}

// ---------------------------------------------------------------- trips

func (h *HTTPHandler) getSummary(w http.ResponseWriter, r *http.Request) {
	if h.trip == nil {
		serviceUnavailable(w)
		return
	}
	summary, err := h.trip.GetItinerarySummary(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "viaje no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(summary) // JSONB crudo, sin re-serializar
}

func (h *HTTPHandler) updateSummary(w http.ResponseWriter, r *http.Request) {
	if h.trip == nil {
		serviceUnavailable(w)
		return
	}
	var raw json.RawMessage
	if !decodeBody(w, r, &raw) {
		return
	}
	if err := h.trip.UpdateItinerarySummary(r.Context(), r.PathValue("id"), raw); err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "viaje no encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------- email (SEC1)

func (h *HTTPHandler) linkEmail(w http.ResponseWriter, r *http.Request) {
	if h.email == nil {
		serviceUnavailable(w)
		return
	}
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.AccessToken == "" {
		writeError(w, http.StatusBadRequest, "access_token es requerido")
		return
	}
	if err := h.email.LinkEmailAccount(r.Context(), r.PathValue("id"), body.AccessToken); err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) revokeEmail(w http.ResponseWriter, r *http.Request) {
	if h.email == nil {
		serviceUnavailable(w)
		return
	}
	if err := h.email.RevokeEmailAccount(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------- share (SEC2)

func (h *HTTPHandler) createShare(w http.ResponseWriter, r *http.Request) {
	if h.share == nil {
		serviceUnavailable(w)
		return
	}
	var body struct {
		DurationHours int `json:"duration_hours"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if body.DurationHours <= 0 {
		writeError(w, http.StatusBadRequest, "duration_hours debe ser mayor a 0")
		return
	}
	token, err := h.share.CreateItineraryShareLink(r.Context(), r.PathValue("id"), time.Duration(body.DurationHours)*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *HTTPHandler) accessShared(w http.ResponseWriter, r *http.Request) {
	if h.share == nil {
		serviceUnavailable(w)
		return
	}
	tripID, err := h.share.AccessSharedItinerary(r.Context(), r.PathValue("token"))
	if err != nil || tripID == "" {
		writeError(w, http.StatusNotFound, "enlace inválido o expirado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"trip_id": tripID})
}

func (h *HTTPHandler) revokeShare(w http.ResponseWriter, r *http.Request) {
	if h.share == nil {
		serviceUnavailable(w)
		return
	}
	if err := h.share.RevokeLink(r.Context(), r.PathValue("token")); err != nil {
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
