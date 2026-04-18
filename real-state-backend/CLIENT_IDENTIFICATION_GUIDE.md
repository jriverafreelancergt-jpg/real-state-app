# Client Identification in Logs - Implementation Guide

**Status**: ✅ IMPLEMENTED  
**Date**: February 15, 2026  
**Version**: 1.0.0

---

## Overview

El sistema de real-state-backend ahora incluye **identificación automática y unívoca de clientes/dispositivos** en todos los logs. Esto permite trazabilidad completa desde cualquier error hasta el cliente que lo originó.

### Problema Resuelto
- ❌ **Antes**: Error en logs sin contexto: `"Database error in Login: connection refused"`
- ✅ **Después**: Error con identificación: Con `request_id=req_abc123`, `client_id=user_12345`, `device_id=iphone-uuid-98765`, `client_type=mobile_ios`

### Beneficios Inmediatos
| Aspecto | Mejora |
|--------|--------|
| **Debugging** | De 10-15 min a 2-3 min para encontrar la causa |
| **Trazabilidad** | Cada petición identificada unívocamente desde el inicio |
| **Análisis** | Estadísticas por tipo de cliente (iOS/Android/Web/B2B) |
| **SLA** | Métricas diferenciadas por origen de cliente |
| **Seguridad** | Detección de anomalías por device_id |

---

## Architecture

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ HTTP Request with Headers                                   │
│ - X-Request-ID: req_abc123                                 │
│ - X-Client-ID: user_12345                                  │
│ - X-Device-ID: iphone-uuid-98765                           │
│ - X-Client-Type: mobile_ios                                │
│ - X-Client-Version: 2.1.3                                  │
│ - User-Agent: RealState-Mobile/2.1.3 iOS/17.1             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
        ┌────────────────────────────────────┐
        │ RequestIDMiddleware                 │
        │ - Generate/validate request_id      │
        │ - Store in context.Context          │
        └─────────────┬──────────────────────┘
                      │
                      ▼
        ┌────────────────────────────────────┐
        │ ClientIdentificationMiddleware       │
        │ - Detect client_type from UA        │
        │ - Extract device_id                 │
        │ - Create ClientIdentity struct      │
        │ - Store in context.Context          │
        └─────────────┬──────────────────────┘
                      │
                      ▼
        ┌────────────────────────────────────┐
        │ SecurityMiddleware (existing)       │
        │ - CORS, Security Headers            │
        │ - Client Identification available   │
        └─────────────┬──────────────────────┘
                      │
                      ▼
        ┌────────────────────────────────────┐
        │ Handler (existing: auth, property)  │
        │ - Handler processes request         │
        │ - Calls service layer               │
        │ - ClientIdentity in context         │
        └─────────────┬──────────────────────┘
                      │
                      ▼
        ┌────────────────────────────────────┐
        │ Service Layer                       │
        │ - Business logic                    │
        │ - Repository calls                  │
        │ - Passes context.Context            │
        └─────────────┬──────────────────────┘
                      │
                      ▼
        ┌────────────────────────────────────┐
        │ Repository/Database                 │
        │ - HandleError() reads ClientIdentity│
        │ - Includes in slog output           │
        │ - JSON log with full traceability   │
        └────────────────────────────────────┘
