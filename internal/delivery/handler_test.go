package delivery

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------- fakes

type fakeTrip struct {
	summary []byte
	err     error
}

func (f *fakeTrip) GetItinerarySummary(ctx context.Context, id string) ([]byte, error) {
	return f.summary, f.err
}
func (f *fakeTrip) UpdateItinerarySummary(ctx context.Context, id string, s []byte) error {
	return f.err
}

type fakeEmail struct{ err error }

func (f *fakeEmail) LinkEmailAccount(ctx context.Context, u, t string) error { return f.err }
func (f *fakeEmail) RevokeEmailAccount(ctx context.Context, u string) error  { return f.err }

type fakeShare struct {
	tripID string
	err    error
}

func (f *fakeShare) CreateItineraryShareLink(ctx context.Context, id string, d time.Duration) (string, error) {
	return "tok123", f.err
}
func (f *fakeShare) AccessSharedItinerary(ctx context.Context, t string) (string, error) {
	return f.tripID, f.err
}
func (f *fakeShare) RevokeLink(ctx context.Context, t string) error { return f.err }

func do(h *HTTPHandler, method, path, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// ---------------------------------------------------------------- tests

// Sin casos de uso conectados no debe haber panic ni 200.
func TestHandler_GetItinerarySummary_Validation(t *testing.T) {
	rec := do(NewHTTPHandler(nil, nil, nil), "GET", "/api/trips/trip-123/summary", "")
	if rec.Code == http.StatusOK {
		t.Errorf("se esperaba un error, se obtuvo %d", rec.Code)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("esperado 503, obtenido %d", rec.Code)
	}
}

func TestHandler_GetSummary_OK(t *testing.T) {
	h := NewHTTPHandler(&fakeTrip{summary: []byte(`{"destination":"Mendoza"}`)}, nil, nil)
	rec := do(h, "GET", "/api/trips/trip-1/summary", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d", rec.Code)
	}
	if rec.Body.String() != `{"destination":"Mendoza"}` {
		t.Errorf("cuerpo inesperado: %s", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type inesperado: %s", ct)
	}
}

func TestHandler_GetSummary_NotFound(t *testing.T) {
	h := NewHTTPHandler(&fakeTrip{err: errors.New("viaje no encontrado")}, nil, nil)
	rec := do(h, "GET", "/api/trips/x/summary", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obtenido %d", rec.Code)
	}
}

func TestHandler_GetSummary_InternalError(t *testing.T) {
	h := NewHTTPHandler(&fakeTrip{err: errors.New("db caída")}, nil, nil)
	rec := do(h, "GET", "/api/trips/x/summary", "")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("esperado 500, obtenido %d", rec.Code)
	}
}

func TestHandler_UpdateSummary(t *testing.T) {
	h := NewHTTPHandler(&fakeTrip{}, nil, nil)
	if rec := do(h, "PUT", "/api/trips/t1/summary", `{"a":1}`); rec.Code != http.StatusNoContent {
		t.Errorf("esperado 204, obtenido %d", rec.Code)
	}
	if rec := do(h, "PUT", "/api/trips/t1/summary", `no-json`); rec.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d", rec.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := NewHTTPHandler(&fakeTrip{}, nil, nil)
	if rec := do(h, "POST", "/api/trips/t1/summary", ""); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("esperado 405, obtenido %d", rec.Code)
	}
}

func TestHandler_LinkEmail(t *testing.T) {
	h := NewHTTPHandler(nil, &fakeEmail{}, nil)
	if rec := do(h, "POST", "/api/users/u1/email", `{"access_token":"abc"}`); rec.Code != http.StatusNoContent {
		t.Errorf("esperado 204, obtenido %d", rec.Code)
	}
	if rec := do(h, "POST", "/api/users/u1/email", `{"access_token":""}`); rec.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d", rec.Code)
	}
	if rec := do(h, "DELETE", "/api/users/u1/email", ""); rec.Code != http.StatusNoContent {
		t.Errorf("esperado 204, obtenido %d", rec.Code)
	}
}

func TestHandler_Share(t *testing.T) {
	h := NewHTTPHandler(nil, nil, &fakeShare{tripID: "trip-9"})

	rec := do(h, "POST", "/api/trips/trip-9/share", `{"duration_hours":24}`)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), "tok123") {
		t.Errorf("crear enlace: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, "POST", "/api/trips/trip-9/share", `{"duration_hours":0}`); rec.Code != http.StatusBadRequest {
		t.Errorf("esperado 400, obtenido %d", rec.Code)
	}

	rec = do(h, "GET", "/api/shared/tok123", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "trip-9") {
		t.Errorf("acceso: %d %s", rec.Code, rec.Body.String())
	}

	if rec := do(h, "DELETE", "/api/shared/tok123", ""); rec.Code != http.StatusNoContent {
		t.Errorf("esperado 204, obtenido %d", rec.Code)
	}
}

func TestHandler_Share_InvalidToken(t *testing.T) {
	h := NewHTTPHandler(nil, nil, &fakeShare{err: errors.New("token inválido o expirado")})
	if rec := do(h, "GET", "/api/shared/bad", ""); rec.Code != http.StatusNotFound {
		t.Errorf("esperado 404, obtenido %d", rec.Code)
	}
}
