# Ejemplos de Logs Generados por el Sistema

Este documento muestra ejemplos reales de logs que genera el nuevo sistema de logging de errores de BD.

## 📝 Escenario 1: Error de conexión rechazada

**Situación**: PostgreSQL no está disponible

```json
{
  "time": "2026-02-14T20:31:14.338236544Z",
  "level": "ERROR",
  "msg": "Database error in GetByUsername on table 'users': connection refused (database server not accessible)",
  "error_type": "connection_refused",
  "operation": "GetByUsername",
  "table": "users",
  "original_error": "connection refused",
  "context_fields": {
    "username": "john_doe"
  }
}
```

**Acciones recomendadas**:
- 🔧 Verificar que PostgreSQL esté ejecutándose
- 🔍 Revisar credenciales de conexión en .env
- 📍 Validar dirección IP/puerto de BD

---

## 📝 Escenario 2: Usuario no encontrado (no_rows)

**Situación**: Se intenta buscar un usuario que no existe

```json
{
  "time": "2026-02-14T20:31:14.338236544Z",
  "level": "INFO",
  "msg": "Database error in GetByUsername on table 'users': no rows found",
  "error_type": "no_rows",
  "operation": "GetByUsername",
  "table": "users",
  "original_error": "sql: no rows in result set",
  "context_fields": {
    "username": "nonexistent_user"
  }
}
```

**Acciones recomendadas**:
- ✅ Es normal, el usuario no existe
- 💡 En handler: Retornar 401 (sin revelar si existe)

---

## 📝 Escenario 3: Query timeout

**Situación**: Query muy lenta, excede el timeout del contexto

```json
{
  "time": "2026-02-14T20:31:18.123456789Z",
  "level": "WARN",
  "msg": "Database error in GetAll on table 'properties': operation timeout (query took too long)",
  "error_type": "operation_timeout",
  "operation": "GetAll",
  "table": "properties",
  "original_error": "context deadline exceeded",
  "context_fields": {
    "limit": 100,
    "offset": 0
  }
}
```

**Acciones recomendadas**:
- 📊 Agregar índices a las columnas usadas en WHERE
- 🔍 Revisar query execution plan
- ⏱️ Aumentar timeout (con cuidado)

---

## 📝 Escenario 4: Violación de constraint único

**Situación**: Se intenta crear usuario con email que ya existe

```json
{
  "time": "2026-02-14T20:31:14.338236544Z",
  "level": "ERROR",
  "msg": "Database error in Create on table 'users': unique constraint violation",
  "error_type": "unique_violation",
  "operation": "Create",
  "table": "users",
  "original_error": "pq: duplicate key value violates unique constraint \"idx_users_email_unique\"",
  "context_fields": {
    "username": "john_doe",
    "email": "john@example.com"
  }
}
```

**Acciones recomendadas**:
- 🔍 En handler: Validar email antes de insertar
- 💡 Retornar 409 Conflict al cliente
- ✅ Es error de lógica en aplicación, no de BD

---

## 📝 Escenario 5: Pool de conexiones agotado

**Situación**: Demasiadas conexiones concurrentes

```json
{
  "time": "2026-02-14T20:31:14.338236544Z",
  "level": "ERROR",
  "msg": "Database error in Create on table 'properties': connection pool exhausted (too many concurrent requests)",
  "error_type": "pool_exhausted",
  "operation": "Create",
  "table": "properties",
  "original_error": "no more connections available",
  "context_fields": {
    "title": "Casa moderna",
    "currency": "USD"
  }
}
```

**Acciones recomendadas**:
- 📈 Aumentar `db.SetMaxOpenConns()` en main.go
- 🔄 Implementar queue/backoff para reintentos
- 📊 Monitorear conexiones activas

---

## 📝 Escenario 6: Violación de foreign key

**Situación**: Se intenta crear propiedad con usuario_id que no existe

```json
{
  "time": "2026-02-14T20:31:14.338236544Z",
  "level": "ERROR",
  "msg": "Database error in Create on table 'properties': foreign key constraint violation",
  "error_type": "foreign_key_violation",
  "operation": "Create",
  "table": "properties",
  "original_error": "pq: insert or update on table \"properties\" violates foreign key constraint \"fk_properties_user_id\"",
  "context_fields": {
    "user_id": "invalid_uuid",
    "title": "Casa moderna"
  }
}
```

**Acciones recomendadas**:
- ✅ Verificar que el usuario_id existe antes de crear propiedad
- 📍 En handler: Validar IDs foráneos
- 💡 Retornar 400 Bad Request

---

## 📝 Escenario 7: Timeout de conexión (red lenta)

**Situación**: Conexión a BD muy lenta

```json
{
  "time": "2026-02-14T20:31:20.338236544Z",
  "level": "ERROR",
  "msg": "Database error in GetByTokenJTI on table 'user_sessions': connection timeout (database server unreachable)",
  "error_type": "connection_timeout",
  "operation": "GetByTokenJTI",
  "table": "user_sessions",
  "original_error": "i/o timeout",
  "context_fields": {
    "jti": "token_jti_123"
  }
}
```

