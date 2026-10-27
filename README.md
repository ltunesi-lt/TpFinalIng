# Trotamundos MVP

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
