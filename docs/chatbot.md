# 🤖 Manual Técnico del Chatbot de Trotamundos

Este documento describe el funcionamiento interno del módulo conversacional. Para los tipos de mensajes y respuestas esperadas, ver [CASOS_DE_USO_CHATBOT.md](CASOS_DE_USO_CHATBOT.md).

---

## Estructura del módulo

El desarrollo sigue los principios de **Clean Architecture**:

* **`internal/domain/chat.go`**: estructuras de entrada y salida.
  * `ChatRequest { Message string; Lang string }` (`lang`: `es` o `en`).
  * `ChatResponse { Reply string; Escalated bool }` (`escalated` es `true` si pasó a la agencia humana).
* **`internal/usecase/chat_usecase.go`**: lógica de negocio. Reglas por palabras clave (en español e inglés) para escalamiento y estado de viaje, y mensajes multilingües mediante el módulo i18n.
* **`internal/delivery/chat_handler.go`**: controlador HTTP (`HandleChat`).
* **`cmd/devserver/main.go`**: servidor en memoria con un viaje precargado (`trip-123`, destino Buenos Aires).
* **`scripts/demo_chat.sh`**: demo automática (ver README).
* **`scripts/test_api.sh`**: validación usada por `make test-e2e`.

---

## Comportamiento del handler HTTP (`POST /api/chat`)

1. **Límite de tamaño:** el body se limita a **1 MB** (`http.MaxBytesReader`) como defensa ante payloads abusivos.
2. **Decodificación:** si el JSON es inválido (o excede el límite), responde `400` con `JSON inválido`.
3. **Procesamiento:** delega en `ChatUsecase.HandleMessage`.
4. **Errores internos:** el error real se registra en el log del servidor, y al cliente se le devuelve `500` con el mensaje genérico `error interno del servidor` (ADR-32), sin filtrar detalles internos.
5. **Éxito:** responde `200` con `Content-Type: application/json` y el cuerpo `{"reply": "...", "escalated": true|false}`.

---

## Estado del viaje

La respuesta a una consulta de estado **no es un texto fijo**: el caso de uso consulta el viaje y **extrae dinámicamente el destino** (por ejemplo `Buenos Aires`) y su estado (`active` / activo). La demo verifica esto explícitamente en los tests 2 y 4.

Las palabras clave se reconocen en ambos idiomas, por ejemplo `estado de mi viaje` (es) y `trip status` (en).

---

## Pruebas automatizadas de la demo

`scripts/demo_chat.sh` compila el `devserver` a un binario temporal, lo levanta, espera hasta ~7,5 s a que responda, ejecuta 4 escenarios (bienvenida ES, estado con destino real, escalamiento, estado EN) y valida el contenido de cada respuesta. Falla (`exit 1`) en el primer caso incorrecto. Al terminar, detiene el servidor y elimina el binario. Detalle completo en el [README](../README.md).
