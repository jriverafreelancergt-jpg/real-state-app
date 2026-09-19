package handlers

import (
	"log/slog"
	"net/http"
	"real-state-backend/internal/core/ports"
)

// SyncHandler maneja las peticiones administrativas de sincronización.
type SyncHandler struct {
	syncService ports.PropertySyncService
}

// NewSyncHandler crea una nueva instancia de SyncHandler.
func NewSyncHandler(s ports.PropertySyncService) *SyncHandler {
	return &SyncHandler{syncService: s}
}

// TriggerSync ejecuta la sincronización de propiedades de Wasi de forma manual.
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	slog.Info("Sincronización manual iniciada por administrador")

	err := h.syncService.SyncAllProperties(r.Context())
	if err != nil {
		slog.Error("Sincronización manual fallida", "error", err)
		writeError(w, http.StatusInternalServerError, "Fallo en la sincronización con Wasi", "sync_failed", "sync", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success","message":"Sincronización de inmuebles finalizada con éxito"}`))
}
