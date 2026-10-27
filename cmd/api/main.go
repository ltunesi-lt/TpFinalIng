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
	"github.com/lt-tunesi/TpFinalIng/config/i18n"
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

	// 1. Cargar Módulo i18n desde config/i18n/ (ADR-20)
	translator, err := i18n.NewTranslator("config/i18n", "es")
	if err != nil {
		log.Printf("Advertencia: No se pudieron cargar los archivos i18n (%v). Usando traducciones por defecto.", err)
	} else {
		log.Println("Módulo i18n cargado exitosamente (es, en)")
		log.Printf("i18n Test: %s", translator.Translate("es", "welcome_message"))
	}

	// 2. Conexión a PostgreSQL Primario (Puerto 5432 - Transaccional / Task Table)
	primaryConnStr := "postgres://postgres:password@localhost:5432/trotamundos?sslmode=disable"
	primaryPool, err := pgxpool.New(ctx, primaryConnStr)
	if err != nil {
		log.Fatalf("Error al conectar con PostgreSQL Primario: %v", err)
	}
	defer primaryPool.Close()

	if err := primaryPool.Ping(ctx); err != nil {
		log.Fatalf("No se pudo verificar la conexión con PostgreSQL Primario: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL Primario (Puerto 5432)")

	// 3. Conexión a PostgreSQL Réplica (Puerto 5433 - Analítica Pesada P4 / ADR-17)
	replicaConnStr := "postgres://postgres:password@localhost:5433/trotamundos?sslmode=disable"
	replicaPool, err := pgxpool.New(ctx, replicaConnStr)
	if err != nil {
		log.Printf("Advertencia: No se pudo conectar a la Réplica PostgreSQL (Puerto 5433): %v", err)
	} else {
		if err := replicaPool.Ping(ctx); err != nil {
			log.Printf("Advertencia: Réplica no responde en el puerto 5433, fallback activo.")
		} else {
			log.Println("Conexión exitosa a PostgreSQL Réplica de Analítica (Puerto 5433)")
		}
		defer replicaPool.Close()
	}

	// 4. Repositorios
	taskRepo := repository.NewPostgresTaskRepository(primaryPool)
	var tripRepo domain.TripRepository = nil
	var userRepo domain.UserRepository = nil
	var shareRepo domain.ShareRepository = nil

	// Usar réplica si está disponible; de lo contrario, conectar al primario como fallback
	var analyticsRepo domain.AnalyticsRepository
	if replicaPool != nil {
		analyticsRepo = repository.NewPostgresAnalyticsRepository(replicaPool)
	} else {
		analyticsRepo = repository.NewPostgresAnalyticsRepository(primaryPool)
	}

	// 5. Servicios y Casos de Uso
	encryptor := &DummyEncryptor{}

	tripUC := usecase.NewTripUsecase(tripRepo)
	emailUC := usecase.NewEmailUsecase(userRepo, encryptor)
	shareUC := usecase.NewShareUsecase(shareRepo)
	_ = usecase.NewAnalyticsUsecase(analyticsRepo)

	// 6. Configurar e Iniciar Background Worker
	taskWorker := worker.NewTaskWorker(taskRepo, 3*time.Second, 3)

	taskWorker.RegisterHandler("EMAIL_SCAN", func(ctx context.Context, payload []byte) error {
		log.Printf("[Handler EMAIL_SCAN] Escaneando correo asíncronamente: %s", string(payload))
		return nil
	})

	taskWorker.RegisterHandler("FETCH_PROVIDER_UPDATE", func(ctx context.Context, payload []byte) error {
		log.Printf("[Handler FETCH_PROVIDER] Actualizando datos de proveedor: %s", string(payload))
		return nil
	})

	taskWorker.Start(ctx)

	// 7. Servidor HTTP
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

	// 8. Graceful Shutdown
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