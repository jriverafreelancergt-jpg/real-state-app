# Implementación de Mejoras P2: Tracking de Sincronización y Caché Redis

## 📋 Resumen de Cambios

Se han implementado las tres mejoras arquitectónicas recomendadas para resolver las limitaciones del sistema de caché local y sync tracking:

---

## 1. ✅ Tracking Global de Sincronización

### Cambios en Base de Datos

**Archivo:** `migrations/000006_add_sync_tracking.up.sql`

- ✅ **Tabla `sync_metadata`**: Registra cada sincronización con Wasi
  - `batch_id` (UUID): Identificador único del batch
  - `sync_started_at`, `sync_completed_at`: Timestamps de inicio y fin
  - `properties_synced`, `properties_deactivated`: Conteos de operaciones
  - `status`: 'IN_PROGRESS', 'COMPLETED', 'FAILED'
  - `error_message`: Captura errores para auditoría
  - `executed_by`: Hostname/pod para rastrabilidad

- ✅ **Columna `properties.last_successful_sync_at`**: Timestamp individual por propiedad
  
- ✅ **Índices optimizados**: 
  - `idx_sync_metadata_status` para queries rápidas por estado
  - `idx_sync_metadata_completed_at` para obtener último sync
  - `idx_properties_last_sync_at` para filtering por freshness

### Dominio Actualizado

**Archivo:** `internal/core/domain/sync.go`

```go
type SyncMetadata struct {
    BatchID              string
    SyncStartedAt        time.Time
    SyncCompletedAt      *time.Time
    PropertiesSynced     int
    PropertiesDeactivated int
    Status               string
}

type SyncStatus struct {
    LastSyncTimestamp    *time.Time
    IsFresh              bool
    MinutesSinceSync     int
    FreshnessThreshold   int
}
```

---

## 2. ✅ Caché Redis L1 (Opcional)

### Implementación

**Archivo:** `internal/services/property_sync_status_service.go`

**Clase:** `PropertyServiceWithCache`

```go
// Wrapper que cachea ListProperties en Redis
func (ps *PropertyServiceWithCache) ListPropertiesCached(
    ctx context.Context, page, pageSize int
) ([]domain.Property, error)
```

**Características:**
- Cache key: `properties:list:p{page}:ps{pageSize}`
- TTL configurable (default: config.CacheTTL)
- Invalidación automática después de sincronización
- Fallback transparente a BD si Redis no disponible

**Integración con main.go:**
```go
// Redis es opcional pero disponible si se configura
if redisCache != nil {
    // Usar caché L1
}
```

---

## 3. ✅ Endpoint de Sync Status para UX

### Endpoint: `GET /v1/sync-status` 

**Protección:** JWT + RBAC

**Ubicación:** `internal/handlers/sync_status_handler.go`

**Respuesta JSON:**
```json
{
  "last_sync_timestamp": "2026-05-27T20:40:11Z",
  "last_sync_batch_id": "98f33918-b030-4ce1-94db-279bd1099cb9",
  "last_sync_status": "COMPLETED",
  "is_fresh": true,
  "minutes_since_sync": 5,
  "freshness_threshold_minutes": 60,
  "properties_count": 1250,
  "last_error_message": null
}
```

**Casos de uso en app móvil:**
```
- if (is_fresh) { showGreenIndicator("Datos actualizados") }
- else if (minutes_since_sync > 24*60) { showRedWarning("Datos obsoletos") }
- else { showYellowWarning("Última actualización: {minutes_since_sync} min") }
```

---

## 4. 🔧 Métodos Nuevos en Repository

**Archivo:** `internal/repository/property_repository.go`

```go
// Tracking de sincronización
GetLastSyncMetadata(ctx context.Context) (*domain.SyncMetadata, error)
CreateSyncMetadata(ctx context.Context, batchID string, executedBy *string) error
CompleteSyncMetadata(ctx context.Context, batchID string, ...) error
FailSyncMetadata(ctx context.Context, batchID string, errorMessage string) error
GetActivePropertiesCount(ctx context.Context) (int, error)
```

---

## 5. 🔧 Actualización de Sync Service

**Archivo:** `internal/services/sync_service.go`

**Cambios:**
- ✅ Registra inicio con `CreateSyncMetadata()`
- ✅ Cuenta propiedades sincronizadas
- ✅ Marca completado con `CompleteSyncMetadata()` (incluye conteos)
- ✅ Captura errores con `FailSyncMetadata()`
- ✅ Registra `executed_by` (hostname) para auditoría

**Logging mejorado:**
```
"Sincronización con Wasi completada exitosamente"
    batch_id=98f33918...
    propiedades_sincronizadas=1250
    propiedades_desactivadas=5
```

