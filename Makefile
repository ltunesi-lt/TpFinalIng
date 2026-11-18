.PHONY: all test-unit db-up db-down test-e2e test-coverage clean run demo

all: test-unit

# Ejecuta el servidor de desarrollo en memoria (Ideal para revisión rápida del profesor)
run:
	@echo "==> Iniciando servidor Trotamundos (Dev Server en memoria)..."
	go run cmd/devserver/main.go

demo:
	@echo "==> Iniciando demostración del chatbot de Trotamundos..."
	@./scripts/demo_chat.sh

# Ejecuta todos los tests unitarios
test-unit:
	@echo "==> Ejecutando tests unitarios del sistema..."
	go test -v ./...

# Levanta la base de datos de desarrollo usando el timeout nativo de pg_isready (sin sleeps)
db-up:
	@echo "==> Levantando bases de datos PostgreSQL con Docker Compose..."
	docker compose up -d
	@echo "==> Esperando a que PostgreSQL esté listo..."
	docker exec trotamundos_db_primary pg_isready -U postgres -d trotamundos -t 30
	@echo "==> ¡Base de datos lista y operativa!"

# Apaga la base de datos
db-down:
	@echo "==> Deteniendo contenedores de Docker..."
	docker compose down

# Ejecuta pruebas End-to-End completas (test_api.sh + demo_chat.sh) con el devserver en memoria
# Ejecuta pruebas End-to-End completas (test_api.sh + demo_chat.sh) con el devserver en memoria
.PHONY: test-e2e
test-e2e:
	@bash scripts/run_e2e.sh


# Cobertura de código
test-coverage:
	@echo "==> Calculando cobertura de código..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "==> Generando reporte visual en HTML (coverage.html)..."
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "==> Limpiando artefactos temporales..."
	rm -f coverage.out coverage.html devserver.log