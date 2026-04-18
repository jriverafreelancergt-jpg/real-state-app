# Mejora de Logging para Errores de Base de Datos

## 📋 Resumen

Este documento describe la implementación de un sistema mejorado y centralizado de logging para errores de base de datos en la aplicación Go del backend de bienes raíces.

## 🎯 Objetivos alcanzados

✅ **Clasificación automática de errores**: El sistema identifica automáticamente el tipo de error (conexión, timeout, integridad, etc.)

✅ **Logging estructurado**: Los errores se loguean en formato JSON con contexto relevante

✅ **Funciones helper**: Utilidades para verificar tipos específicos de errores

✅ **Cobertura completa**: Implementación en todos los repositorios (users, properties, sessions, audit, security_config)

✅ **Cero cambios en interfaces**: Backward compatible con el código existente

## 📁 Estructura de archivos

```
pkg/database/
├── errors.go        # ← Lógica principal (clasificación, logging)
├── doc.go           # ← Documentación con ejemplos
└── examples.go      # ← Ejemplos de integración en handlers
```

## 🔍 Tipos de errores clasificados

### Errores de Conexión (CRITICAL)
- `connection_refused`: Servidor BD no accesible
- `connection_timeout`: Timeout en la conexión
- `connection_lost`: Conexión perdida durante operación
- `pool_exhausted`: Pool de conexiones agotado

### Errores de Operación
- `no_rows`: Sin resultados (común en SELECT)
- `operation_timeout`: Query muy lenta
- `context_canceled`: Contexto cancelado
- `transaction_done`: Transacción ya completada

### Errores de Integridad
- `unique_violation`: Violación de constraint único
- `foreign_key_violation`: Violación de FK
- `check_violation`: Violación de check constraint

## 💻 Ejemplo de uso

### Antes (sin logging mejorado)

```go
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
    query := `SELECT ... FROM users WHERE username = $1`
    user := &domain.User{}
    err := r.db.QueryRowContext(ctx, query, username).Scan(...)
    if err != nil {
        return nil, err  // ❌ Sin contexto, sin clasificación
    }
    return user, nil
}
```

**Log resultante**: Solo el error bruto sin contexto

### Después (con logging mejorado)

```go
import "real-state-backend/pkg/database"

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
    query := `SELECT ... FROM users WHERE username = $1`
    user := &domain.User{}
    err := r.db.QueryRowContext(ctx, query, username).Scan(...)
    if err != nil {
        // ✅ Con contexto, clasificación y logging automático
        return nil, database.HandleError(ctx, err, "GetByUsername", "users", map[string]interface{}{
            "username": username,
        })
    }
    return user, nil
}
```

**Log resultante**: Estructurado con tipo de error, operación, tabla y contexto

```json
{
  "time": "2026-02-14T20:31:14Z",
  "level": "ERROR",
  "msg": "Database error in GetByUsername on table 'users': connection refused (database server not accessible)",
  "error_type": "connection_refused",
  "operation": "GetByUsername",
  "table": "users",
  "original_error": "connection refused",
  "context_fields": {"username": "john_doe"}
}
```

## 🛠️ Funciones disponibles

### `HandleError(ctx, err, operation, table, fields)`
Maneja y loguea un error de BD con clasificación automática.

```go
return database.HandleError(ctx, err, "Create", "users", map[string]interface{}{
    "username": user.Username,
    "email": user.Email,
})
```

### `HandleErrorSimple(err, operation, table)`
Versión simplificada sin contexto personalizado.

```go
return database.HandleErrorSimple(err, "GetByID", "users")
```

### Funciones helper (para handlers HTTP)

```go
// Verificar si es error de conexión (retornar 503)
if database.IsConnectionError(err) {
    return http.StatusServiceUnavailable
}

// Verificar si es error de no encontrado (retornar 404)
if database.IsNotFound(err) {
    return http.StatusNotFound
}

// Verificar si es timeout (retornar 504)
if database.IsTimeout(err) {
    return http.StatusGatewayTimeout
}

// Verificar si es error de integridad (retornar 409)
if database.IsIntegrityError(err) {
    return http.StatusConflict
}
```

## 📊 Niveles de logging por tipo de error

