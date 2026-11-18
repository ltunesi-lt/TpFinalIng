package usecase

import (
	"context"
	"encoding/json"
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
		lang = "es"
	}

	msgLower := strings.ToLower(req.Message)

	// 1. Detección de escalamiento / problemas graves
	if strings.Contains(msgLower, "cancelar") || strings.Contains(msgLower, "problema") || strings.Contains(msgLower, "urgente") || strings.Contains(msgLower, "grave") {
		var replyText string
		if lang == "en" {
			replyText = "We have detected a complex issue. Your case has been forwarded to our partner travel agency for personalized assistance."
		} else {
			replyText = "Hemos detectado un inconveniente complejo. Tu caso ha sido derivado a nuestra agencia de viajes asociada para una atención personalizada."
		}
		return domain.ChatResponse{
			Reply:     replyText,
			Escalated: true,
		}, nil
	}

	// 2. Consulta de estado del viaje (conectado al TripUsecase real)
	if strings.Contains(msgLower, "estado") || strings.Contains(msgLower, "viaje") {
		summaryBytes, err := uc.tripUC.GetItinerarySummary(ctx, "trip-123")
		var replyText string

		if err != nil {
			if lang == "en" {
				replyText = "We could not retrieve your current trip status at this moment. Please check your dashboard."
			} else {
				replyText = "No pudimos recuperar el estado actual de tu viaje en este momento. Por favor, revisá tu panel."
			}
		} else {
			var tripData map[string]interface{}
			if json.Unmarshal(summaryBytes, &tripData) == nil {
				if dest, ok := tripData["destination"].(string); ok {
					if lang == "en" {
						replyText = "Your trip to " + dest + " is active and your itineraries are up to date."
					} else {
						replyText = "Tu viaje a " + dest + " se encuentra activo y con los itinerarios al día en el sistema."
					}
				} else {
					if lang == "en" {
						replyText = "Your current trip is active and your itineraries are up to date."
					} else {
						replyText = "Tu viaje actual se encuentra activo y con los itinerarios al día en el sistema."
					}
				}
			} else {
				if lang == "en" {
					replyText = "Your current trip is active and your itineraries are up to date."
				} else {
					replyText = "Tu viaje actual se encuentra activo y con los itinerarios al día en el sistema."
				}
			}
		}

		return domain.ChatResponse{
			Reply:     replyText,
			Escalated: false,
		}, nil
	}

	// 3. Respuesta por defecto / bienvenida
	var defaultReply string
	if lang == "en" {
		defaultReply = "Hello, I am Trotamundos virtual assistant. How can I help you today?"
	} else {
		defaultReply = "Hola, soy el asistente virtual de Trotamundos. ¿En qué puedo ayudarte?"
	}

	return domain.ChatResponse{
		Reply:     defaultReply,
		Escalated: false,
	}, nil
}