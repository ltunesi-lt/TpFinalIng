# 🌍 Trotamundos API - Mock up

Este proyecto implementa el soporte de primer nivel para la plataforma **Trotamundos** mediante un chatbot conversacional integrado dentro de una arquitectura limpia en Go (`domain`, `usecase`, `delivery`). El sistema gestiona estados de viaje, derivaciones complejas a agencias asociadas y soporte multilingüe (`es` / `en`) sin modificar el núcleo existente.

---

## 🚀 Guía de Ejecución Rápida

El proyecto incluye un entorno de desarrollo en memoria (`devserver`) que permite ejecutar y probar toda la API y el chatbot de manera instantánea, **sin necesidad de levantar bases de datos PostgreSQL ni contenedores Docker** para las pruebas básicas.

### 1. Clonar el repositorio

```bash
git clone <url-del-repositorio>
cd <carpeta-del-proyecto>
```
### 2.0 Prueba de la funcionalidad + chatbot
make test-e2e

### 2.1 Prueba de funcionalidad API
En una terminal: make run
bAbrir otra terminal y poner: bash scripts/test_api.sh

### 2. 2. Elegir cómo probar el chatbot

| Opción | Comando | Cuándo usarla |
| --- | --- | --- |
| **Demo automática** | `make demo` | Recomendada. Compila y levanta el servidor, ejecuta todos los escenarios, valida las respuestas y apaga el servidor sin intervención manual. |
| **Servidor manual + `curl`** | `make run` | Para explorar la API a mano, enviando tus propios mensajes. |

### Opción A — Demo automática (recomendada)

```bash
make demo
```

O directamente:

```bash
PORT=8080 ./scripts/demo_chat.sh
```

> Si el script no tiene permisos de ejecución: `chmod +x scripts/demo_chat.sh`.

El script `scripts/demo_chat.sh` realiza el siguiente flujo:

1. **Libera el puerto** si está ocupado (usa `lsof` / `fuser`; si no están disponibles, continúa igual).
2. **Compila** el `devserver` a un binario temporal (`/tmp/devserver_test_bin`).
3. **Inicia el binario** en segundo plano en el puerto configurado.
4. **Espera a que el servidor responda** (hasta 15 reintentos cada 0,5 s, aprox. 7,5 segundos).
5. **Ejecuta y verifica los escenarios** del chatbot mostrando `OK` o `FALLÓ` por cada uno.
6. **Limpia al terminar**: detiene el servidor y borra el binario temporal, incluso si el script se interrumpe (`trap` en `EXIT`).

**Escenarios validados:**

| # | Escenario | Mensaje | Idioma | Se valida que la respuesta... |
| --- | --- | --- | --- | --- |
| 1 | Bienvenida | `Hola, necesito asistencia` | `es` | contenga `asistente virtual` |
| 2 | Estado de viaje (destino real) | `quiero ver el estado de mi viaje` | `es` | contenga el destino `Buenos Aires` (extracción dinámica desde el viaje precargado) |
| 3 | Caso complejo: derivación a la agencia | `Tengo un problema urgente y quiero cancelar mi pasaje` | `es` | contenga `"escalated":true` |
| 4 | Soporte i18n con keywords en inglés | `What is my trip status?` | `en` | contenga `Buenos Aires` y `active` |

**Salida de ejemplo:**

```text
==> Iniciando demostración del chatbot de Trotamundos...
▶ Compilando devserver...
▶ Iniciando binario en el puerto :8080...
⏳ Esperando a que el servidor esté listo...
==> Ejecutando y verificando casos de prueba del chat...
Test 1 (Bienvenida ES): OK
Test 2 (Estado del Viaje con destino real): OK
Test 3 (Escalamiento): OK
Test 4 (Soporte i18n EN con keywords en inglés): OK
✔ ¡Demostración y pruebas del chatbot validadas de forma impecable!
```

**Código de salida:** el script termina con `0` si todos los casos pasan. Ante el **primer** caso que falle (o si el servidor no levanta a tiempo) imprime `FALLÓ` / `❌` y termina con código `1`, por lo que también puede usarse en un pipeline de CI.

**Requisitos:** `go`, `curl` y `bash`. Opcionalmente `lsof` y `fuser` para liberar el puerto automáticamente.

**Variable opcional:** `PORT` define el puerto del servidor y de las peticiones (por defecto `8080`).

### Opción B — Pruebas manuales con `curl`

#### 1. Iniciar el servidor localmente

```bash
make run
```

*Levanta el servidor de desarrollo en el puerto `:8080` con un viaje precargado de ejemplo (`trip-123`, destino Buenos Aires).*

#### 2. Probar el chatbot

Desde otra terminal, enviá peticiones `POST` al endpoint `/api/chat`:

**Saludo / consulta general (i18n):**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hola, necesito asistencia", "lang": "es"}'
```

**Consulta de estado de viaje (interacción con dominio):**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "quiero ver el estado de mi viaje", "lang": "es"}'
```

**Estado de viaje en inglés:**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "What is my trip status?", "lang": "en"}'
```

**Caso complejo / derivación automática a la agencia (`escalated: true`):**

```bash
curl -s -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Tengo un problema urgente y quiero cancelar mi pasaje", "lang": "es"}'
```

#### Contrato del endpoint `POST /api/chat`

| Campo | Dirección | Tipo | Descripción |
| --- | --- | --- | --- |
| `message` | request | `string` | Mensaje del usuario. |
| `lang` | request | `string` | Idioma: `es` o `en`. |
| `reply` | response | `string` | Respuesta del asistente. |
| `escalated` | response | `bool` | `true` si el caso fue derivado a la agencia. |

| Status | Cuándo |
| --- | --- |
| `200` | Mensaje procesado correctamente. |
| `400` | Body con JSON inválido (`"JSON inválido"`) o que supera el límite de 1 MB. |
| `500` | Error interno. El detalle se registra en el log del servidor y al cliente solo se devuelve `"error interno del servidor"` (ADR-32). |

---

## 🛠️ Comandos Disponibles en el Makefile

| Comando | Descripción |
| --- | --- |
| **`make run`** | Levanta el servidor de desarrollo en memoria en el puerto `8080` para pruebas manuales con `curl`. |
| **`make demo`** | Compila y levanta el servidor, ejecuta los escenarios del chatbot (`scripts/demo_chat.sh`), valida las respuestas y apaga el servidor al finalizar. |
| **`make test-unit`** | Ejecuta todos los tests unitarios. |
| **`make test-e2e`** | Ejecuta las pruebas End-to-End con el servidor en memoria y el script `scripts/test_api.sh`. |
| **`make test-coverage`** | Calcula la cobertura de código y genera un reporte HTML (`coverage.html`). |
| **`make db-up`** | Levanta PostgreSQL (Primario y Réplica) con Docker Compose (opcional, para entornos reales). |
| **`make db-down`** | Detiene los contenedores de la base de datos. |
| **`make clean`** | Limpia artefactos temporales y reportes de cobertura. |

> **Diferencia entre `make demo` y `make test-e2e`:** `demo` es una demostración rápida y legible del comportamiento del chatbot; `test-e2e` valida la API en general mediante `scripts/test_api.sh`.

---

## 📚 Documentación

* [Manual del Chatbot](docs/chatbot.md): detalle técnico del módulo.
* [Casos de uso del Chatbot](docs/CASOS_DE_USO_CHATBOT.md): tipos de mensajes y respuestas.
* [Arquitectura del Sistema](docs/arquitectura.md): diseño, tecnologías y estructura del proyecto.
