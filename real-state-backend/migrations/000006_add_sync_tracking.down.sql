-- Revertir cambios de sincronización

-- Eliminar triggers
DROP TRIGGER IF EXISTS sync_metadata_updated_at_trigger ON sync_metadata;
DROP FUNCTION IF EXISTS update_sync_metadata_timestamp();

-- Eliminar índices
DROP INDEX IF EXISTS idx_sync_metadata_status;
DROP INDEX IF EXISTS idx_sync_metadata_completed_at;
DROP INDEX IF EXISTS idx_properties_last_sync_at;
DROP INDEX IF EXISTS idx_properties_last_sync_id_origin;

-- Eliminar tabla de metadatos
DROP TABLE IF EXISTS sync_metadata;

-- Remover columna de tracking
ALTER TABLE properties
DROP COLUMN IF EXISTS last_successful_sync_at;
