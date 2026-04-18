# Arquitectura: Identificación de Cliente/Dispositivo en Logs
**Fecha**: 15 de febrero de 2026  
**Rol**: Arquitecto SRE  
**Objetivo**: Enriquecer logs con identificación unívoca de cliente/dispositivo para observabilidad y debugging

---

## 1. Problema Identificado

**Síntomas**:
- Error en logs sin contexto de quién lo originó
- No podemos distinguir si es app móvil (Android/iOS), web (React), o integrador B2B
- Imposible correlacionar errores con usuarios/dispositivos específicos
- Debugging y análisis de incidentes sin trazabilidad

**Impacto**:
- ⏱️ Tiempo de MTTR (Mean Time To Repair) aumentado
- 📊 Imposible análisis de patrones por cliente
- 🔒 Imposible detección de abuso por dispositivo
- 📱 Imposible SLA diferenciado por tipo de cliente

---

## 2. Solución Propuesta: Context-Based Client Identification

### 2.1 Componentes Principales

```
HTTP Request
    ↓
[RequestID Middleware] → Generar request_id único
    ↓
[Client Identification Middleware] → Detectar origin, device_type, client_id
    ↓
[Context Enrichment] → Agregar a context.Context
    ↓
[All Handlers/Services] → Leer desde context, propagar en logs
    ↓
[slog Output] → JSON con request_id, client_id, device_type, origin
```

### 2.2 Headers Capturados

| Header | Tipo | Origen | Descripción |
|--------|------|--------|-------------|
| `X-Request-ID` | String | Cliente/Generado | ID único de request (UUID o client-gen) |
| `X-Client-ID` | String | Cliente | ID único del cliente (user_id, app_id, integrator_id) |
| `X-Device-ID` | String | Cliente | ID único del dispositivo (IMEI, UUID, fingerprint) |
| `X-Client-Type` | String | Cliente | Tipo: `mobile_ios`, `mobile_android`, `web`, `b2b` |
| `X-Client-Version` | String | Cliente | Versión: `1.2.3` o `web-build-456` |
| `User-Agent` | String | HTTP | Browser/App info |
| `Authorization` | String | HTTP | Bearer token (extrae user_id de JWT) |

### 2.3 Estructura de Contexto (context.Context)

```go
type ClientIdentity struct {
    RequestID      string    // UUID único para la petición
    ClientID       string    // ID del cliente (usuario, app, integrador)
    ClientType     string    // mobile_ios, mobile_android, web, b2b
    DeviceID       string    // ID único del dispositivo
    DeviceName     string    // iPhone 14, Samsung S24, etc.
    Origin         string    // IP del cliente
    UserID         string    // User ID del JWT (si autenticado)
    SessionID      string    // JTI del token (si autenticado)
    ClientVersion  string    // Versión del cliente
    Timestamp      time.Time // Hora de inicio de request
}
```

### 2.4 Estrategia de Propagación

**Nivel 1: Request Entry (Middleware)**
```
→ Capturar headers
→ Generar request_id si no existe
→ Crear ClientIdentity
→ Agregar a context.Context
```

**Nivel 2: Service Layer**
```
→ Extraer ClientIdentity del contexto
→ Pasar a repository/database operations
```

**Nivel 3: Repository/Database**
```
→ Extraer ClientIdentity del contexto
→ Incluir en logs de error
→ Propagar para trazabilidad
```

**Nivel 4: Response Handler**
```
→ Incluir request_id en header X-Request-ID
→ Incluir en response JSON (si aplica)
```

---

## 3. Identificación por Origen

### 3.1 Detección Automática

