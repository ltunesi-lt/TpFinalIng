# 🤖 Casos de Uso e Interacciones del Chatbot de Trotamundos

A continuación se detalla una lista con los principales tipos de mensajes que el usuario puede enviar y cómo responde el asistente en cada escenario.

| Categoría de Intención | Ejemplo de Mensaje del Usuario (`message`) | Respuesta del Chatbot (`reply`) | `escalated` |
| --- | --- | --- | --- |
| **Saludo / Ayuda General** | `"Hola, necesito asistencia"` | "Hola, soy el asistente virtual de Trotamundos. ¿En qué puedo ayudarte con tu viaje? (Consultar estado, reprogramaciones menores, etc.)" | `false` |
| **Consulta de Estado** | `"estado de mi viaje"` | "Tu viaje actual se encuentra activo y con los itinerarios al día en el sistema." | `false` |
| **Caso Complejo / Urgente** | `"Tengo un problema grave y quiero cancelar mi pasaje urgente"` | "Hemos detectado un inconveniente complejo. Tu caso ha sido derivado a nuestra agencia de viajes asociada para una atención personalizada." | `true` |
| **Soporte en Inglés (i18n)** | `"Hello, I need some assistance with my booking"` | Mensaje traducido al idioma configurado (ej. *"Hello, I am Trotamundos virtual assistant..."*) | `false` |
