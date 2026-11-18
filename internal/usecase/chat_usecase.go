package usecase

import (
	"context"
	"strings"
	"github.com/lt-tunesi/TpFinalIng/config/i18n"
	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

type ChatUsecase struct {
	tripUC     *TripUsecase
	translator *i18n.Translator
}

func NewChatUsecase(tripUC *TripUsecase, translator *i18n.Translator) *ChatUsecase {
	return &ChatUsecase{
		tripUC:     tripUC,
		translator: translator,
	}
}

func (uc *ChatUsecase) HandleMessage(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	lang := req.Lang
	if lang == "" {
		lang = "es" // por defecto español
	}

	msg := strings.ToLower(req.Message)

	// 1. Detección de casos complejos (Derivación a la agencia de viajes preferida)
	complexKeywords := []string{"cancelar", "cancelación", "reembolso", "problema grave", "error de cobro", "disputa", "urgente"}
	for _, keyword := range complexKeywords {
		if strings.Contains(msg, keyword) {
			replyText := "Hemos detectado un inconveniente complejo. Tu caso ha sido derivado a nuestra agencia de viajes asociada para una atención personalizada."
			if lang == "en" {
				replyText = "We have detected a complex issue. Your case has been escalated to our preferred travel agency for personalized support."
			}
			return domain.ChatResponse{
				Reply:     replyText,
				Escalated: true,
			}, nil
		}
	}

	// 2. Consulta de Estado o Itinerario (Primer nivel interactivo)
	if strings.Contains(msg, "estado") || strings.Contains(msg, "viaje") || strings.Contains(msg, "itinerario") {
		// Podrías consultar un viaje por defecto o invocar tripUC si se pasa un ID
		replyText := "Tu viaje actual se encuentra activo y con los itinerarios al día en el sistema."
		if lang == "en" {
			replyText = "Your current trip is active and your itineraries are up to date in the system."
		}
		return domain.ChatResponse{
			Reply:     replyText,
			Escalated: false,
		}, nil
	}

	// 3. Respuesta por defecto (Ayuda general usando i18n o mensajes estándar)
	// 3. Respuesta por defecto (Ayuda general usando i18n o mensajes estándar)
	defaultReply := uc.translator.Translate(lang, "welcome_message")
	
	// Usamos defaultReply en lugar de escribir el texto hardcodeado
	return domain.ChatResponse{
		Reply:     defaultReply,
		Escalated: false,
	}, nil
}