```go
func DetectOrigin(r *http.Request) (clientType string, clientID string) {
    // 1. Headers explícitos (PRIORIDAD 1)
    if clientType := r.Header.Get("X-Client-Type"); clientType != "" {
        return clientType, extractClientID(r)
    }
    
    // 2. User-Agent patterns (PRIORIDAD 2)
    ua := r.Header.Get("User-Agent")
    
    if strings.Contains(ua, "RealState-Mobile") {
        // App móvil oficial
        if strings.Contains(ua, "iOS") {
            return "mobile_ios", extractClientID(r)
        }
        if strings.Contains(ua, "Android") {
            return "mobile_android", extractClientID(r)
        }
    }
    
    if strings.Contains(ua, "RealState-Web") {
        // Portal web oficial
        return "web", extractClientID(r)
    }
    
    // 3. SDK patterns (PRIORIDAD 3)
    if strings.Contains(ua, "gRPC") || strings.Contains(ua, "Postman") {
        return "b2b", extractClientID(r)
    }
    
    // 4. Fallback a Browser detection
    if strings.Contains(ua, "Chrome") || strings.Contains(ua, "Safari") {
        return "web", extractClientID(r)
    }
    
    // 5. Default
    return "unknown", extractClientID(r)
}
```

### 3.2 Client ID Hierarchy

```
1. JWT sub (user_id) si está autenticado
2. X-Client-ID header explícito
3. Device ID como fallback
4. IP + User-Agent hash como último recurso
```

---

## 4. Casos de Uso

### Caso 1: Error en app móvil
```json
{
  "timestamp": "2026-02-15T14:30:45Z",
  "level": "error",
  "message": "Failed to fetch properties",
  "error_type": "connection_refused",
  "diagnostic": "Database connection failed",
  
  // 🎯 IDENTIFICACIÓN NUEVA
  "request_id": "req_abc123def456",
  "client_id": "user_12345",
  "client_type": "mobile_ios",
  "device_id": "iphone-uuid-98765",
  "device_name": "iPhone 14 Pro",
  "client_version": "2.1.3",
  "origin": "192.168.1.100",
  
  "user_id": "user_12345",
  "operation": "GetAll",
  "table": "properties"
}
```

### Caso 2: Error en integración B2B
```json
{
  "timestamp": "2026-02-15T14:30:45Z",
  "level": "error",
  "message": "Invalid request parameters",
  "error_type": "validation_error",
  
  // 🎯 IDENTIFICACIÓN B2B
  "request_id": "req_xyz789",
  "client_id": "integrator_acme_corp",
  "client_type": "b2b",
  "device_id": "api_key_acme_001",
  "client_version": "1.0.0",
  "origin": "203.0.113.45",
  
  "operation": "CreateProperty",
  "fields_validated": ["price", "location", "status"]
}
```

### Caso 3: Error en portal web
```json
{
  "timestamp": "2026-02-15T14:30:45Z",
  "level": "error",
  "message": "Authentication failed",
  "error_type": "unique_violation",
  
  // 🎯 IDENTIFICACIÓN WEB
  "request_id": "req_web_456",
  "client_id": "user_67890",
  "client_type": "web",
  "device_id": "browser-fingerprint-11111",
  "device_name": "Chrome on MacOS",
  "client_version": "web-build-2026.02.15",
  "origin": "192.168.1.200",
  
  "user_id": "user_67890",
  "operation": "Login",
  "table": "users"
}
```

---

## 5. Implementación por Componentes

### Componente 1: pkg/clientid/types.go
```go
// Define ClientIdentity struct
// Métodos: MarshalJSON, String
```

### Componente 2: pkg/middleware/client_identification.go
```go
// ClientIdentificationMiddleware
// Funciones:
// - ExtractClientIdentity(r *http.Request) *ClientIdentity
// - DetectOrigin(r *http.Request) string
// - ExtractClientID(r *http.Request) string
```

### Componente 3: pkg/middleware/request_id.go
```go
// RequestIDMiddleware
// Genera o valida request_id
```

### Componente 4: pkg/context/context.go
```go
// Helpers para obtener ClientIdentity del contexto
// - FromContext(ctx context.Context) *ClientIdentity
// - WithClientIdentity(ctx context.Context, ci *ClientIdentity) context.Context
```

### Componente 5: Modificaciones a pkg/database/errors.go
```go
// Extraer ClientIdentity del contexto en HandleError()
// Agregar a campos de log
```

