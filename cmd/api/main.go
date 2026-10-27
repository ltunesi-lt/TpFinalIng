package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lt-tunesi/TpFinalIng/internal/delivery"
	"github.com/lt-tunesi/TpFinalIng/internal/domain"
	"github.com/lt-tunesi/TpFinalIng/internal/repository"
	"github.com/lt-tunesi/TpFinalIng/internal/usecase"
	"github.com/lt-tunesi/TpFinalIng/internal/worker"
)

type DummyEncryptor struct{}

func (d *DummyEncryptor) Encrypt(plaintext string) (string, error) {
	return "encrypted_" + plaintext, nil
}

func (d *DummyEncryptor) Decrypt(ciphertext string) (string, error) {
	return ciphertext, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Conexión a PostgreSQL usando el pool nativo de pgx/v5
	connStr := "postgres://postgres:password@localhost:5432/trotamundos?sslmode=disable"
	dbPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Error al conectar con la base de datos: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("No se pudo verificar la conexión con la base de datos: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL (pgxpool)")

	// 2. Repositorios (inyectando dbPool que sí implementa DBTX)
	taskRepo := repository.NewPostgresTaskRepository(dbPool)
	var tripRepo domain.TripRepository = nil
	var userRepo domain.UserRepository = nil
	var shareRepo domain.ShareRepository = nil

	// 3. Servicios
	encryptor := &DummyEncryptor{}

	// 4. Casos de Uso
	tripUC := usecase.NewTripUsecase(tripRepo)
	emailUC := usecase.NewEmailUsecase(userRepo, encryptor)
	shareUC := usecase.NewShareUsecase(shareRepo)

	// 5. Configurar e Iniciar Background Task Worker
	taskWorker := worker.NewTaskWorker(taskRepo, 3*time.Second, 3)

	taskWorker.RegisterHandler("EMAIL_SCAN", func(ctx context.Context, payload []byte) error {
		log.Printf("[Handler EMAIL_SCAN] Escaneando correo asíncronamente con payload: %s", string(payload))
		return nil
	})

	taskWorker.RegisterHandler("FETCH_PROVIDER_UPDATE", func(ctx context.Context, payload []byte) error {
		log.Printf("[Handler FETCH_PROVIDER] Actualizando datos de proveedor externo: %s", string(payload))
		return nil
	})

	taskWorker.Start(ctx)

	// 6. Servidor HTTP
	mux := http.NewServeMux()
	handler := delivery.NewHTTPHandler(tripUC, emailUC, shareUC)
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Servidor Trotamundos API ejecutándose en puerto :8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en el servidor HTTP: %v", err)
		}
	}()

	// 7. Captura de señales para Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("\nIniciando apagado controlado del sistema (Graceful Shutdown)...")

	taskWorker.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error al cerrar el servidor HTTP: %v", err)
	}

	log.Println("Sistema Trotamundos detenido correctamente.")
}