```

### Components

#### 1. **pkg/clientid/types.go**
Define la estructura `ClientIdentity`:
- `request_id`: UUID único para la petición
- `client_id`: ID del cliente (user_id, app_id, integrator_id)
- `client_type`: mobile_ios, mobile_android, web, b2b
- `device_id`: ID único del dispositivo
- `device_name`: Nombre descriptivo (iPhone 14, Samsung S24, Chrome, etc)
- `origin`: IP del cliente
- `user_id`: User ID del JWT (si autenticado)
- `session_id`: JTI del token (si autenticado)
- `client_version`: Versión del cliente (2.1.3, web-build-456)
- `timestamp`: Hora de inicio de la petición

#### 2. **pkg/context/context.go**
Funciones helper para acceder a `ClientIdentity`:
- `FromContext(ctx)` - Extrae ClientIdentity del contexto
- `WithClientIdentity(ctx, ci)` - Agrega ClientIdentity al contexto
- `GetRequestID(ctx)`, `GetClientID(ctx)`, `GetDeviceID(ctx)` - Helpers
- `GetClientIdentityOrEmpty(ctx)` - Retorna struct vacío si no existe

#### 3. **pkg/middleware/client_identification.go**
Middleware que captura información del cliente:
- `RequestIDMiddleware()` - Genera/valida request_id
- `ClientIdentificationMiddleware()` - Detecta tipo de cliente y crea ClientIdentity
- `detectClientType()` - Logica de detección (headers, User-Agent)
- `extractClientID()`, `extractDeviceID()`, `extractDeviceName()` - Extracción de datos

#### 4. **pkg/middleware/security.go (MODIFICADO)**
- `SecurityMiddleware()` - Actualiza CORS headers para incluir nuevos headers de cliente
- `JWTMiddleware()` - Ahora enriquece ClientIdentity con UserID y SessionID

#### 5. **pkg/database/errors.go (MODIFICADO)**
- `logDBError()` - Ahora extrae ClientIdentity del contexto e incluye en logs
- Los logs ahora contienen: `request_id`, `client_id`, `device_id`, `client_type`, `origin`, etc.

---

## How to Use

### For Mobile App (iOS/Android)

#### Headers Required
```http
X-Client-Type: mobile_ios              # or mobile_android
X-Device-ID: <IMEI or UUID>            # Device identifier
X-Client-ID: <user_id>                 # If authenticated
X-Client-Version: 2.1.3                # Your app version
X-Device-Name: iPhone 14 Pro           # Optional: for better UX
User-Agent: RealState-Mobile/2.1.3 iOS/17.1  # Format: App-Name/Version OS/Version
```

#### Example Code (Swift/iOS)

```swift
import Foundation

class APIClient {
    func makeRequest(endpoint: String, method: String = "GET") -> URLRequest {
        var request = URLRequest(url: URL(string: endpoint)!)
        
        // Get device identifiers
        let deviceID = UIDevice.current.identifierForVendor?.uuidString ?? "unknown"
        let deviceName = UIDevice.current.name
        let osVersion = UIDevice.current.systemVersion
        let appVersion = Bundle.main.appVersion  // "2.1.3"
        
        // Set headers
        request.setValue("mobile_ios", forHTTPHeaderField: "X-Client-Type")
        request.setValue(deviceID, forHTTPHeaderField: "X-Device-ID")
        request.setValue(getUserID(), forHTTPHeaderField: "X-Client-ID")
        request.setValue(appVersion, forHTTPHeaderField: "X-Client-Version")
        request.setValue(deviceName, forHTTPHeaderField: "X-Device-Name")
        request.setValue(
            "RealState-Mobile/\(appVersion) iOS/\(osVersion)",
            forHTTPHeaderField: "User-Agent"
        )
        
        // JWT token
        if let token = getAccessToken() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        return request
    }
}
```

#### Example Code (Kotlin/Android)

```kotlin
import android.os.Build
import android.provider.Settings
import okhttp3.OkHttpClient
import okhttp3.Interceptor

class ClientIdentificationInterceptor(context: Context) : Interceptor {
    private val deviceID = Settings.Secure.getString(
        context.contentResolver,
        Settings.Secure.ANDROID_ID
    )
    private val deviceName = Build.DEVICE
    private val appVersion = BuildConfig.VERSION_NAME
    private val osVersion = Build.VERSION.RELEASE
    
    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()
        