---

## 6. Impacto en Código Existente

### Minimal (Non-Breaking):
1. `middleware/security.go` → Agregar client_identification middleware
2. `cmd/api/main.go` → Registrar new middleware
3. `pkg/database/errors.go` → Leer ClientIdentity del contexto
4. `handlers/*.go` → Ya usan context, solo lectura adicional

### No Required Changes:
- DTOs (request/response)
- Services (inyección de dependencias)
- Repository interfaces

---

## 7. Beneficios por Rol

### Para SRE
- 🔍 **Trazabilidad completa**: Request → Client → Device → Error
- 📊 **Análisis por origen**: Estadísticas por iOS/Android/Web/B2B
- 🎯 **Debugging rápido**: grep -i "device_id: iphone-uuid-98765" logs

### Para Desarrolladores
- 🐛 **Reproducir errores**: "Error en iPhone, user_12345, 2026-02-15 14:30:45"
- 📱 **Testing por cliente**: Tests específicos para mobile vs web
- 🔌 **Integración B2B**: Auditoría completa de cada llamada de integrador

### Para Product
- 📈 **Análisis de uso**: Cuál versión de app tiene más errores
- 🚀 **Feature rollout**: Monitorear errores por versión de cliente
- 💰 **SLA diferenciado**: Métricas separadas por tipo de cliente

### Para Seguridad
- 🔐 **Detección de anomalías**: Múltiples devices para mismo user_id
- 📍 **Análisis geográfico**: Patrones de ubicación por device
- ⚠️ **Alertas**: "100 requests desde IP desconocida para user_12345"

---

## 8. Roadmap de Implementación

### Fase 1: Core (Esta sesión)
- ✅ Diseño de arquitectura
- ⬜ Crear pkg/clientid/types.go
- ⬜ Crear pkg/middleware/client_identification.go
- ⬜ Crear pkg/middleware/request_id.go
- ⬜ Actualizar middleware/security.go
- ⬜ Actualizar cmd/api/main.go
- ⬜ Enriquecer pkg/database/errors.go

### Fase 2: Integration
- ⬜ Verificar todos los handlers propaguen contexto
- ⬜ Agregar logs a servicios clave
- ⬜ Actualizar DTOs de respuesta (opcional: incluir request_id)

### Fase 3: Observabilidad
- ⬜ Queries de Grafana por client_type, device_type, origin
- ⬜ Alertas por rate de errores por cliente
- ⬜ Dashboard de salud por origen

### Fase 4: Documentation
- ⬜ Guía de integración para clientes móviles
- ⬜ Ejemplos de headers esperados
- ⬜ JSON log structure reference

---

## 9. Ejemplos de Queries Futuras

```bash
# Error en iOS en últimas 2 horas
grep -i 'client_type.*mobile_ios' logs | grep error | tail -1000

# Errores por integrador
jq 'select(.client_type=="b2b") | {client_id, error_type, timestamp}' logs

# Dispositivos con más errores
jq '[.device_id] | group_by(.) | sort_by(length) | reverse' logs

# Trazabilidad completa de un request
grep 'request_id.*req_abc123' logs | jq .
```

---

## 10. Seguridad y Consideraciones

### ✅ Incluir en Logs
- request_id (anónimo)
- client_type (categórico)
- device_id (anónimo)
- origin (IP)

### ❌ NO Incluir en Logs
- Contraseñas
- Tokens JWT
- Números de tarjeta
- Datos personales PII (excepto user_id en contexto autenticado)

### 🔒 Rate Limiting
- Por `origin` IP
- Por `client_id` (throttling por usuario)
- Por `device_id` (throttling por dispositivo)

---

## Conclusión

Este diseño proporciona **trazabilidad completa** sin introducir complejidad innecesaria. Los cambios son principalmente aditivos (middleware nuevo, contexto extendido) con impacto mínimo en el código existente.

**Status**: Listo para implementación ✅
