 Trotamundos MVP

> Sistema backend desarrollado en Go bajo los principios de **Clean Architecture** y **Puertos y Adaptadores**, diseñado para la gestión y sincronización de itinerarios de viajes con un panel de alto rendimiento y procesamiento asíncrono de tareas.

---

## Características Principales

* **Arquitectura Limpia (Clean Architecture):** Desacoplamiento estricto entre dominio, casos de uso, infraestructura y delivery (HTTP).
* **Panel de Viajes Optimizado (P2, P3 / ADR-09):** Almacenamiento y consulta de resúmenes de itinerarios mediante columnas `JSONB` en PostgreSQL para máxima velocidad de respuesta sin *joins* pesados.
* **Control Optimista de Concurrencia (ADR-09):** Mecanismo de control mediante versiones (`version`) para evitar colisiones de datos ante actualizaciones concurrentes de workers o ediciones manuales.
* **Cola de Tareas Asíncronas (ADR-06 / ADR-16):** Procesamiento en segundo plano utilizando `FOR UPDATE SKIP LOCKED` para garantizar la ejecución segura de tareas (`EMAIL_SCAN`, `FETCH_PROVIDER_UPDATE`) sin bloqueo de lecturas.
* **Seguridad y Privacidad (SEC1 & SEC2):** 
  * Cifrado robusto de tokens OAuth de correo electrónico.
  * Generación de enlaces compartidos de itinerarios con soporte para revocación instantánea.
* **Automatización Total:** Gestión completa del ciclo de pruebas y contenedores mediante un `Makefile`.

---

## Tecnologías Utilizadas

* **Lenguaje:** Go (Golang)
* **Base de Datos:** PostgreSQL 15 (con soporte para replicación y contenedores Docker)
* **Driver DB:** `pgx/v5` (con soporte para `DBTX`)
* **Infraestructura:** Docker Compose
* **Automatización:** Make / Bash scripts

---

## Estructura del Proyecto

```text
.
├── cmd/               # Puntos de entrada de la aplicación (servidores, CLI)
├── config/            # Configuraciones globales y traducciones (i18n)
├── internal/          # Lógica interna del negocio (Clean Architecture)
│   ├── delivery/      # Controladores HTTP y rutas
│   ├── domain/        # Entidades de dominio, contratos y puertos
│   ├── repository/    # Adaptadores de base de datos (PostgreSQL / pgx)
│   ├── usecase/       # Casos de uso de la aplicación
│   └── worker/        # Procesadores de tareas en segundo plano
├── scripts/           # Scripts de pruebas E2E (test_api.sh)
├── docker-compose.yaml# Orquestación de bases de datos para desarrollo
├── Makefile           # Automatización de tareas (tests, build, e2e)
└── main.go            # Archivo principal de ejecución

## Ejecucion del MVP
Para ejecutar se debe clonar el repositorio. Para ejeuctar los test:
make test-unit (test unitarios)
make test-coverage (evalua el porcentaje de cobertura del total del proyecto con el MVP)
make test-e2e (comprueba que la API, enrutadores y reglas del negocio se integran y responden)
=======

# 🌍 Trotamundos API — Módulo Asistente Conversacional (Chatbot)

Este proyecto implementa el soporte de primer nivel para la plataforma **Trotamundos** mediante un chatbot conversacional integrado dentro de una arquitectura limpia en Go (`domain`, `usecase`, `delivery`). El sistema gestiona estados, derivaciones complejas a agencias asociadas y soporte multilingüe sin modificar el núcleo existente.

---

El script `scripts/demo_chat.sh` se encarga de todo el flujo:

1. Inicia el `devserver` en segundo plano (log en `/tmp/trotamundos_devserver.log`).
2. Espera hasta 15 segundos a que el servidor responda.
3. Ejecuta los escenarios del chatbot y muestra, para cada uno, el request enviado, el status HTTP y el body recibido.
4. Valida automáticamente que cada respuesta tenga status `200` y el valor esperado del campo `escalated`.
5. Detiene el servidor al terminar (incluso si el script se interrumpe).
6. Imprime un resumen final con la cantidad de casos OK y fallidos.

**Escenarios validados:**

| Escenario | Mensaje | Idioma | `escalated` esperado |
| --- | --- | --- | --- |
| Saludo / consulta general (i18n) | `Hola, necesito asistencia` | `es` | `false` |
| Estado de viaje (dominio) | `estado de mi viaje` | `es` | `false` |
| Caso complejo: derivación a la agencia | `Tengo un problema grave y quiero cancelar mi pasaje urgente` | `es` | `true` |
| Soporte i18n | `Hello, I need help` | `en` | `false` |

**Salida de ejemplo:**

```text
▶ Iniciando devserver en :8080...

