# Guía de Errores Auditables

## Descripción General

Este sistema implementa una estrategia de auditoría diferenciada para errores de seguridad. Los errores se clasifican en tres niveles:

1. **Errores Auditables** → Se registran en `audit_logs` + logs técnicos
2. **Errores No Auditables** → Solo se registran en logs técnicos (slog)
3. **Errores de Validación** → Se descartan (no son relevantes para seguridad)

## Arquitectura

### Capas Implicadas

```
HTTP Handler (auth_handler.go)
    ↓
    └─→ handleAuthError() [audita + responde]
        ├─→ AuditableError → audit_logs ✅
        └─→ Otros → slog warning ⚠️
    ↓
Auth Service (auth_service.go)
    ↓
    └─→ Lanza AuditableError específicos
        ├─→ Login() → ErrInvalidCredentials()
        ├─→ RefreshToken() → ErrTokenExpired()
        └─→ LogAuditError() [registra en BD]
    ↓
Audit Repository (audit_repository.go)
    ↓
    └─→ LogEvent() con JSONB
        └─→ new_values: { error_code, error_message, status_code, details }
```

## Tipos de Errores Auditables

Están definidos en `pkg/audit/errors.go`:

### Errores de Autenticación

```go
// Token inválido o malformado
audit.ErrInvalidToken("reason")
// Resultado en audit_logs:
// {
//   "error_code": "INVALID_TOKEN",
//   "error_message": "Invalid token",
//   "status_code": 401,
//   "reason": "..."
// }

// Token expirado
audit.ErrTokenExpired()

// Credenciales incorrectas
audit.ErrInvalidCredentials()

// Código MFA inválido
audit.ErrInvalidMFA()

// Cuenta bloqueada
audit.ErrAccountLocked()
```

### Errores de Sesión

```go
// Sesión revocada
audit.ErrSessionRevoked()

// Dispositivo no coincide
audit.ErrDeviceMismatch()

// JTI inválido
audit.ErrInvalidJTI()

// Sesión inválida/expirada
audit.ErrSessionInvalid()
```

### Errores de Autorización

```go
// No autorizado (genérico)
audit.ErrUnauthorized()

// Acceso prohibido (403)
audit.ErrForbidden()
```

## Uso en Servicios

### En `auth_service.go`

**Antes (sin auditoría):**
```go
func (s *AuthService) Login(...) (dto.LoginResponseDTO, error) {
    user, err := s.userRepo.GetByUsername(ctx, req.Username)
    if err != nil {
        return dto.LoginResponseDTO{}, errors.New("invalid credentials")
    }
    // ...
}
```

**Después (con auditoría):**
```go
func (s *AuthService) Login(...) (dto.LoginResponseDTO, error) {
    user, err := s.userRepo.GetByUsername(ctx, req.Username)
    if err != nil {
        // Crear error auditable
        auditErr := audit.ErrInvalidCredentials()
        // Registrarlo en BD
        s.LogAuditError(ctx, "LOGIN_FAILURE", auditErr, nil)
        // Devolverlo al handler
        return dto.LoginResponseDTO{}, auditErr
    }
    // ...
}
```

## Uso en Handlers

### En `auth_handler.go`

**Método helper para auditar automáticamente:**
```go
func (h *AuthHandler) handleAuthError(w http.ResponseWriter, r *http.Request, 
    err error, eventType string, userID *string) {
    
    if auditErr, ok := err.(*audit.AuditableError); ok {
        // Log técnico
        slog.Error("auth_error",
            "event_type", eventType,
            "error_code", auditErr.Code,
        )
        
        // Log de auditoría en BD
        h.service.LogAuditError(r.Context(), eventType, err, userID)
        
        // Respuesta al cliente (sin detalles técnicos)
        writeError(w, auditErr.StatusCode, auditErr.Message, string(auditErr.Code), "auth", nil)
        return
    }
    
    // Error no auditable
    slog.Warn("non_auditable_error", "error", err.Error())
    writeError(w, http.StatusInternalServerError, "Internal server error", "internal_error", "auth", nil)
}
```

**En endpoint de login:**
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ... validación ...
    
    resp, err := h.service.Login(r.Context(), req, ...)
    if err != nil {
        // Manejo unificado con auditoría automática
        h.handleAuthError(w, r, err, "LOGIN_FAILURE", nil)
        return
    }
    
    // ... éxito ...
}
```

## Datos Almacenados en `audit_logs`

### Registro Exitoso (LOGIN_SUCCESS)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "event_type": "LOGIN_SUCCESS",
  "user_id": "user-123",
  "resource": "auth",
  "action": "login",
  "old_values": null,
  "new_values": {
    "mfa_required": false
  },
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "timestamp": "2026-02-15T10:30:00Z"
}
```

### Registro de Error (LOGIN_FAILURE)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "event_type": "LOGIN_FAILURE",
  "user_id": null,
  "resource": "auth",
  "action": "error_occurred",
  "old_values": null,
  "new_values": {
    "error_code": "INVALID_CREDENTIALS",
    "error_message": "Invalid credentials",
    "status_code": 401
  },
  "ip_address": null,
  "user_agent": null,
  "timestamp": "2026-02-15T10:29:55Z"
}
```

### Registro de Intento de Hijacking (SESSION_HIJACK_ATTEMPT)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "event_type": "SESSION_HIJACK_ATTEMPT",
  "user_id": "user-123",
  "resource": "auth",
  "action": "error_occurred",
  "old_values": null,
  "new_values": {
    "error_code": "DEVICE_MISMATCH",
    "error_message": "Device mismatch detected",
    "status_code": 401,
    "severity": "high"
  },
  "ip_address": "192.168.1.200",
  "user_agent": "Mozilla/5.0...",
  "timestamp": "2026-02-15T10:28:00Z"
}
```

