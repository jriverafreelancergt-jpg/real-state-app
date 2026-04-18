# Estructura JSON Mejorada de Logs

## 📋 Campos del log JSON

### Campos estándar de slog
```json
{
  "time": "2026-02-15T00:00:00.000000000Z",     // ISO 8601 timestamp
  "level": "ERROR|WARN|INFO",                   // Log level
  "msg": "Descripción del error"                // Mensaje principal
}
```

### Campos de diagnóstico de BD (NUEVOS)
```json
{
  "error_type": "connection_refused",           // Tipo de error (11 categorías)
  "operation": "LogEvent",                      // Método donde ocurrió
  "table": "audit_logs",                        // Tabla afectada
  "original_error": "dial tcp: lookup...",      // Error original de Go/Postgres
  "diagnostic": "DNS resolution failed: ...",   // Análisis detallado (NUEVO)
  "remediation": "1. Check if PostgreSQL...",   // Acciones recomendadas (NUEVO)
  "context_fields": {                           // Datos contextuales
    "action": "login",
    "event_type": "LOGIN_FAILURE",
    "resource": "auth"
  }
}
```

---

## 🎯 Ejemplo completo para cada tipo de error

### 1. Connection Refused (Contenedor down)

```json
{
  "time": "2026-02-15T00:05:32.123456789Z",
  "level": "ERROR",
  "msg": "Database error in LogEvent on table 'audit_logs': connection refused (database server not accessible)",
  "error_type": "connection_refused",
  "operation": "LogEvent",
  "table": "audit_logs",
  "original_error": "dial tcp: lookup postgres_server_realstate on 127.0.0.11:53: server misbehaving",
  "diagnostic": "DNS resolution failed: the database hostname could not be resolved. The container might be down, the hostname is incorrect, or there's a network issue.",
  "remediation": "1. Check if PostgreSQL container is running: docker ps | grep postgres\n2. If not running, start it: docker start postgres_server_realstate\n3. Verify DATABASE_URL in .env has correct hostname and port\n4. Check network connectivity between application and database\n5. Review database server logs for errors",
  "context_fields": {
    "event_type": "LOGIN_FAILURE"
  }
}
```

**Acción recomendada**: `docker start postgres_server_realstate`

---

### 2. Unique Violation

