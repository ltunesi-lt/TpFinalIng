# 🤖 Casos de Uso e Interacciones del Chatbot de Trotamundos

A continuación se detalla una lista con los principales tipos de mensajes que el usuario puede enviar y cómo responde el asistente en cada escenario. Todos se envían como `POST /api/chat` con el cuerpo `{"message": "...", "lang": "es|en"}`.

| Categoría de Intención | Ejemplo de Mensaje (`message`) | `lang` | Respuesta del Chatbot (`reply`) | `escalated` |
| --- | --- | --- | --- | --- |
| **Saludo / Ayuda General** | `"Hola, necesito asistencia"` | `es` | "Hola, soy el asistente virtual de Trotamundos. ¿En qué puedo ayudarte con tu viaje? (Consultar estado, reprogramaciones menores, etc.)" | `false` |
| **Consulta de Estado (ES)** | `"quiero ver el estado de mi viaje"` | `es` | Informa el estado del viaje e **incluye el destino real extraído del viaje** (ej. *Buenos Aires*), indicando que está activo y con itinerarios al día. | `false` |
| **Consulta de Estado (EN)** | `"What is my trip status?"` | `en` | Mismo contenido en inglés: incluye el destino (ej. *Buenos Aires*) y el estado `active`. | `false` |
| **Caso Complejo / Urgente** | `"Tengo un problema urgente y quiero cancelar mi pasaje"` | `es` | "Hemos detectado un inconveniente complejo. Tu caso ha sido derivado a nuestra agencia de viajes asociada para una atención personalizada." | `true` |
| **Soporte en Inglés (i18n)** | `"Hello, I need some assistance with my booking"` | `en` | Mensaje traducido al idioma configurado (ej. *"Hello, I am Trotamundos virtual assistant..."*) | `false` |

## Respuestas de error del endpoint

| Situación | Status | Respuesta |
| --- | --- | --- |
| JSON inválido o body mayor a 1 MB | `400` | `JSON inválido` |
| Error interno del servidor | `500` | `error interno del servidor` (el detalle solo queda en el log, ADR-32) |

## Cobertura en la demo automática

`make demo` verifica automáticamente: saludo (ES), estado de viaje (ES y EN) y escalamiento (ver [README](../README.md)).