**Acciones recomendadas**:
- 🌐 Verificar conectividad de red
- 📊 Revisar latencia de red
- ⏱️ Aumentar timeout de conexión si es necesario

---

## 📝 Escenario 8: Contexto cancelado

**Situación**: Cliente canceló la solicitud

```json
{
  "time": "2026-02-14T20:31:14.338236544Z",
  "level": "INFO",
  "msg": "Database error in GetPermissions on table 'permissions': operation canceled",
  "error_type": "context_canceled",
  "operation": "GetPermissions",
  "table": "permissions",
  "original_error": "context canceled",
  "context_fields": {
    "user_id": "user_123"
  }
}
```

**Acciones recomendadas**:
- ✅ Es normal, cliente desconectó
- 💡 No requiere acción, es esperado

---

## 📝 Escenario 9: Error en múltiples operaciones (transacción)

**Situación**: Falla dentro de una transacción

```json
[
  {
    "time": "2026-02-14T20:31:14.123456789Z",
    "level": "INFO",
    "msg": "Database error in Create on table 'audit_logs': no rows found",
    "error_type": "no_rows",
    "operation": "Create",
    "table": "audit_logs"
  },
  {
    "time": "2026-02-14T20:31:14.456789012Z",
    "level": "ERROR",
    "msg": "Database error in LogEvent on table 'audit_logs': transaction done",
    "error_type": "transaction_done",
    "operation": "LogEvent",
    "table": "audit_logs",
    "original_error": "sql: transaction has already been committed or rolled back"
  }
]
```

**Acciones recomendadas**:
- 🔍 Revisar lógica de transacciones
- ✅ Asegurar que commit/rollback se hace correctamente

---

## 📊 Matriz de errores → Códigos HTTP

| Tipo de Error | Nivel Log | Código HTTP | Mensaje Cliente |
|---|---|---|---|
| `connection_refused` | ERROR | 503 | Database service temporarily unavailable |
| `connection_timeout` | ERROR | 503 | Database service temporarily unavailable |
| `pool_exhausted` | ERROR | 503 | Too many requests, please retry |
| `operation_timeout` | WARN | 504 | Request timeout, please retry |
| `no_rows` | INFO | 404 | Resource not found |
| `unique_violation` | ERROR | 409 | Resource already exists |
| `foreign_key_violation` | ERROR | 400 | Invalid reference |
| `context_canceled` | INFO | 499 | Request canceled |

---

## 🔍 Cómo leer los logs

### Estructura JSON
```json
{
  "time": "timestamp ISO 8601",
  "level": "ERROR|WARN|INFO",
  "msg": "Mensaje descriptivo con tipo de error",
  "error_type": "clasificación automática",
  "operation": "nombre del método",
  "table": "tabla afectada",
  "original_error": "error original de Go/PostgreSQL",
  "context_fields": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

### Interpretación rápida

1. **Leer `error_type`**: Identifica rápidamente el problema
2. **Leer `level`**: ERROR = crítico, WARN = transitorios, INFO = informativos
3. **Leer `operation`**: Dónde ocurrió exactamente
4. **Leer `table`**: En qué tabla
5. **Leer `context_fields`**: Qué datos estaban involucrados

---

## 🚨 Patrones de errores a buscar

### Patrón 1: Conexión inestable
```json
{
  "error_type": "connection_refused"  // Repetido varias veces
}
```
→ **Acción**: Reiniciar PostgreSQL

### Patrón 2: Queries lentas
```json
{
  "error_type": "operation_timeout"   // Frecuente en GetAll
}
```
→ **Acción**: Agregar índices

### Patrón 3: Pool agotado
```json
{
  "error_type": "pool_exhausted"      // Bajo carga
}
```
→ **Acción**: Aumentar max connections

### Patrón 4: Datos duplicados
```json
{
  "error_type": "unique_violation"    // Para usuarios/emails
}
```
→ **Acción**: Validación en aplicación

---

## 📈 Monitoreo y alertas (futuro)

```go
// Pseudocódigo para alertas
if errorType == "connection_refused" {
    sendAlert("CRITICAL: Database unreachable")
}

if errorType == "pool_exhausted" {
    sendAlert("WARNING: Connection pool exhausted")
}

if errorType == "operation_timeout" {
    incrementMetric("db_timeouts", 1)
}
```

---

## 🧪 Testing con logs

```go
func TestLoginWithDatabaseDown(t *testing.T) {
    // Setup: Mock DB que retorna connection_refused
    mockErr := database.HandleError(context.Background(),
        errors.New("connection refused"),
        "GetByUsername", "users", nil)
    
    assert.True(t, database.IsConnectionError(mockErr))
}
```

---

**Pro tip**: Los logs generados por este sistema son totalmente searchable y filtrable usando herramientas como `jq`, `loki`, `datadog`, etc.

```bash
# Buscar solo errores críticos de conexión
jq 'select(.error_type == "connection_refused")' logs.json

# Buscar timeouts en tabla 'properties'
jq 'select(.error_type == "operation_timeout" and .table == "properties")' logs.json

# Contar errores por tipo
jq -r '.error_type' logs.json | sort | uniq -c
```
