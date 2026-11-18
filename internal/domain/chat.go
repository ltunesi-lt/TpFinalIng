package domain

// ChatRequest representa el mensaje que envía el usuario.
type ChatRequest struct {
	Message string `json:"message"`
	Lang    string `json:"lang"` // 'es' o 'en'
}

// ChatResponse representa la respuesta del bot y si fue derivado a la agencia.
type ChatResponse struct {
	Reply     string `json:"reply"`
	Escalated bool   `json:"escalated"` // True si pasó a la agencia humana
}