-- Agregar columna de tracking de sincronización a propiedades individuales
ALTER TABLE properties
ADD COLUMN last_successful_sync_at TIMESTAMP NULL DEFAULT NULL;

-- Crear tabla de metadatos globales para tracking de sincronización
CREATE TABLE IF NOT EXISTS sync_metadata (
    id SERIAL PRIMARY KEY,
    -- Identificador único del batch de sincronización
    batch_id UUID NOT NULL UNIQUE,
    -- Timestamp del inicio de sincronización
    sync_started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Timestamp de finalización exitosa (NULL si en progreso o falló)
    sync_completed_at TIMESTAMP NULL DEFAULT NULL,
    -- Número de propiedades sincronizadas en este batch
    properties_synced INT NOT NULL DEFAULT 0,
    -- Número de propiedades desactivadas en este batch
    properties_deactivated INT NOT NULL DEFAULT 0,
    -- Estado: 'IN_PROGRESS', 'COMPLETED', 'FAILED'
    status VARCHAR(20) NOT NULL DEFAULT 'IN_PROGRESS',
    -- Mensaje de error si el sync falló
    error_message TEXT NULL,
    -- Hostname/pod que ejecutó la sincronización (para rastrabilidad)
    executed_by VARCHAR(255) NULL,
    -- Timestamps de auditoría
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Índices para optimizar queries
CREATE INDEX idx_sync_metadata_status ON sync_metadata(status);
CREATE INDEX idx_sync_metadata_completed_at ON sync_metadata(sync_completed_at DESC);
CREATE INDEX idx_properties_last_sync_at ON properties(last_successful_sync_at DESC);
CREATE INDEX idx_properties_last_sync_id_origin ON properties(last_sync_id, origin) WHERE status = 'ACTIVE';

-- Trigger para actualizar updated_at en sync_metadata automáticamente
CREATE OR REPLACE FUNCTION update_sync_metadata_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_metadata_updated_at_trigger
BEFORE UPDATE ON sync_metadata
FOR EACH ROW
EXECUTE FUNCTION update_sync_metadata_timestamp();
