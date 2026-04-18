package audit

// ErrorType define el tipo de error para clasificación
type ErrorType string

const (
	// Errores de autenticación y autorización
	ErrorInvalidToken       ErrorType = "INVALID_TOKEN"
	ErrorTokenExpired       ErrorType = "TOKEN_EXPIRED"
	ErrorInvalidCredentials ErrorType = "INVALID_CREDENTIALS"
	ErrorInvalidMFA         ErrorType = "INVALID_MFA"
	ErrorAccountLocked      ErrorType = "ACCOUNT_LOCKED"
	ErrorUnauthorized       ErrorType = "UNAUTHORIZED"
	ErrorForbidden          ErrorType = "FORBIDDEN"
	ErrorSessionRevoked     ErrorType = "SESSION_REVOKED"
	ErrorDeviceMismatch     ErrorType = "DEVICE_MISMATCH"
	ErrorUserNotFound       ErrorType = "USER_NOT_FOUND"
	ErrorSessionInvalid     ErrorType = "SESSION_INVALID"
	ErrorInvalidJTI         ErrorType = "INVALID_JTI"

	// Errores de validación (no auditables por defecto)
	ErrorValidation   ErrorType = "VALIDATION_ERROR"
	ErrorJSONParse    ErrorType = "JSON_PARSE_ERROR"
	ErrorMissingField ErrorType = "MISSING_FIELD"

	// Errores técnicos (no auditables)
	ErrorDatabase ErrorType = "DATABASE_ERROR"
	ErrorInternal ErrorType = "INTERNAL_ERROR"
)

// AuditableError representa un error que debe ser registrado en auditoría
type AuditableError struct {
	Code       ErrorType
	Message    string
	StatusCode int
	Details    map[string]interface{}
}

// Error implementa la interfaz error
func (e *AuditableError) Error() string {
	return e.Message
}

// IsAuditableError verifica si un error es auditable
func IsAuditableError(err error) bool {
	if err == nil {
		return false
	}

	// Si es un AuditableError, siempre es auditable
	if _, ok := err.(*AuditableError); ok {
		return true
	}

	// Códigos de error específicos que son auditables
	auditableCodes := map[string]bool{
		"invalid_credentials": true,
		"invalid_token":       true,
		"token_expired":       true,
		"session_invalid":     true,
		"invalid_mfa":         true,
		"unauthorized":        true,
		"forbidden":           true,
		"device_mismatch":     true,
		"account_locked":      true,
	}

	// Buscar en el mensaje si contiene códigos auditables
	errStr := err.Error()
	for code := range auditableCodes {
		if code == errStr {
			return true
		}
	}

	return false
}

// NewAuditableError crea un nuevo error auditable
func NewAuditableError(code ErrorType, message string, statusCode int) *AuditableError {
	return &AuditableError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetails agrega detalles al error auditable
func (e *AuditableError) WithDetails(key string, value interface{}) *AuditableError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Constructores convenientes para errores comunes

// ErrInvalidToken error de token inválido
func ErrInvalidToken(reason string) *AuditableError {
	e := NewAuditableError(ErrorInvalidToken, "Invalid token", 401)
	return e.WithDetails("reason", reason)
}

// ErrTokenExpired error de token expirado
func ErrTokenExpired() *AuditableError {
	return NewAuditableError(ErrorTokenExpired, "Token expired", 401)
}

// ErrInvalidCredentials error de credenciales inválidas
func ErrInvalidCredentials() *AuditableError {
	return NewAuditableError(ErrorInvalidCredentials, "Invalid credentials", 401)
}

// ErrInvalidMFA error de MFA inválido
func ErrInvalidMFA() *AuditableError {
	return NewAuditableError(ErrorInvalidMFA, "Invalid MFA code", 401)
}

// ErrAccountLocked error de cuenta bloqueada
func ErrAccountLocked() *AuditableError {
	return NewAuditableError(ErrorAccountLocked, "Account locked", 401)
}

// ErrUnauthorized error de no autorizado
func ErrUnauthorized() *AuditableError {
	return NewAuditableError(ErrorUnauthorized, "Unauthorized", 401)
}

// ErrForbidden error de acceso prohibido
func ErrForbidden() *AuditableError {
	return NewAuditableError(ErrorForbidden, "Forbidden", 403)
}

// ErrSessionRevoked error de sesión revocada
func ErrSessionRevoked() *AuditableError {
	return NewAuditableError(ErrorSessionRevoked, "Session revoked", 401)
}

// ErrDeviceMismatch error de dispositivo no coincide
func ErrDeviceMismatch() *AuditableError {
	return NewAuditableError(ErrorDeviceMismatch, "Device mismatch detected", 401).
		WithDetails("severity", "high")
}

// ErrUserNotFound error de usuario no encontrado (auditable cuando es auth)
func ErrUserNotFound() *AuditableError {
	return NewAuditableError(ErrorUserNotFound, "User not found", 404)
}

// ErrSessionInvalid error de sesión inválida
func ErrSessionInvalid() *AuditableError {
	return NewAuditableError(ErrorSessionInvalid, "Session invalid or expired", 401)
}

// ErrInvalidJTI error de JTI inválido
func ErrInvalidJTI() *AuditableError {
	return NewAuditableError(ErrorInvalidJTI, "Invalid JTI", 401)
}
