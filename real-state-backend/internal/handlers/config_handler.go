package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
)

type ConfigHandler struct {
	configRepo ports.SecurityConfigRepository
	auditRepo  ports.AuditRepository // Para registrar cambios en configuración
}

// NewConfigHandler crea un handler con repositorios necesarios
func NewConfigHandler(configRepo ports.SecurityConfigRepository, auditRepo ports.AuditRepository) *ConfigHandler {
	return &ConfigHandler{configRepo: configRepo, auditRepo: auditRepo}
}

// GetSecurityConfig obtiene todas las configuraciones de seguridad
// Busca en la BD cada key. Si falla la lectura, devuelve un valor por defecto
// y registra un warning para facilitar el diagnóstico.
func (h *ConfigHandler) GetSecurityConfig(w http.ResponseWriter, r *http.Request) {
	// Valores por defecto razonables (coinciden con defaults en config.LoadConfig)
	configs := map[string]interface{}{
		"ACCESS_TOKEN_TTL_MINUTES": 15, // minutos
		"REFRESH_TOKEN_TTL_DAYS":   7,  // días
		"MAX_FAILED_ATTEMPTS":      5,
		"LOCKOUT_DURATION_MINUTES": 15, // minutos
	}

	// Obtener cada config desde la BD; si hay error, mantener el default y loguear
	if val, err := h.configRepo.GetInt(r.Context(), "ACCESS_TOKEN_TTL_MINUTES"); err == nil {
		configs["ACCESS_TOKEN_TTL_MINUTES"] = val
	} else {
		slog.Warn("Failed to read ACCESS_TOKEN_TTL_MINUTES from DB, using default", "error", err)
	}
	if val, err := h.configRepo.GetInt(r.Context(), "REFRESH_TOKEN_TTL_DAYS"); err == nil {
		configs["REFRESH_TOKEN_TTL_DAYS"] = val
	} else {
		slog.Warn("Failed to read REFRESH_TOKEN_TTL_DAYS from DB, using default", "error", err)
	}
	if val, err := h.configRepo.GetInt(r.Context(), "MAX_FAILED_ATTEMPTS"); err == nil {
		configs["MAX_FAILED_ATTEMPTS"] = val
	} else {
		slog.Warn("Failed to read MAX_FAILED_ATTEMPTS from DB, using default", "error", err)
	}
	if val, err := h.configRepo.GetInt(r.Context(), "LOCKOUT_DURATION_MINUTES"); err == nil {
		configs["LOCKOUT_DURATION_MINUTES"] = val
	} else {
		slog.Warn("Failed to read LOCKOUT_DURATION_MINUTES from DB, using default", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configs)
}

// UpdateSecurityConfig actualiza una configuración con validaciones y audit
// Acepta tanto query params (?key=...&value=...) como JSON en el body:
// { "key": "ACCESS_TOKEN_TTL_MINUTES", "value": 15 }
func (h *ConfigHandler) UpdateSecurityConfig(w http.ResponseWriter, r *http.Request) {
	// Primero intentar decodificar JSON si Content-Type es application/json
	var key string
	var v int
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var body struct {
			Key   string `json:"key"`
			Value *int   `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error": "Invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		if body.Key == "" || body.Value == nil {
			http.Error(w, `{"error": "Missing key or value in body"}`, http.StatusBadRequest)
			return
		}
		key = body.Key
		v = *body.Value
	} else {
		// Fallback a query params por compatibilidad
		key = r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" || value == "" {
			http.Error(w, `{"error": "Missing key or value"}`, http.StatusBadRequest)
			return
		}
		var err error
		v, err = strconv.Atoi(value)
		if err != nil {
			http.Error(w, `{"error": "Value must be numeric"}`, http.StatusBadRequest)
			return
		}
	}

	// Validar clave permitida
	allowed := map[string]struct{}{
		"ACCESS_TOKEN_TTL_MINUTES": {},
		"REFRESH_TOKEN_TTL_DAYS":   {},
		"MAX_FAILED_ATTEMPTS":      {},
		"LOCKOUT_DURATION_MINUTES": {},
	}
	if _, ok := allowed[key]; !ok {
		http.Error(w, `{"error": "Invalid config key"}`, http.StatusBadRequest)
		return
	}

	// Validaciones por rango según key
	switch key {
	case "ACCESS_TOKEN_TTL_MINUTES":
		if v < 1 || v > 60 {
			http.Error(w, `{"error": "ACCESS_TOKEN_TTL_MINUTES must be between 1 and 60"}`, http.StatusBadRequest)
			return
		}
	case "REFRESH_TOKEN_TTL_DAYS":
		if v < 1 || v > 365 {
			http.Error(w, `{"error": "REFRESH_TOKEN_TTL_DAYS must be between 1 and 365"}`, http.StatusBadRequest)
			return
		}
	case "MAX_FAILED_ATTEMPTS":
		if v < 1 || v > 20 {
			http.Error(w, `{"error": "MAX_FAILED_ATTEMPTS must be between 1 and 20"}`, http.StatusBadRequest)
			return
		}
	case "LOCKOUT_DURATION_MINUTES":
		if v < 1 || v > 1440 {
			http.Error(w, `{"error": "LOCKOUT_DURATION_MINUTES must be between 1 and 1440"}`, http.StatusBadRequest)
			return
		}
	}

	// Obtener valores antiguos para auditoría (intentar leer, si falla usar nil)
	oldValPtr := interface{}(nil)
	if ov, err := h.configRepo.GetInt(r.Context(), key); err == nil {
		oldValPtr = ov
	}

	// Intentar actualizar en BD
	if err := h.configRepo.UpdateConfig(r.Context(), key, strconv.Itoa(v)); err != nil {
		slog.Error("Failed to update config", "error", err)
		http.Error(w, `{"error": "Failed to update config"}`, http.StatusInternalServerError)
		return
	}

	// Registrar auditoría (si auditRepo está disponible)
	if h.auditRepo != nil {
		userIDPtr := (*string)(nil)
		if uid := r.Context().Value("user_id"); uid != nil {
			if s, ok := uid.(string); ok {
				userIDPtr = &s
			}
		}
		log := &domain.AuditLog{
			EventType: "config_change",
			UserID:    userIDPtr,
			Resource:  "security_config",
			Action:    "update",
			OldValues: map[string]interface{}{key: oldValPtr},
			NewValues: map[string]interface{}{key: v},
			IPAddress: r.RemoteAddr,
			UserAgent: r.Header.Get("User-Agent"),
			Timestamp: time.Now(),
		}
		if err := h.auditRepo.LogEvent(r.Context(), log); err != nil {
			slog.Warn("Failed to log audit event", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Config updated successfully",
		"key":     key,
		"value":   strconv.Itoa(v),
	})
}