| Tipo de Error | Nivel | Razón |
|---|---|---|
| `connection_*` | ERROR | Crítico: afecta todo el servicio |
| `pool_exhausted` | ERROR | Crítico: sin capacidad |
| `unique_violation`, `foreign_key_violation` | ERROR | Error de lógica en aplicación |
| `operation_timeout` | WARN | Transitorios, pueden reintentar |
| `no_rows` | INFO | No es error, es resultado válido |
| `context_canceled` | INFO | Normal, usuario canceló |

## 🔄 Ciclo de vida del error

```
Operación DB
    ↓
[Error ocurre]
    ↓
HandleError() clasifica automáticamente
    ↓
Genera mensaje descriptivo
    ↓
Loguea con contexto estructurado (JSON)
    ↓
Retorna DBError con metadatos
    ↓
Handler HTTP verifica tipo con helper
    ↓
Retorna código HTTP apropiado
```

## 🔧 Implementación en repositorios

Todos los repositorios han sido actualizados:

- ✅ `user_repository.go`: 6 métodos actualizados
- ✅ `property_repository.go`: 3 métodos actualizados
- ✅ `session_repository.go`: 5 métodos actualizados
- ✅ `security_config_repository.go`: 4 métodos actualizados
- ✅ `audit_repository.go`: 1 método actualizado

**Total: 19 métodos con logging mejorado**

## 📈 Mejoras en debugging

### Antes
```
Error: connection refused
```

### Después
```json
{
  "time": "2026-02-14T20:31:14Z",
  "level": "ERROR",
  "msg": "Database error in GetByUsername on table 'users': connection refused (database server not accessible)",
  "error_type": "connection_refused",
  "operation": "GetByUsername",
  "table": "users",
  "original_error": "connection refused",
  "context_fields": {"username": "john_doe"}
}
```

**Beneficios**:
- 🎯 Tipo de error específico
- 📍 Ubicación exacta (tabla, operación)
- 🔍 Contexto relevante (qué datos se buscaban)
- 📅 Timestamp exacto
- 🔗 Trazabilidad completa

## 🔐 Consideraciones de seguridad

- ✅ Los logs contienen datos públicos (username, email) pero NO contraseñas
- ✅ Los IDs se loguean para trazabilidad
- ✅ Los errores originales son internos, no se exponen a clientes
- ✅ En handlers, se retornan mensajes amigables al usuario

## 🧪 Testing

### Test unitario básico

```go
func TestGetUserNotFound(t *testing.T) {
    ctx := context.Background()
    repo := NewUserRepository(db)
    
    // Usar usuario que no existe
    user, err := repo.GetByUsername(ctx, "nonexistent")
    
    assert.Nil(t, user)
    assert.NotNil(t, err)
    assert.True(t, database.IsNotFound(err))
}
```

### Test de error de conexión

```go
func TestGetUserConnectionError(t *testing.T) {
    ctx := context.Background()
    // Mock DB que simula conexión rechazada
    repo := &MockUserRepository{
        GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
            return nil, database.HandleError(ctx, 
                errors.New("connection refused"), 
                "GetByUsername", "users", nil)
        },
    }
    
    user, err := repo.GetByUsername(ctx, "test")
    
    assert.Nil(t, user)
    assert.True(t, database.IsConnectionError(err))
}
```

## 🚀 Próximos pasos (futuro)

1. **Métricas de Prometheus**: Contar errores por tipo
2. **Circuit Breaker**: Detección automática de BD caída
3. **Distributed Tracing**: Integración con OpenTelemetry
4. **Alertas automáticas**: NotificacionesErrores críticos
5. **Retry automático**: Reintentos con backoff exponencial

## 📞 FAQ

**P: ¿Esto es un breaking change?**
R: No, es completamente backward compatible. Los repositorios siguen retornando el mismo tipo de error.

**P: ¿Se pierden los logs anteriores?**
R: No, los logs anteriores se mantienen. Esto solo agrega contexto adicional.

**P: ¿Puedo desactivar el logging?**
R: No directly, pero puedes cambiar el nivel de slog en config (DEBUG, INFO, WARN, ERROR).

**P: ¿Cómo manejar errores en transacciones?**
R: Usa HandleError en cada operación dentro de la transacción. El contexto se propaga automáticamente.

**P: ¿Puedo agregar mi propio tipo de error?**
R: Sí, agrega el tipo a `DBErrorType` y actualiza `ClassifyError()` para detectarlo.

## 📖 Referencias

- [slog documentation](https://pkg.go.dev/log/slog)
- [database/sql package](https://pkg.go.dev/database/sql)
- [Error handling patterns](https://go.dev/blog/error-handling-and-go)