## Patrones de Búsqueda en BD

### Encontrar todos los intentos fallidos de login
```sql
SELECT * FROM audit_logs 
WHERE event_type = 'LOGIN_FAILURE' 
  AND new_values->>'error_code' = 'INVALID_CREDENTIALS'
ORDER BY timestamp DESC
LIMIT 50;
```

### Detectar intentos de hijacking de sesión
```sql
SELECT * FROM audit_logs 
WHERE new_values->>'error_code' = 'DEVICE_MISMATCH'
ORDER BY timestamp DESC;
```

### Intentos de tokens inválidos
```sql
SELECT 
  timestamp,
  user_id,
  new_values->>'error_message' as error_msg,
  COUNT(*) as attempts
FROM audit_logs 
WHERE new_values->>'error_code' = 'INVALID_TOKEN'
GROUP BY user_id, timestamp::date
HAVING COUNT(*) > 5
ORDER BY timestamp DESC;
```

### Timeline de seguridad para un usuario
```sql
SELECT timestamp, event_type, new_values->>'error_code' as error_code
FROM audit_logs 
WHERE user_id = 'user-123'
ORDER BY timestamp DESC
LIMIT 20;
```

## Flujo de Errores Paso a Paso

### Ejemplo: Login con credenciales inválidas

**1. Handler recibe request**
```
POST /api/auth/login
{ "username": "admin", "password": "wrong" }
```

**2. Handler valida DTO**
```go
if err := req.Validate(); err != nil {
    writeError(w, http.StatusBadRequest, err.Error(), "validation_error", ...)
    return // No audita (validación básica)
}
```

**3. Handler llama al servicio**
```go
resp, err := h.service.Login(r.Context(), req, ...)
```

**4. Servicio verifica usuario**
```go
user, err := s.userRepo.GetByUsername(ctx, req.Username)
if err != nil {
    auditErr := audit.ErrInvalidCredentials()
    s.LogAuditError(ctx, "LOGIN_FAILURE", auditErr, nil)
    return dto.LoginResponseDTO{}, auditErr
}
```

**5. Handler captura error auditable**
```go
if err != nil {
    h.handleAuthError(w, r, err, "LOGIN_FAILURE", nil)
    return
}
```

**6. Handler audita automáticamente**
- Log técnico: `slog.Error("auth_error", "error_code", "INVALID_CREDENTIALS", ...)`
- BD: Inserta en `audit_logs` con error_code, status_code, etc.
- Respuesta: `{ "error": "Invalid credentials" }` (sin detalles técnicos)

**7. Resultado en audit_logs**
```json
{
  "event_type": "LOGIN_FAILURE",
  "action": "error_occurred",
  "new_values": {
    "error_code": "INVALID_CREDENTIALS",
    "error_message": "Invalid credentials",
    "status_code": 401
  },
  "timestamp": "2026-02-15T10:30:00Z"
}
```

## Mejores Prácticas

### ✅ Haz esto

```go
// Crear error con contexto
err := audit.ErrInvalidToken("signature_verification_failed")

// Auditar en servicio
s.LogAuditError(ctx, "LOGIN_FAILURE", err, nil)

// Handler delega auditoría al servicio
h.handleAuthError(w, r, err, eventType, userID)
```

### ❌ NO hagas esto

```go
// NO: Exponer detalles técnicos al cliente
writeError(w, 401, "JWT signature verification failed", ...)

// NO: Insertar en audit_logs errores de validación
if err := req.Validate(); err != nil {
    s.LogAuditError(ctx, "LOGIN_FAILURE", err, nil) // ❌ Wrong
}

// NO: Auditar errores técnicos internos
if dbErr != nil {
    s.LogAuditError(ctx, "DATABASE_ERROR", dbErr, nil) // ❌ Wrong
}
```

## Testing

### Test para verificar auditoría

```go
func TestLoginFailureAuditing(t *testing.T) {
    // Arrange
    mockAuditRepo := &MockAuditRepository{}
    service := NewAuthService(..., mockAuditRepo)
    
    // Act
    _, err := service.Login(ctx, invalidCredentials, ...)
    
    // Assert
    require.Error(t, err)
    require.IsType(t, &audit.AuditableError{}, err)
    
    // Verificar que se auditó
    assert.True(t, mockAuditRepo.LogEventCalled)
    assert.Equal(t, "LOGIN_FAILURE", mockAuditRepo.LastLog.EventType)
}
```

## Resumen

| Aspecto | Anterior | Ahora |
|---------|----------|-------|
| Errores de auth genéricos | `errors.New("invalid credentials")` | `audit.ErrInvalidCredentials()` |
| Auditoría manual | Dispersa en servicios | Centralizada en handlers |
| Registro de errores | Solo logs técnicos | Logs técnicos + audit_logs |
| Búsqueda de eventos | Por `event_type` | Por `error_code` + `event_type` |
| Seguridad | Básica | Trazabilidad completa |

## Ficheros Modificados

- `pkg/audit/errors.go` - Nueva: Tipos de errores auditables
- `internal/services/auth_service.go` - Extendido: Lanza AuditableError + LogAuditError()
- `internal/handlers/auth_handler.go` - Extendido: handleAuthError() para auditar automáticamente
- `internal/core/ports/ports.go` - Actualizado: AuthService con LogAuditError()
