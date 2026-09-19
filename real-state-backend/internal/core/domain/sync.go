package domain

import "time"

// SyncMetadata representa los metadatos de una sincronización con Wasi.
type SyncMetadata struct {
	ID                    int        `json:"id"`
	BatchID               string     `json:"batch_id"`                    // UUID del batch
	SyncStartedAt         time.Time  `json:"sync_started_at"`             // Cuándo comenzó
	SyncCompletedAt       *time.Time `json:"sync_completed_at,omitempty"` // Cuándo terminó (NULL si en progreso/falló)
	PropertiesSynced      int        `json:"properties_synced"`           // Propiedades descargadas
	PropertiesDeactivated int        `json:"properties_deactivated"`      // Propiedades marcadas inactivas
	Status                string     `json:"status"`                      // 'IN_PROGRESS', 'COMPLETED', 'FAILED'
	ErrorMessage          *string    `json:"error_message,omitempty"`     // Mensaje de error si aplica
	ExecutedBy            *string    `json:"executed_by,omitempty"`       // Hostname/pod que ejecutó
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// SyncStatus consolida información de freshness para la app móvil
type SyncStatus struct {
	LastSyncTimestamp  *time.Time `json:"last_sync_timestamp,omitempty"` // Cuándo fue el último sync exitoso
	LastSyncBatchID    string     `json:"last_sync_batch_id,omitempty"`  // ID del último batch
	LastSyncStatus     string     `json:"last_sync_status,omitempty"`    // COMPLETED, FAILED, IN_PROGRESS
	IsFresh            bool       `json:"is_fresh"`                      // Datos actualizados dentro de X minutos
	MinutesSinceSync   int        `json:"minutes_since_sync,omitempty"`  // Minutos desde último sync exitoso
	FreshnessThreshold int        `json:"freshness_threshold_minutes"`   // Umbral configurado para "fresh"
	PropertiesCount    int        `json:"properties_count"`              // Total de propiedades activas
	LastErrorMessage   *string    `json:"last_error_message,omitempty"`  // Último error si existe
}
