# Plan de Implementación: Servicio de Sincronización Wasi con Worker Pool y Reconciliación

Este plan detalla el diseño y desarrollo de la lógica de sincronización concurrente utilizando un Worker Pool y la reconciliación en la base de datos (PostgreSQL) usando la técnica "Mark and Sweep".

---

## 🎯 Objetivos

1. **Migración de Base de Datos**: Agregar las columnas `last_sync_id`, `origin` y `status` a la tabla `properties`.
2. **Definir Puertos y Servicios**:
   - Agregar el puerto `PropertySyncService` en `internal/core/ports/wasi.go` o en `ports.go`.
   - Modificar `PropertyRepository` para soportar `UpsertBatch` y `Sweep` (eliminación lógica).
3. **Adaptar el Repositorio de BD**: Implementar `UpsertBatch` y `Sweep` en `internal/repository/property_repository.go`.
4. **Implementar el Servicio de Sincronización (`SyncService`)**:
   - Crear `internal/services/sync_service.go` que maneje el Worker Pool, canal de errores, rate limiting, y persistencia por lotes.
5. **Registrar e Inyectar Dependencias**: Inicializar el nuevo servicio en `cmd/api/main.go`.
6. **Verificación**: Escribir pruebas para el `SyncService` y verificar el flujo de sincronización concurrente.

---

## 📁 Cambios Propuestos

### 1. Base de Datos (Migración)

#### [NEW] [000005_add_wasi_sync_fields.up.sql](file:///home/jrivera/work/real-state-app/real-state-backend/migrations/000005_add_wasi_sync_fields.up.sql)
```sql
ALTER TABLE properties ADD COLUMN last_sync_id VARCHAR(50);
ALTER TABLE properties ADD COLUMN origin VARCHAR(20) DEFAULT 'LOCAL';
ALTER TABLE properties ADD COLUMN status VARCHAR(20) DEFAULT 'ACTIVE';

CREATE INDEX idx_properties_last_sync_id ON properties(last_sync_id);
CREATE INDEX idx_properties_origin ON properties(origin);
CREATE INDEX idx_properties_status ON properties(status);
```

#### [NEW] [000005_add_wasi_sync_fields.down.sql](file:///home/jrivera/work/real-state-app/real-state-backend/migrations/000005_add_wasi_sync_fields.down.sql)
```sql
DROP INDEX IF EXISTS idx_properties_status;
DROP INDEX IF EXISTS idx_properties_origin;
DROP INDEX IF EXISTS idx_properties_last_sync_id;

ALTER TABLE properties DROP COLUMN IF EXISTS status;
ALTER TABLE properties DROP COLUMN IF EXISTS origin;
ALTER TABLE properties DROP COLUMN IF EXISTS last_sync_id;
```

---

### 2. Capa de Puertos (Core)

#### [MODIFY] [wasi.go](file:///home/jrivera/work/real-state-app/real-state-backend/internal/core/ports/wasi.go)
Agregar la interfaz `PropertySyncService` que define las operaciones de sincronización:
```go
type PropertySyncService interface {
    SyncAllProperties(ctx context.Context) error
}
```

#### [MODIFY] [ports.go](file:///home/jrivera/work/real-state-app/real-state-backend/internal/core/ports/ports.go)
Agregar métodos a la interfaz `PropertyRepository` para la reconciliación masiva:
```go
type PropertyRepository interface {
    // ... métodos existentes ...
    UpsertBatch(ctx context.Context, properties []domain.Property, syncBatchID string) error
    Sweep(ctx context.Context, syncBatchID string, origin string) (int64, error)
}
```

---

### 3. Capa de Adaptadores (Repositorio PostgreSQL)

#### [MODIFY] [property_repository.go](file:///home/jrivera/work/real-state-app/real-state-backend/internal/repository/property_repository.go)
Implementar los dos nuevos métodos del repositorio de BD:
- `UpsertBatch`: Inserta o actualiza un lote de propiedades de forma eficiente en un solo query con `ON CONFLICT (id) DO UPDATE`, registrando `last_sync_id` y `origin`.
- `Sweep`: Marca como `INACTIVE` todas las propiedades cuyo origen sea Wasi pero que no fueron incluidas en el lote de sincronización actual (es decir, su `last_sync_id` es diferente del actual).

---

### 4. Capa de Servicios (Orquestación del Worker Pool)

#### [NEW] [sync_service.go](file:///home/jrivera/work/real-state-app/real-state-backend/internal/services/sync_service.go)
Implementar el servicio que orquesta la sincronización concurrente:
- **Generar ID de Lote**: Crear un UUID al iniciar para el "Mark and Sweep".
- **Canal de Trabajos (Jobs)**: Canal que distribuye las páginas a ser procesadas.
- **Worker Pool**: Lanzar `N` workers concurrentes para procesar páginas utilizando la interfaz `WasiProvider`.
- **Canal de Resultados (Results)**: Recibir los slices de propiedades de cada página procesada.
- **Persistencia Concurrente**: Consumir del canal de resultados y persistir en la base de datos usando `UpsertBatch`.
- **Manejo de Errores y Contextos**: Cancelar tareas si el contexto expira o si hay errores críticos repetitivos.
- **Limpieza (Sweep)**: Al finalizar la sincronización exitosamente de todas las páginas, ejecutar `Sweep` para limpiar lógicamente propiedades eliminadas en Wasi.

---

### 5. Configuración e Inicio (`cmd/api/main.go`)

#### [MODIFY] [main.go](file:///home/jrivera/work/real-state-app/real-state-backend/cmd/api/main.go)
- Inicializar el cliente Wasi `wasi.NewClient(...)` y el adaptador `wasi.NewWasiAdapter(...)`.
- Inicializar el servicio de sincronización `services.NewPropertySyncService(...)`.
- *Opcional*: Registrar una ruta administrativa protegida `POST /v1/admin/sync` para forzar la sincronización manualmente si se desea probar.

---

## 🧪 Plan de Verificación

### Pruebas Unitarias
- Escribir `internal/services/sync_service_test.go` para simular una sincronización exitosa utilizando un mock de `WasiProvider` y `PropertyRepository`.
- Verificar que el Worker Pool procesa correctamente todas las páginas.
- Verificar que se ejecutan los métodos `UpsertBatch` y `Sweep` en el repositorio en el orden correcto.

### Ejecución de Pruebas y Aplicación de Migraciones
1. Aplicar la migración de base de datos dentro del contenedor.
2. Ejecutar los tests de todo el proyecto con:
   ```bash
   go test -v ./...
   ```
3. Levantar la aplicación para comprobar que compila correctamente.
