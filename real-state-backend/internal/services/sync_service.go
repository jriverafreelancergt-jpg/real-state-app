package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"sync"

	"github.com/google/uuid"
)

// PropertySyncService implementa el puerto ports.PropertySyncService.
type PropertySyncService struct {
	wasiProvider ports.WasiProvider
	repo         ports.PropertyRepository
	numWorkers   int
}

// Asegurar que PropertySyncService implementa la interfaz en tiempo de compilación.
var _ ports.PropertySyncService = (*PropertySyncService)(nil)

// NewPropertySyncService crea una instancia del servicio de sincronización.
func NewPropertySyncService(wasiProvider ports.WasiProvider, repo ports.PropertyRepository, numWorkers int) *PropertySyncService {
	if numWorkers <= 0 {
		numWorkers = 5 // Fallback por defecto
	}
	return &PropertySyncService{
		wasiProvider: wasiProvider,
		repo:         repo,
		numWorkers:   numWorkers,
	}
}

// SyncAllProperties descarga concurrentemente todas las propiedades de Wasi y las reconcilia en base de datos.
func (s *PropertySyncService) SyncAllProperties(ctx context.Context) error {
	syncBatchID := uuid.New().String()
	slog.Info("Iniciando sincronización con Wasi", "batch_id", syncBatchID)

	// 0. Registrar inicio de sincronización en metadatos
	executedBy := os.Getenv("HOSTNAME")
	if err := s.repo.CreateSyncMetadata(ctx, syncBatchID, &executedBy); err != nil {
		slog.Warn("Failed to create sync metadata record", "batch_id", syncBatchID, "error", err)
		// No es error crítico, continuamos con sync
	}

	// 1. Obtener la primera página para determinar la cantidad de páginas totales
	firstBatch, totalPages, err := s.wasiProvider.FetchProperties(ctx, 1)
	if err != nil {
		errMsg := fmt.Sprintf("error al obtener la primera página de Wasi: %v", err)
		_ = s.repo.FailSyncMetadata(ctx, syncBatchID, errMsg)
		return errors.New(errMsg)
	}

	if len(firstBatch) == 0 || totalPages == 0 {
		slog.Info("No se encontraron propiedades en Wasi para sincronizar")
		_ = s.repo.CompleteSyncMetadata(ctx, syncBatchID, 0, 0)
		return nil
	}

	slog.Info("Páginas totales encontradas en Wasi", "total_pages", totalPages)

	// Canales de control
	jobs := make(chan int, totalPages)
	results := make(chan []domain.Property, totalPages)
	errs := make(chan error, totalPages)

	// Cancelar dinámicamente si un worker falla
	workerCtx, cancelWorker := context.WithCancel(ctx)
	defer cancelWorker()

	var wg sync.WaitGroup

	// 2. Iniciar Worker Pool
	for w := 0; w < s.numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for page := range jobs {
				// Comprobar si el contexto ya fue cancelado
				if workerCtx.Err() != nil {
					return
				}

				slog.Debug("Worker descargando página", "worker", workerID, "page", page)
				props, _, err := s.wasiProvider.FetchProperties(workerCtx, page)
				if err != nil {
					errs <- fmt.Errorf("worker %d falló en página %d: %w", workerID, page, err)
					cancelWorker() // Cancelar el resto de workers ante falla crítica
					return
				}

				select {
				case <-workerCtx.Done():
					return
				case results <- props:
				}
			}
		}(w)
	}

	// 3. Encolar el resto de páginas (de la 2 en adelante)
	for p := 2; p <= totalPages; p++ {
		jobs <- p
	}
	close(jobs)

	// 4. Goroutine para esperar que los workers terminen y cerrar canales
	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	// 5. Insertar la primera página que ya tenemos en memoria
	if err := s.repo.UpsertBatch(ctx, firstBatch, syncBatchID); err != nil {
		errMsg := fmt.Sprintf("error al guardar lote inicial: %v", err)
		_ = s.repo.FailSyncMetadata(ctx, syncBatchID, errMsg)
		return errors.New(errMsg)
	}

	// 6. Recibir y guardar resultados a medida que se completen las descargas
	var syncErr error
	propertiesSyncedCount := len(firstBatch) // Contar primera página
	for {
		select {
		case props, ok := <-results:
			if !ok {
				results = nil
			} else if len(props) > 0 && syncErr == nil {
				if err := s.repo.UpsertBatch(ctx, props, syncBatchID); err != nil {
					syncErr = fmt.Errorf("error guardando lote en BD: %w", err)
					cancelWorker()
				} else {
					propertiesSyncedCount += len(props)
				}
			}
		case err, ok := <-errs:
			if !ok {
				errs = nil
			} else if err != nil {
				syncErr = err
			}
		}

		if results == nil && errs == nil {
			break
		}
	}

	// 7. Si la sincronización falló, abortamos para no borrar registros reales ("Sweep" de reconciliación)
	if syncErr != nil {
		errMsg := fmt.Sprintf("sincronización fallida o interrumpida. Se omitirá el barrido (Sweep): %v", syncErr)
		_ = s.repo.FailSyncMetadata(ctx, syncBatchID, errMsg)
		return errors.New(errMsg)
	}

	// 8. Fase de Reconciliación ("Sweep"): Desactivar inmuebles que ya no existen en Wasi
	rowsAffected, err := s.repo.Sweep(ctx, syncBatchID, "WASI")
	if err != nil {
		errMsg := fmt.Sprintf("error durante la reconciliación (Sweep): %v", err)
		_ = s.repo.FailSyncMetadata(ctx, syncBatchID, errMsg)
		return errors.New(errMsg)
	}

	// 9. Marcar sincronización como completada en metadatos
	if err := s.repo.CompleteSyncMetadata(ctx, syncBatchID, propertiesSyncedCount, int(rowsAffected)); err != nil {
		slog.Warn("Failed to complete sync metadata record", "batch_id", syncBatchID, "error", err)
		// No es error crítico, el sync fue exitoso
	}

	slog.Info("Sincronización con Wasi completada exitosamente",
		"batch_id", syncBatchID,
		"propiedades_sincronizadas", propertiesSyncedCount,
		"propiedades_desactivadas", rowsAffected,
	)

	return nil
}
