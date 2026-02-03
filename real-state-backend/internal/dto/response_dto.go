package dto

// ErrorResponse estandariza las respuestas de error
type ErrorResponse struct {
	Status    string                 `json:"status"`
	Code      int                    `json:"code"`
	Message   string                 `json:"message"`
	ErrorCode string                 `json:"error_code"`
	Meta      map[string]interface{} `json:"meta"`
	TraceID   string                 `json:"trace_id"`
}