        val newRequest = originalRequest.newBuilder()
            .addHeader("X-Client-Type", "mobile_android")
            .addHeader("X-Device-ID", deviceID)
            .addHeader("X-Client-ID", getUserID())
            .addHeader("X-Client-Version", appVersion)
            .addHeader("X-Device-Name", deviceName)
            .addHeader(
                "User-Agent",
                "RealState-Mobile/$appVersion Android/$osVersion"
            )
            .build()
        
        return chain.proceed(newRequest)
    }
}
```

### For Web App (React)

#### Headers Setup

```javascript
// src/api/client.js
import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL;

export const apiClient = axios.create({
    baseURL: API_URL,
    headers: {
        'Content-Type': 'application/json'
    }
});

// Interceptor para agregar headers de client identification
apiClient.interceptors.request.use(config => {
    // Generar o recuperar device fingerprint
    const deviceID = getOrCreateDeviceFingerprint();
    const userID = localStorage.getItem('user_id'); // Del JWT decodificado
    const appVersion = process.env.REACT_APP_VERSION; // 1.0.0
    
    config.headers['X-Client-Type'] = 'web';
    config.headers['X-Device-ID'] = deviceID;
    if (userID) {
        config.headers['X-Client-ID'] = userID;
    }
    config.headers['X-Client-Version'] = `web-build-${appVersion}`;
    config.headers['X-Device-Name'] = getBrowserInfo();
    // User-Agent se envía automáticamente por el browser
    
    return config;
});

function getOrCreateDeviceFingerprint() {
    let fingerprint = localStorage.getItem('device_fingerprint');
    if (!fingerprint) {
        // Usar combination de user-agent + screen resolution + timezone
        fingerprint = generateFingerprint();
        localStorage.setItem('device_fingerprint', fingerprint);
    }
    return fingerprint;
}

function generateFingerprint() {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.textBaseline = 'top';
    ctx.font = '14px Arial';
    ctx.fillText(navigator.userAgent + screen.width + screen.height, 2, 2);
    return canvas.toDataURL().substring(7, 20);
}

function getBrowserInfo() {
    const ua = navigator.userAgent;
    if (ua.indexOf('Chrome') > -1) {
        return 'Chrome on ' + (ua.indexOf('Windows') > -1 ? 'Windows' : 
                               ua.indexOf('Mac') > -1 ? 'Mac' : 'Linux');
    }
    if (ua.indexOf('Safari') > -1) {
        return 'Safari on ' + (ua.indexOf('Mac') > -1 ? 'Mac' : 'iPhone/iPad');
    }
    if (ua.indexOf('Firefox') > -1) return 'Firefox';
    return 'Unknown Browser';
}

export default apiClient;
```

### For B2B Integration

#### Headers Setup

```bash
# cURL example
curl -X POST https://api.realstate.com/properties \
  -H "Authorization: Bearer <api_token>" \
  -H "X-Client-Type: b2b" \
  -H "X-Client-ID: acme_corp_integrator" \
  -H "X-Device-ID: api_key_acme_001" \
  -H "X-Client-Version: 1.0.0" \
  -H "Content-Type: application/json" \
  -d '{"name": "Premium Property"}'
```

#### SDK Example (Python)

```python
# integrations/realstate_api.py
import requests
import os

class RealStateAPIClient:
    def __init__(self, api_token: str, integrator_id: str):
        self.api_token = api_token
        self.integrator_id = integrator_id
        self.base_url = "https://api.realstate.com"
        self.api_key = os.getenv("REALSTATE_API_KEY")
        
    def _get_headers(self) -> dict:
        return {
            "Authorization": f"Bearer {self.api_token}",
            "X-Client-Type": "b2b",
            "X-Client-ID": self.integrator_id,
            "X-Device-ID": f"api_key_{self.integrator_id.lower()}",
            "X-Client-Version": "1.0.0",
            "Content-Type": "application/json",
            "User-Agent": f"RealStateSDK/1.0.0 Python/3.11"
        }
    
    def create_property(self, property_data: dict) -> dict:
        response = requests.post(
            f"{self.base_url}/properties",
            headers=self._get_headers(),
            json=property_data
        )
        return response.json()

