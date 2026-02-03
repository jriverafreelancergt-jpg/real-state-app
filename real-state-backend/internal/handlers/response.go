package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"real-state-backend/internal/dto"
)

// writeError escribe la respuesta de error estandarizada y añade un trace id
func writeError(w http.ResponseWriter, status int, message, errorCode, tracePrefix string, meta map[string]interface{}) {
	trace := tracePrefix + "-" + uuid.New().String()
	resp := dto.ErrorResponse{
		Status:    "error",
		Code:      status,
		Message:   message,
		ErrorCode: errorCode,
		Meta:      meta,
		TraceID:   trace,
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Trace-ID", trace)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