── Caso complejo: derivación a la agencia
   → {"message": "Tengo un problema grave y quiero cancelar mi pasaje urgente", "lang": "es"}
   ← [200] {...,"escalated":true}
   ✓ OK

════════════════════════════════
Resultado: 4 OK, 0 fallidos
════════════════════════════════
```

**Código de salida:** el script termina con código `0` si todos los casos pasan y distinto de `0` si alguno falla (o si el servidor no logra iniciar), por lo que también puede usarse en un pipeline de CI.

**Requisitos:** `go`, `curl` y `bash` disponibles en el entorno.

**Variable opcional:** `PORT` define el puerto contra el que el script realiza las peticiones (por defecto `8080`). Debe coincidir con el puerto en el que escucha el `devserver`.

```bash
PORT=8080 ./scripts/demo_chat.sh
```

> Si el script no tiene permisos de ejecución, otorgalos con `chmod +x scripts/demo_chat.sh`.

### Opción B — Pruebas manuales con `curl`

#### 1. Iniciar el servidor localmente

```bash
make run
```

*Esto levantará el servidor de desarrollo en el puerto `:8080` utilizando un viaje precargado de ejemplo (`trip-123`).*

#### 2. Probar el Chatbot

Abrí otra terminal y probá los distintos escenarios del asistente enviando peticiones POST al endpoint `/api/chat`:

**Consulta general o saludo (soporte i18n):**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hola, necesito asistencia", "lang": "es"}'
```

**Consulta de estado de viaje (interacción con dominio):**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "estado de mi viaje", "lang": "es"}'
```

**Caso complejo / derivación automática a la agencia (`escalated: true`):**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Tengo un problema grave y quiero cancelar mi pasaje urgente", "lang": "es"}'
```

---

## 🛠️ Comandos Disponibles en el Makefile

El proyecto cuenta con un `Makefile` optimizado para simplificar la ejecución de pruebas y tareas administrativas:

| Comando | Descripción |
| --- | --- |
| **`make run`** | Levanta el servidor de desarrollo en memoria en el puerto `8080` para pruebas manuales con `curl`. |
| **`make demo`** | Levanta el servidor, ejecuta automáticamente los escenarios del chatbot (`scripts/demo_chat.sh`), valida las respuestas y apaga el servidor al finalizar. |
| **`make test-unit`** | Ejecuta todos los tests unitarios del sistema de forma automatizada. |
| **`make test-e2e`** | Ejecuta las pruebas End-to-End utilizando el servidor en memoria y el script de validación bash (`scripts/test_api.sh`). |
| **`make db-up`** | Levanta los contenedores oficiales de PostgreSQL (Primario y Réplica) mediante Docker Compose (opcional para entornos reales). |
| **`make db-down`** | Detiene los contenedores de Docker de la base de datos. |
| **`make test-coverage`** | Calcula la cobertura de código y genera un reporte visual en HTML (`coverage.html`). |
| **`make clean`** | Limpia los artefactos temporales y reportes de cobertura generados. |

> **Diferencia entre `make demo` y `make test-e2e`:** `demo` está pensado como una demostración rápida y legible del comportamiento del chatbot (muestra request y response de cada escenario), mientras que `test-e2e` valida la API en general mediante `scripts/test_api.sh`.

---

## 📂 Estructura del Módulo del Chatbot

El desarrollo sigue estrictamente los principios de **Clean Architecture**:

* **`internal/domain/chat.go`**: Define las estructuras de datos de entrada y salida (`ChatRequest`, `ChatResponse`).
* **`internal/usecase/chat_usecase.go`**: Contiene la lógica de negocio, reglas basadas en palabras clave (para escalamiento) y soporte multilingüe mediante el módulo i18n.
* **`internal/delivery/chat_handler.go`**: Expone el controlador HTTP que recibe las peticiones JSON y maneja las respuestas del servidor.
* **`cmd/devserver/main.go`**: Servidor en memoria optimizado para pruebas limpias y rápidas.
* **`scripts/demo_chat.sh`**: Script de demostración que levanta el `devserver`, ejecuta los escenarios del chatbot y valida automáticamente las respuestas.
* **`scripts/test_api.sh`**: Script de validación utilizado por `make test-e2e`.