---

## 6. 🛣️ Rutas Registradas

| Ruta | Método | Protección | Descripción |
|------|--------|-----------|------------|
| `/v1/sync-status` | GET | JWT | Estado actual de sincronización + freshness |
| `/sync-status` | GET | JWT | (Backward compatibility) |

---

## 📊 Arquitectura Final

```
App Móvil (Flutter)
    │
    ├─→ GET /v1/sync-status ──→ SyncStatusHandler
    │                          │
    │                          ├─→ GetSyncStatusService
    │                          │   │
    │                          │   ├─→ GetLastSyncMetadata() [BD]
    │                          │   ├─→ GetActivePropertiesCount() [BD]
    │                          │   └─→ Calcula: IsFresh, MinutesSinceSync
    │                          │
    │                          └─→ Response: SyncStatus JSON
    │
    └─→ GET /v1/properties ──→ PropertyHandler
                              │
                              ├─→ PropertyServiceWithCache
                              │   │
                              │   ├─→ Redis GET (cache miss → DB)
                              │   ├─→ PropertyService.ListProperties()
                              │   ├─→ PropertyRepository.GetAll()
                              │   └─→ Redis SET (con TTL)
                              │
                              └─→ Response: Properties[] JSON
```

---

## 🔄 Flujo de Sincronización Mejorado

```
POST /v1/admin/sync
    │
    ├─→ SyncService.SyncAllProperties()
    │
    ├─1. CreateSyncMetadata(batch_id, hostname)
    │    (status='IN_PROGRESS')
    │
    ├─2. Download propiedades de Wasi (worker pool)
    │    → UpsertBatch() → BD
    │
    ├─3. Sweep: desactivar faltantes
    │
    ├─4. SUCCESS:
    │    CompleteSyncMetadata(batch_id, synced_count, deactivated_count)
    │    (status='COMPLETED', sync_completed_at=NOW)
    │
    └─5. ERROR:
         FailSyncMetadata(batch_id, error_message)
         (status='FAILED', sync_completed_at=NOW)
```

---

## 📈 Métricas Capturadas

| Métrica | Ubicación | Uso |
|---------|-----------|-----|
| `sync_metadata.batch_id` | BD | Auditoría + trazabilidad |
| `sync_metadata.sync_completed_at` | BD | Cálculo de freshness |
| `sync_metadata.properties_synced` | BD | Dashboard de administrador |
| `sync_metadata.error_message` | BD | Debugging de fallos |
| `properties.last_sync_id` | BD | Identificar qué batch sincronizó cada propiedad |
| Redis cache hit ratio | Observabilidad | Monitoreo de eficiencia L1 |

---

## ⚙️ Configuración Necesaria

### Variables de Entorno
```bash
# Redis (opcional, pero recomendado en producción)
REDIS_ADDR=redis:6379

# Caché TTL
CACHE_TTL=300  # 5 minutos

# Freshness threshold para UX
SYNC_FRESHNESS_THRESHOLD_MINUTES=60
```

### Database Migrations
```bash
# Ejecutar nueva migración
migrate -path ./migrations -database $DATABASE_URL up

# Para rollback
migrate -path ./migrations -database $DATABASE_URL down
```

---

## ✅ Testing Recomendado

### Unit Tests (Prioridad)
- `TestGetSyncStatus_Fresh()` - Sync reciente
- `TestGetSyncStatus_Stale()` - Sync antiguo
- `TestGetSyncStatus_NeverSynced()` - Primer sync
- `TestPropertyServiceWithCache_CacheHit()`
- `TestPropertyServiceWithCache_CacheMiss()`

### Integration Tests
- `TestSyncAllProperties_TrackingMetadata()` - Verifica sync_metadata
- `TestSyncStatus_Endpoint()` - End-to-end GET /v1/sync-status

---

## 📝 Notas de Implementación

1. **Redis es Opcional**: Si Redis no está disponible, caché es NULL y queries van directo a BD
2. **Backward Compatibility**: Rutas `/sync-status` sin versión funcionan igual
3. **Auditoría Completa**: Cada sync deja rastro en `sync_metadata`
4. **Transacciones Atómicas**: UPSERT + Sweep mantienen consistencia
5. **Observabilidad**: Logs JSON incluyen batch_id para correlacionar eventos

---

## 🚀 Próximos Pasos (P3)

- Agregar endpoint `GET /v1/admin/sync-history?limit=10` para dashboard
- Implementar alertas si `minutes_since_sync > threshold`
- Cache warming en background después de cada sync
- Métricas Prometheus para Redis hit ratio
