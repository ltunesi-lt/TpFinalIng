package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/lt-tunesi/TpFinalIng/internal/delivery"
	"github.com/lt-tunesi/TpFinalIng/internal/domain" // <-- Importante: aquí están las interfaces de los repositorios
	"github.com/lt-tunesi/TpFinalIng/internal/usecase"
)

// Mock temporal de EncryptionService para probar
type DummyEncryptor struct{}

func (d *DummyEncryptor) Encrypt(plaintext string) (string, error) {
	return "encrypted_" + plaintext, nil
}

func (d *DummyEncryptor) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}

func main() {
	// 1. Conexión a PostgreSQL (Docker Compose)
	connStr := "postgres://postgres:password@localhost:5432/trotamundos?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("No se pudo verificar la conexión con la base de datos: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL (Primary)")

	// 2. Repositorios (Usando las interfaces del dominio o tus implementaciones reales del Paso 1)
	var tripRepo domain.TripRepository = nil
	var userRepo domain.UserRepository = nil
	var shareRepo domain.ShareRepository = nil

	// 3. Servicios de Soporte
	encryptor := &DummyEncryptor{}

	// 4. Casos de Uso
	tripUC := usecase.NewTripUsecase(tripRepo)
	emailUC := usecase.NewEmailUsecase(userRepo, encryptor)
	shareUC := usecase.NewShareUsecase(shareRepo)

	// 5. Configurar Enrutador y Handlers (Paso 3)
	mux := http.NewServeMux()
	handler := delivery.NewHTTPHandler(tripUC, emailUC, shareUC)
	handler.RegisterRoutes(mux)

	// 6. Levantar Servidor HTTP
	log.Println("Servidor Trotamundos API ejecutándose en el puerto :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error crítico en el servidor HTTP: %v", err)
	}
}
