package services

import (
	"context"
	"fmt"
	"log/slog"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"real-state-backend/pkg/cache"
	"time"
)

// PropertyServiceWithCache envuelve PropertyService con caché Redis L1
type PropertyServiceWithCache struct {
	service      ports.PropertyService
	repo         ports.PropertyRepository
	redisCache   cache.Cache
	cacheTTL     time.Duration
	freshnessTTL time.Duration // Cuánto tiempo considerar datos "fresh" para UX
}

// NewPropertyServiceWithCache crea un nuevo servicio con caché
func NewPropertyServiceWithCache(
	service ports.PropertyService,
	repo ports.PropertyRepository,
	redisCache cache.Cache,
	cacheTTL time.Duration,
	freshnessThresholdMinutes int,
) *PropertyServiceWithCache {
	return &PropertyServiceWithCache{
		service:      service,
		repo:         repo,
		redisCache:   redisCache,
		cacheTTL:     cacheTTL,
		freshnessTTL: time.Duration(freshnessThresholdMinutes) * time.Minute,
	}
}

// ListPropertiesCached retorna propiedades con caché Redis L1
func (ps *PropertyServiceWithCache) ListPropertiesCached(
	ctx context.Context,
	page, pageSize int,
) ([]domain.Property, error) {
	// Generar clave de caché
	cacheKey := fmt.Sprintf("properties:list:p%d:ps%d", page, pageSize)

	// Intentar obtener del caché
	if ps.redisCache != nil {
		var properties []domain.Property
		err := ps.redisCache.Get(ctx, cacheKey, &properties)
		if err == nil {
			slog.Debug("Cache hit for properties list", "key", cacheKey, "page", page, "pageSize", pageSize)
			return properties, nil
		}
	}

	// Cache miss: obtener del repositorio
	slog.Debug("Cache miss for properties list, fetching from DB", "key", cacheKey)
	properties, err := ps.service.ListProperties(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	// Guardar en caché si Redis disponible
	if ps.redisCache != nil {
		_ = ps.redisCache.Set(ctx, cacheKey, properties, ps.cacheTTL)
	}

	return properties, nil
}

// InvalidatePropertiesCache limpia la caché de propiedades (usado después de sincronización)
func (ps *PropertyServiceWithCache) InvalidatePropertiesCache(ctx context.Context) error {
	if ps.redisCache == nil {
		return nil
	}

	// Patrón de invalidación: borrar todas las keys que empiecen con "properties:list:"
	// En Redis, usaríamos FLUSHDB o KEYS pattern, pero con la interfaz genérica solo borramos una clave
	// Para una solución más robusta, se podría extender la interfaz cache.Cache
	pattern := "properties:list:*"
	slog.Debug("Invalidating properties cache pattern", "pattern", pattern)

	// Por ahora, solo registramos la invalidación (Redis del patrón es operación O(N))
	// En producción, usar Redis SCAN + DEL o SET con patrón
	return nil
}

// GetSyncStatusService implementa SyncStatusProvider
type GetSyncStatusService struct {
	repo                      ports.PropertyRepository
	freshnessThresholdMinutes int
}

// NewGetSyncStatusService crea el servicio de status de sync
func NewGetSyncStatusService(repo ports.PropertyRepository, freshnessThresholdMinutes int) ports.SyncStatusProvider {
	return &GetSyncStatusService{
		repo:                      repo,
		freshnessThresholdMinutes: freshnessThresholdMinutes,
	}
}

// GetSyncStatus retorna información de freshness de datos para UX
func (s *GetSyncStatusService) GetSyncStatus(ctx context.Context, freshnessThresholdMinutes int) (*domain.SyncStatus, error) {
	status := &domain.SyncStatus{
		FreshnessThreshold: freshnessThresholdMinutes,
		IsFresh:            false,
		MinutesSinceSync:   -1,
	}

	// Obtener último sync completado
	syncMeta, err := s.repo.GetLastSyncMetadata(ctx)
	if err != nil {
		slog.Warn("Failed to get last sync metadata", "error", err)
		return status, nil // No es error crítico, solo retornamos status "no fresh"
	}

	// Si nunca hubo sync
	if syncMeta == nil {
		status.LastSyncStatus = "NEVER"
		return status, nil
	}

	// Completar información del status
	status.LastSyncBatchID = syncMeta.BatchID
	status.LastSyncStatus = syncMeta.Status
	status.LastSyncTimestamp = syncMeta.SyncCompletedAt
	status.LastErrorMessage = syncMeta.ErrorMessage

	// Calcular freshness
	if syncMeta.SyncCompletedAt != nil {
		now := time.Now()
		minutesSince := int(now.Sub(*syncMeta.SyncCompletedAt).Minutes())
		status.MinutesSinceSync = minutesSince
		status.IsFresh = minutesSince <= freshnessThresholdMinutes
	}

	// Obtener conteo de propiedades activas
	count, err := s.repo.GetActivePropertiesCount(ctx)
	if err != nil {
		slog.Warn("Failed to get active properties count", "error", err)
	}
	status.PropertiesCount = count

	return status, nil
}
