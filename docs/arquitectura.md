 # Trotamundos MVP — Arquitectura

> Sistema backend desarrollado en Go bajo los principios de **Clean Architecture** y **Puertos y Adaptadores**, diseñado para la gestión y sincronización de itinerarios de viajes con un panel de alto rendimiento y procesamiento asíncrono de tareas. Incluye un módulo de asistente conversacional (chatbot) con derivación a agencias asociadas y soporte multilingüe.

---

## Características Principales

* **Arquitectura Limpia (Clean Architecture):** desacoplamiento estricto entre dominio, casos de uso, infraestructura y delivery (HTTP).
* **Panel de Viajes Optimizado (P2, P3 / ADR-09):** almacenamiento y consulta de resúmenes de itinerarios mediante columnas `JSONB` en PostgreSQL, sin *joins* pesados.
* **Control Optimista de Concurrencia (ADR-09):** control por versiones (`version`) para evitar colisiones ante actualizaciones concurrentes de workers o ediciones manuales.
* **Cola de Tareas Asíncronas (ADR-06 / ADR-16):** procesamiento en segundo plano con `FOR UPDATE SKIP LOCKED` para ejecutar tareas (`EMAIL_SCAN`, `FETCH_PROVIDER_UPDATE`) sin bloquear lecturas.
* **Seguridad y Privacidad (SEC1 & SEC2):**
  * Cifrado robusto de tokens OAuth de correo electrónico.
  * Enlaces compartidos de itinerarios con revocación instantánea.
* **Manejo seguro de errores (ADR-32):** los errores internos se registran en el servidor y al cliente se le devuelve un mensaje genérico, sin exponer detalles.
* **Defensa ante payloads abusivos:** el endpoint del chatbot limita el body a 1 MB.
* **Automatización total:** gestión del ciclo de pruebas y contenedores mediante un `Makefile`.

---

## Tecnologías Utilizadas

* **Lenguaje:** Go (Golang)
* **Base de datos:** PostgreSQL 15 (con soporte para replicación y contenedores Docker)
* **Driver DB:** `pgx/v5` (con soporte para `DBTX`)
* **Infraestructura:** Docker Compose
* **Automatización:** Make / Bash scripts

---

## Estructura del Proyecto

```text
.
├── cmd/               # Puntos de entrada (servidores, CLI). Incluye devserver (en memoria)
├── config/            # Configuraciones globales y traducciones (i18n)
├── docs/              # Documentación (arquitectura, manual y casos de uso del chatbot)
├── internal/          # Lógica interna del negocio (Clean Architecture)
│   ├── delivery/      # Controladores HTTP y rutas (incluye chat_handler.go)
│   ├── domain/        # Entidades de dominio, contratos y puertos (incluye chat.go)
│   ├── repository/    # Adaptadores de base de datos (PostgreSQL / pgx)
│   ├── usecase/       # Casos de uso (incluye chat_usecase.go)
│   └── worker/        # Procesadores de tareas en segundo plano
├── scripts/           # Scripts de pruebas: test_api.sh (E2E) y demo_chat.sh (demo del chatbot)
├── docker-compose.yaml# Orquestación de bases de datos para desarrollo
├── Makefile           # Automatización de tareas (tests, build, e2e, demo)
└── main.go            # Archivo principal de ejecución
```

---

## Módulo del Chatbot

Integrado sin modificar el núcleo existente, siguiendo las mismas capas:

| Capa | Archivo | Responsabilidad |
| --- | --- | --- |
| Domain | `internal/domain/chat.go` | `ChatRequest` y `ChatResponse`. |
| Usecase | `internal/usecase/chat_usecase.go` | Reglas por palabras clave (es/en), estado del viaje con destino dinámico, escalamiento e i18n. |
| Delivery | `internal/delivery/chat_handler.go` | `POST /api/chat`: límite de 1 MB, errores `400`/`500` y respuesta JSON. |

Detalle en el [Manual del Chatbot](chatbot.md).

---

## Ejecución del MVP

Para ejecutar, clonar el repositorio. Pruebas disponibles:

* `make test-unit`: tests unitarios.
* `make test-coverage`: porcentaje de cobertura del proyecto (reporte `coverage.html`).
* `make test-e2e`: verifica que la API, los enrutadores y las reglas de negocio se integran y responden.
* `make demo`: demostración automática del chatbot.

Guía completa en el [README](../README.md).
