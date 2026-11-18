package delivery

import (
	"encoding/json"
	"net/http"
	"github.com/lt-tunesi/TpFinalIng/internal/domain"
	"github.com/lt-tunesi/TpFinalIng/internal/usecase"
)

type ChatHandler struct {
	chatUC *usecase.ChatUsecase
}

func NewChatHandler(chatUC *usecase.ChatUsecase) *ChatHandler {
	return &ChatHandler{chatUC: chatUC}
}

func (h *ChatHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/chat", h.HandleChat)
}

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req domain.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	res, err := h.chatUC.HandleMessage(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}