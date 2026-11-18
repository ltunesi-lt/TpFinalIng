package delivery

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lt-tunesi/TpFinalIng/internal/domain"
	"github.com/lt-tunesi/TpFinalIng/internal/usecase"
)

type ChatHandler struct {
	chatUC *usecase.ChatUsecase
}

func NewChatHandler(chatUC *usecase.ChatUsecase) *ChatHandler {
	return &ChatHandler{
		chatUC: chatUC,
	}
}

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Límite defensivo de tamaño en el body (1 MB)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req domain.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	res, err := h.chatUC.HandleMessage(r.Context(), req)
	if err != nil {
		// Logueamos internamente el error real, pero devolvemos un genérico al cliente (ADR-32)
		log.Printf("error interno en chat usecase: %v", err)
		http.Error(w, "error interno del servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("error codificando respuesta de chat: %v", err)
	}
}