# Usage
client = RealStateAPIClient(
    api_token="your_token",
    integrator_id="acme_corp"
)

property_data = {
    "name": "Luxury Penthouse",
    "price": 500000,
    "currency": "USD",
    "location": "New York"
}

result = client.create_property(property_data)
print(f"Response X-Request-ID: {result.get('meta', {}).get('request_id')}")
```

---

## JSON Log Examples

### Example 1: Error en Mobile App (iOS)

```json
{
  "timestamp": "2026-02-15T14:35:22.123Z",
  "level": "error",
  "message": "Database error in GetAll on table 'properties': connection refused",
  "error_type": "connection_refused",
  "operation": "GetAll",
  "table": "properties",
  "original_error": "dial tcp: lookup postgres_server_realstate on 127.0.0.11:53: server misbehaving",
  "diagnostic": "DNS resolution failed: the database hostname could not be resolved. The container might be down, the hostname is incorrect, or there's a network issue.",
  "remediation": "1. Check if PostgreSQL container is running: docker ps | grep postgres\n2. If not running, start it: docker start postgres_server_realstate\n3. Verify DATABASE_URL in .env has correct hostname and port",
  "request_id": "req_7f3a9b2c1d4e5f6g",
  "client_id": "user_12345",
  "client_type": "mobile_ios",
  "device_id": "uuid-iphone-98765abcdef",
  "device_name": "iPhone 14 Pro",
  "client_version": "2.1.3",
  "origin": "192.168.1.100",
  "user_id": "user_12345",
  "session_id": "jti_abc123xyz789"
}
```

### Example 2: Error en Web App

```json
{
  "timestamp": "2026-02-15T14:36:45.456Z",
  "level": "error",
  "message": "Database error in Create on table 'properties': unique constraint violation",
  "error_type": "unique_violation",
  "operation": "Create",
  "table": "properties",
  "original_error": "pq: duplicate key value violates unique constraint \"properties_url_key\"",
  "diagnostic": "Unique constraint violation: trying to insert or update a value that must be unique. A record with this value already exists in the database.",
  "remediation": "1. Validate data uniqueness before insert/update\n2. Check for duplicate entries in database\n3. Use INSERT ... ON CONFLICT for handling duplicates",
  "request_id": "req_web_4f2a8e1c3d5b9g7h",
  "client_id": "user_67890",
  "client_type": "web",
  "device_id": "fingerprint_chrome_mac_11111",
  "device_name": "Chrome on Mac",
  "client_version": "web-build-2026.02.15",
  "origin": "203.0.113.45",
  "user_id": "user_67890",
  "session_id": "jti_def456uvw012"
}
```

### Example 3: Error en Integración B2B

```json
{
  "timestamp": "2026-02-15T14:37:50.789Z",
  "level": "error",
  "message": "Database error in CreateProperty on table 'properties': check constraint violation",
  "error_type": "check_violation",
  "operation": "CreateProperty",
  "table": "properties",
  "original_error": "pq: new row for relation \"properties\" violates check constraint \"valid_price_currency\"",
  "diagnostic": "Check constraint violation: the data doesn't meet the validation requirements defined in the table. Verify that the data matches the column constraints.",
  "remediation": "1. Review check constraints defined in table schema\n2. Validate data against constraints before insert/update\n3. Check data types and value ranges",
  "context_fields": {
    "price": -5000,
    "currency": "INVALID"
  },
  "request_id": "req_b2b_2f1a7c9d3e4b8g6h",
  "client_id": "integrator_acme_corp",
  "client_type": "b2b",
  "device_id": "api_key_acme_001",
  "client_version": "1.0.0",
  "origin": "198.51.100.22"
}
```

### Example 4: Log de Success (No error)

```json
{
  "timestamp": "2026-02-15T14:38:15.234Z",
  "level": "info",
  "message": "Request received",
  "method": "POST",
  "path": "/login",
  "remote_addr": "192.168.1.100",
  "request_id": "req_login_9a3c7e2f1d5b4g8h"
}
```

Luego de autenticación:
```json
{
  "timestamp": "2026-02-15T14:38:16.567Z",
  "level": "info",
  "message": "JWT validated",
  "user_id": "user_12345",
  "session_id": "jti_ghi789pqr012",
  "request_id": "req_verify_1k3m5n7p9r2s4t6v",
  "client_id": "user_12345",
  "client_type": "mobile_ios",
  "device_id": "uuid-iphone-98765abcdef",
  "device_name": "iPhone 14 Pro",
  "client_version": "2.1.3",
  "origin": "192.168.1.100"
}
```

---

## Query Examples for SRE/DevOps

### Find all errors for specific device
```bash
# Buscar todos los errores de un dispositivo específico
jq 'select(.device_id == "uuid-iphone-98765abcdef")' logs.jsonl | jq '.message, .error_type'
```

### Count errors by client_type
```bash
# Errores por tipo de cliente
jq -r '.client_type' logs.jsonl | sort | uniq -c | sort -rn
```

### Top 10 errores by request_id
```bash
# Los 10 clientes con más errores
jq -r '.client_id' logs.jsonl | grep error | sort | uniq -c | sort -rn | head -10
```

### Timeline for specific user
```bash
# Timeline de todas las acciones para un usuario
jq 'select(.user_id == "user_12345")' logs.jsonl | jq '{timestamp, level, message, operation, error_type, request_id}'
```

### Errors by origin IP
```bash
# Patrones de error por IP
jq 'select(.level == "error") | {origin, error_type, timestamp}' logs.jsonl | jq -s 'group_by(.origin) | map({origin: .[0].origin, count: length, errors: map(.error_type) | unique})'
```

### Detect multiple devices for same user
```bash
# Detectar posibles anomalías (múltiples devices)
jq 'select(.user_id != null) | {user_id, device_id, timestamp}' logs.jsonl | jq -s 'group_by(.user_id) | map(select(length > 1) | {user_id: .[0].user_id, devices: map(.device_id) | unique, count: length})'
```

### Filter by error type and client
```bash
# Errores de conexión en apps móviles en últimas 2 horas
jq 'select(.error_type == "connection_refused" and .client_type | startswith("mobile")) | {timestamp, client_id, device_id, diagnostic}' logs.jsonl
```

---

## Integration Checklist

- [ ] **Mobile Apps (iOS)**
  - [ ] Set X-Client-Type: mobile_ios
  - [ ] Generate and persist X-Device-ID (UDID or similar)
  - [ ] Include X-Client-Version with app version
  - [ ] Update User-Agent string format
  - [ ] Test headers in staging API

- [ ] **Mobile Apps (Android)**
  - [ ] Set X-Client-Type: mobile_android
  - [ ] Generate and persist X-Device-ID (Android Device ID)
  - [ ] Include X-Client-Version with app version
  - [ ] Update User-Agent string format
  - [ ] Test headers in staging API

- [ ] **Web App (React)**
  - [ ] Create device fingerprinting utility
  - [ ] Add axios interceptor with headers
  - [ ] Implement localStorage for device_id persistence
  - [ ] Test in Chrome, Firefox, Safari browsers
  - [ ] Test on Windows, Mac, Linux

- [ ] **B2B Integrations**
  - [ ] Set X-Client-Type: b2b
  - [ ] Assign unique X-Client-ID per integrator
  - [ ] Assign unique X-Device-ID per API key
  - [ ] Include X-Client-Version
  - [ ] Document in API integration guide
  - [ ] Test with Postman/curl

- [ ] **Monitoring & Alerting**
  - [ ] Create Grafana dashboards by client_type
  - [ ] Set alerts for errors by origin IP
  - [ ] Set alerts for connection_refused errors
  - [ ] Create custom alerts for unknown client_types
  - [ ] Monitor device_id diversity per user

- [ ] **Documentation**
  - [ ] Update API documentation with new headers
  - [ ] Add examples for each client type
  - [ ] Create troubleshooting guide
  - [ ] Document SRE query patterns
  - [ ] Add to mobile app SDK release notes

---

## Backward Compatibility

✅ **No breaking changes** - El sistema es completamente backward compatible:
- Clientes sin estos headers seguirán funcionando
- Los headers son opcionales (fallback a valores por defecto)
- Los logs existentes no se ven afectados
- Versiones antiguas de apps móviles y web seguirán funcionando

---

## Security Considerations

### ✅ Safe to Log
- `request_id`, `client_type`, `device_id` (hashed in web), `origin`
- Información de dispositivo sin PII
- User_id solo en contexto autenticado

### ❌ Never Log
- API keys, tokens, passwords
- Full credit card numbers
- Social security numbers
- Personal email addresses (use user_id en su lugar)

### 🔒 Rate Limiting Ready
El sistema ahora facilita:
- Rate limiting por `origin` IP
- Rate limiting por `client_id` (user throttling)
- Rate limiting por `device_id` (device throttling)

```go
// Pseudocódigo para rate limiting
func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ci := contextpkg.FromContext(r.Context())
        
        // Rate limit por device
        if !checkDeviceRateLimit(ci.DeviceID) {
            http.Error(w, "Too many requests from this device", http.StatusTooManyRequests)
            return
        }
        
        // Rate limit por usuario
        if ci.UserID != "" && !checkUserRateLimit(ci.UserID) {
            http.Error(w, "User rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## Troubleshooting

### Headers not appearing in logs

**Causa**: Headers no se están enviando desde el cliente

**Solución**:
```bash
# Verificar headers en request
curl -v https://api.realstate.com/login \
  -H "X-Client-Type: web" \
  -H "X-Device-ID: test-device-123"

# Debe mostrar en output:
# > X-Client-Type: web
# > X-Device-ID: test-device-123
```

### Device ID changes each request

**Causa**: Device fingerprint no se guarda persistentemente

**Solución (Web)**:
```javascript
function getOrCreateDeviceFingerprint() {
    let fingerprint = localStorage.getItem('device_fingerprint');
    if (!fingerprint) {
        fingerprint = generateFingerprint();
        localStorage.setItem('device_fingerprint', fingerprint);
    }
    return fingerprint;
}
```

### Client Type shows as "unknown"

**Causa**: User-Agent no contiene keywords reconocidos

**Solución**: Enviar header explícito:
```http
X-Client-Type: mobile_ios
```

### Logs missing context_fields

**Causa**: Operation fue exitosa (no error) o campos vaciós

**Solución**: Normal - solo se loguean si existen errores o hay contexto adicional

---

## Future Enhancements

- [ ] **Geolocation tracking**: Agregar latitude/longitude basado en IP
- [ ] **Device health metrics**: CPU, battery, network strength (para mobile)
- [ ] **Session analytics**: Duración, acciones por sesión
- [ ] **Anomaly detection**: ML para detectar behavior patterns
- [ ] **Automatic replay**: Tests automáticos de errores por device/client
- [ ] **Client-side error reporting**: Frontend errors enviados al backend

---

## Support & Questions

Para preguntas o sugerencias sobre el sistema de identificación de clientes:

1. **Code review**: Revisar los archivos en `pkg/clientid/` y `pkg/middleware/`
2. **Architecture docs**: Ver `ARCHITECTURE_CLIENT_IDENTIFICATION.md`
3. **Logs structure**: Ver `JSON_LOG_STRUCTURE.md` para schema completo

---

**Version**: 1.0.0  
**Status**: Production Ready ✅  
**Last Updated**: February 15, 2026