```json
{
  "time": "2026-02-15T00:05:33.234567890Z",
  "level": "ERROR",
  "msg": "Database error in Create on table 'users': unique constraint violation",
  "error_type": "unique_violation",
  "operation": "Create",
  "table": "users",
  "original_error": "pq: duplicate key value violates unique constraint \"idx_users_email_unique\"",
  "diagnostic": "Unique constraint violation: trying to insert or update a value that must be unique. A record with this value already exists in the database.",
  "remediation": "1. Validate data uniqueness before insert/update\n2. Check for duplicate entries in database\n3. Use INSERT ... ON CONFLICT for handling duplicates\n4. Review application business logic\n5. Implement proper error handling for this constraint",
  "context_fields": {
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

**Acción recomendada**: Validar email antes de insertar

---

### 3. Operation Timeout

```json
{
  "time": "2026-02-15T00:05:34.345678901Z",
  "level": "WARN",
  "msg": "Database error in GetAll on table 'properties': operation timeout (query took too long)",
  "error_type": "operation_timeout",
  "operation": "GetAll",
  "table": "properties",
  "original_error": "context deadline exceeded",
  "diagnostic": "Query execution timeout: the query took longer than expected to complete. The query might be inefficient, or the database is under heavy load.",
  "remediation": "1. Add indexes to frequently queried columns\n2. Run EXPLAIN ANALYZE to review query plan\n3. Optimize the query (reduce data transferred)\n4. Increase query timeout if appropriate\n5. Check database server resources (CPU, disk I/O)",
  "context_fields": {
    "limit": 1000,
    "offset": 0
  }
}
```

**Acción recomendada**: Agregar índices o optimizar query

---

### 4. Foreign Key Violation

```json
{
  "time": "2026-02-15T00:05:35.456789012Z",
  "level": "ERROR",
  "msg": "Database error in Create on table 'properties': foreign key constraint violation",
  "error_type": "foreign_key_violation",
  "operation": "Create",
  "table": "properties",
  "original_error": "pq: insert or update on table \"properties\" violates foreign key constraint \"fk_properties_user_id\"",
  "diagnostic": "Foreign key constraint violation: trying to reference a record that doesn't exist. The parent record in the referenced table must exist before creating this record.",
  "remediation": "1. Verify parent record exists before creating this record\n2. Check foreign key references are correct\n3. Review data relationships and dependencies\n4. Use transactions to ensure data consistency\n5. Add validation before database operations",
  "context_fields": {
    "user_id": "invalid_uuid",
    "title": "Casa moderna"
  }
}
```

**Acción recomendada**: Verificar que el usuario existe antes de crear propiedad

---

### 5. No Rows (No es error, es normal)

```json
{
  "time": "2026-02-15T00:05:36.567890123Z",
  "level": "INFO",
  "msg": "Database error in GetByUsername on table 'users': no rows found",
  "error_type": "no_rows",
  "operation": "GetByUsername",
  "table": "users",
  "original_error": "sql: no rows in result set",
  "diagnostic": "No records found: the query executed successfully but returned no results. This is normal when searching for non-existent data.",
  "remediation": "1. This is usually normal behavior\n2. Handle this case in application logic\n3. Return appropriate 404 response to client\n4. No action required unless this happens unexpectedly",
  "context_fields": {
    "username": "nonexistent_user"
  }
}
```

**Acción recomendada**: Retornar 404 (es normal)

---

### 6. Connection Timeout

```json
{
  "time": "2026-02-15T00:05:37.678901234Z",
  "level": "ERROR",
  "msg": "Database error in GetByID on table 'users': connection timeout (database server unreachable)",
  "error_type": "connection_timeout",
  "operation": "GetByID",
  "table": "users",
  "original_error": "i/o timeout",
  "diagnostic": "Connection timeout: unable to reach the database server within the configured timeout. The server might be overloaded, network is slow, or server is temporarily unavailable.",
  "remediation": "1. Check database server status and resources (CPU, memory)\n2. Verify network connectivity and latency\n3. Increase connection timeout if appropriate\n4. Check if server is overloaded with too many queries\n5. Review slow query logs if available",
  "context_fields": {
    "id": "user_123"
  }
}
```

**Acción recomendada**: Verificar recursos del servidor, reintentar con backoff

---

### 7. Pool Exhausted

```json
{
  "time": "2026-02-15T00:05:38.789012345Z",
  "level": "ERROR",
  "msg": "Database error in Create on table 'properties': connection pool exhausted (too many concurrent requests)",
  "error_type": "pool_exhausted",
  "operation": "Create",
  "table": "properties",
  "original_error": "no more connections available",
  "diagnostic": "Connection pool exhausted: all available connections are in use. Too many concurrent requests are trying to access the database simultaneously.",
  "remediation": "1. Increase MaxOpenConns in db.SetMaxOpenConns()\n2. Review application code for connection leaks\n3. Monitor active connections with metrics\n4. Implement connection pooling best practices\n5. Consider database connection limits",
  "context_fields": {
    "concurrent_requests": 25,
    "pool_size": 25
  }
}
```

**Acción recomendada**: Aumentar `MaxOpenConns` en `main.go`

---

### 8. Context Canceled

```json
{
  "time": "2026-02-15T00:05:39.890123456Z",
  "level": "INFO",
  "msg": "Database error in GetPermissions on table 'permissions': operation canceled",
  "error_type": "context_canceled",
  "operation": "GetPermissions",
  "table": "permissions",
  "original_error": "context canceled",
  "diagnostic": "Operation canceled: the request was canceled before completion. This usually happens when the client disconnects or the operation times out.",
  "remediation": "1. This is usually normal behavior\n2. Implement timeout handling in requests\n3. Add context cancellation callbacks if needed\n4. Monitor if this happens frequently\n5. Increase timeout if legitimate operations are being canceled",
  "context_fields": {
    "user_id": "user_123"
  }
}
```

**Acción recomendada**: Es normal (cliente desconectó)

---

## 🔄 Flujo de procesamiento

```
Error ocurre en BD
        ↓
ClassifyError() detiene el tipo
        ↓
generateDiagnostic() analiza la causa
        ↓
generateRemediation() genera pasos
        ↓
DBError struct completado con todos los campos
        ↓
logDBError() serializa a JSON con slog
        ↓
Log disponible para consulta
```

---

## 💡 Cómo usar en handlers

```go
resp, err := h.service.GetUser(r.Context(), userID)
if err != nil {
    // Si es un DBError con diagnóstico
    if dbErr, ok := err.(*database.DBError); ok {
        // El log ya tiene toda la info, pero puedes usarla también:
        slog.Error("Operation failed",
            "diagnostic", dbErr.Diagnostic,
            "remediation", dbErr.Remediation,
        )
        
        // Mapear a código HTTP
        if database.IsConnectionError(err) {
            writeError(w, http.StatusServiceUnavailable, 
                "Database service unavailable", 
                "service_unavailable")
            return
        }
    }
}
```

---

## 📊 Matriz de campos por error type

| Error Type | error_type | level | diagnostic | remediation |
|---|---|---|---|---|
| DNS fail | connection_refused | ERROR | ✅ Específico | ✅ Docker commands |
| Unique violation | unique_violation | ERROR | ✅ Específico | ✅ Validation steps |
| Query slow | operation_timeout | WARN | ✅ Específico | ✅ Index/optimize |
| No rows | no_rows | INFO | ✅ Normal | ✅ N/A |
| Pool full | pool_exhausted | ERROR | ✅ Específico | ✅ Config changes |

---

## ✅ Beneficios principales

✅ **Diagnósticos específicos**: No más "server misbehaving" genérico  
✅ **Pasos de resolución**: Qué hacer exactamente  
✅ **Auto-documentado**: El log incluye toda la info  
✅ **Timesaving**: Reduce tiempo de debugging  
✅ **Escalable**: Funciona para cualquier tipo de error  

---

**Versión**: 2.0  
**Fecha**: 15 de febrero de 2026  
**Status**: ✅ Producción lista
