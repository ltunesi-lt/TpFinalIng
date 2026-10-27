package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lt-tunesi/TpFinalIng/internal/domain"
)

// TaskHandler define la función responsable de procesar la lógica de un tipo de tarea
type TaskHandler func(ctx context.Context, payload []byte) error

// TaskWorker gestiona el procesamiento asíncrono en segundo plano
type TaskWorker struct {
	taskRepo    domain.TaskRepository
	handlers    map[string]TaskHandler
	interval    time.Duration
	maxAttempts int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewTaskWorker crea una nueva instancia del worker
func NewTaskWorker(taskRepo domain.TaskRepository, interval time.Duration, maxAttempts int) *TaskWorker {
	return &TaskWorker{
		taskRepo:    taskRepo,
		handlers:    make(map[string]TaskHandler),
		interval:    interval,
		maxAttempts: maxAttempts,
		stopChan:    make(chan struct{}),
	}
}

// RegisterHandler asocia un tipo de tarea (ej: "EMAIL_SCAN", "FETCH_PROVIDER") con su handler
func (w *TaskWorker) RegisterHandler(taskType string, handler TaskHandler) {
	w.handlers[taskType] = handler
}

// Start inicia el bucle de procesamiento en segundo plano (non-blocking)
func (w *TaskWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.loop(ctx)
	log.Printf("[Worker] Background Task Worker iniciado (Sondeo: %v, Max Reintentos: %d)", w.interval, w.maxAttempts)
}

// Stop realiza un Graceful Shutdown del worker esperando que terminen las tareas activas
func (w *TaskWorker) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	log.Println("[Worker] Background Task Worker detenido correctamente.")
}

func (w *TaskWorker) loop(ctx context.Context) {
	defer w.wg.Done()
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] Contexto cancelado, deteniendo worker...")
			return
		case <-w.stopChan:
			log.Println("[Worker] Señal de parada recibida, cerrando worker...")
			return
		case <-ticker.C:
			w.processNextTask(ctx)
		}
	}
}

func (w *TaskWorker) processNextTask(ctx context.Context) {
	// 1. Reclamar la siguiente tarea PENDING usando FOR UPDATE SKIP LOCKED
	task, err := w.taskRepo.ClaimNextPending(ctx)
	if err != nil {
		log.Printf("[Worker] Error al consultar Task Table: %v", err)
		return
	}

	if task == nil {
		// No hay tareas pendientes en este ciclo de polling
		return
	}

	log.Printf("[Worker] Tarea tomada - ID: %s | Tipo: %s | Intento: %d/%d", task.ID, task.Type, task.Attempts, w.maxAttempts)

	// 2. Buscar el handler correspondiente al tipo de tarea
	handler, exists := w.handlers[task.Type]
	if !exists {
		errMsg := fmt.Sprintf("no existe un handler registrado para el tipo de tarea '%s'", task.Type)
		log.Printf("[Worker] Fallo en tarea %s: %s", task.ID, errMsg)
		w.handleFailure(ctx, task, errors.New(errMsg))
		return
	}

	// 3. Ejecución de la lógica de negocio asíncrona
	execErr := handler(ctx, task.Payload)
	if execErr != nil {
		log.Printf("[Worker] Falló la ejecución de la tarea %s: %v", task.ID, execErr)
		w.handleFailure(ctx, task, execErr)
		return
	}

	// 4. Éxito: Pasar estado a DONE
	if err := w.taskRepo.MarkDone(ctx, task.ID); err != nil {
		log.Printf("[Worker] Error al actualizar estado a DONE para tarea %s: %v", task.ID, err)
		return
	}

	log.Printf("[Worker] Tarea %s procesada exitosamente (DONE)", task.ID)
}

func (w *TaskWorker) handleFailure(ctx context.Context, task *domain.Task, execErr error) {
	if task.Attempts >= w.maxAttempts {
		// Superó el límite de reintentos -> FAILED definitivo
		log.Printf("[Worker] Tarea %s excedió los reintentos máximos (%d). Marcando como FAILED.", task.ID, w.maxAttempts)
		if err := w.taskRepo.MarkFailed(ctx, task.ID, execErr.Error()); err != nil {
			log.Printf("[Worker] Error marcando tarea %s como FAILED: %v", task.ID, err)
		}
	} else {
		// Aún tiene reintentos disponibles -> Vuelve a PENDING
		log.Printf("[Worker] Reprogramando tarea %s a PENDING para reintento posterior", task.ID)
		if err := w.taskRepo.MarkRetry(ctx, task.ID); err != nil {
			log.Printf("[Worker] Error devolviendo tarea %s a PENDING: %v", task.ID, err)
		}
	}
}