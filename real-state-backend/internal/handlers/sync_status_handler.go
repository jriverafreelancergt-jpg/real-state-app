package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"real-state-backend/internal/core/ports"
)

// SyncStatusHandler maneja los endpoints relacionados con el estado de sincronización
type SyncStatusHandler struct {
	syncStatusProvider ports.SyncStatusProvider
	freshnessThreshold int // Minutos para considerar datos "fresh"
}

// NewSyncStatusHandler crea un nuevo handler
func NewSyncStatusHandler(provider ports.SyncStatusProvider, freshnessThresholdMinutes int) *SyncStatusHandler {
	if freshnessThresholdMinutes <= 0 {
		freshnessThresholdMinutes = 60 // Default: 1 hora
	}
	return &SyncStatusHandler{
		syncStatusProvider: provider,
		freshnessThreshold: freshnessThresholdMinutes,
	}
}

// GetSyncStatus retorna el estado actual de sincronización y freshness de datos
// GET /v1/sync-status
func (h *SyncStatusHandler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.syncStatusProvider.GetSyncStatus(r.Context(), h.freshnessThreshold)
	if err != nil {
		slog.Error("Failed to get sync status", "error", err)
		writeError(w, http.StatusInternalServerError, "Error obteniendo estado de sincronización", "sync_status_error", "sync", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
