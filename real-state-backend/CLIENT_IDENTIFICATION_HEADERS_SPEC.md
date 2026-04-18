# Client Identification Headers - API Specification

**Specification Version**: 1.0  
**Date**: February 15, 2026  
**Status**: Active

---

## Overview

Los headers de Client Identification son opcionales pero altamente recomendados para:
- Mejorar debugging y troubleshooting
- Análisis de uso por tipo de cliente
- Detección de anomalías y fraude
- Métricas diferenciadas por cliente/dispositivo

---

## Standard Headers

### 1. X-Request-ID
**Propósito**: Identificador único para la petición  
**Tipo**: String (UUID v4 format recomendado)  
**Requerido**: No (se genera automáticamente si no se proporciona)  
**Ejemplo**: `req_550e8400-e29b-41d4-a716-446655440000`  
**Generado por**: Cliente o servidor (si no está presente)  
**Longitud máxima**: 36 caracteres

**Validación del servidor**:
- Si el cliente lo proporciona, se usa tal cual
- Si no lo proporciona, se genera automáticamente como `req_<UUID>`
- Se retorna en response header `X-Request-ID`

**Uso**:
```http
Request:
X-Request-ID: req_550e8400-e29b-41d4-a716-446655440000

Response:
X-Request-ID: req_550e8400-e29b-41d4-a716-446655440000
```

---

### 2. X-Client-ID
**Propósito**: Identificador único del cliente  
**Tipo**: String (alphanumeric + underscore)  
**Requerido**: Recomendado para requests autenticadas  
**Ejemplo**: `user_12345` o `integrator_acme_corp`  
**Longitud máxima**: 64 caracteres  
**Caracteres válidos**: `[a-zA-Z0-9_-]`

**Formatos recomendados por tipo de cliente**:
- **Usuarios**: `user_<id>` ej: `user_12345`
- **Integradores B2B**: `integrator_<name>` ej: `integrator_acme_corp`
- **Apps internas**: `app_<name>` ej: `app_mobile_ios`

**Validación del servidor**:
- Si se proporciona en header, se usa
- Si es request autenticada, se extrae del JWT (`sub` claim)
- Si no, se usa device_id como fallback
- Nunca será vacío en logs (fallback a "unknown_client")

---

### 3. X-Device-ID
**Propósito**: Identificador único del dispositivo  
**Tipo**: String (UUID, IMEI, fingerprint, etc)  
**Requerido**: Recomendado para todas las apps móviles  
**Ejemplo**: 
- iOS: `uuid-iphone-550e8400-e29b-41d4-a716-446655440000`
- Android: `android-device-id-12345678901234567890`
- Web: `fingerprint_browser_a1b2c3d4e5f6`  
**Longitud máxima**: 128 caracteres

**Formatos recomendados**:
- **iOS**: UUID del dispositivo (UIDevice.identifierForVendor)
- **Android**: Android Device ID (Settings.Secure.ANDROID_ID)
- **Web**: Fingerprint generado (browser + screen + timezone)
- **B2B**: API key o integrador ID

**Validación del servidor**:
- Si se proporciona, se usa tal cual
- Para web sin header, se genera fingerprint basado en User-Agent + IP
- Debe persistir durante la sesión del cliente

---

### 4. X-Client-Type
**Propósito**: Clasificación del tipo de cliente  
**Tipo**: Enum  
**Requerido**: No (se detecta automáticamente)  
**Valores válidos**:
- `mobile_ios` - Aplicación iOS
- `mobile_android` - Aplicación Android
- `web` - Web browser
- `b2b` - Integración B2B/API
- `unknown` - No se pudo determinar

**Detección automática (si no se proporciona header)**:
1. Analizar header `User-Agent`
2. Buscar keywords: "RealState-Mobile", "iOS", "Android", "Chrome", "Firefox", etc
3. Patrón: Si contiene "Mobile" pero no es browser → mobile
4. Default: "unknown"

**Ejemplos de User-Agent y detección**:
```
"RealState-Mobile/2.1.3 iOS/17.1" → mobile_ios
"RealState-Mobile/2.1.3 Android/13" → mobile_android
"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" → web
"grpc-go/1.50.0" → b2b (o explicit X-Client-Type: b2b)
```

---

### 5. X-Client-Version
**Propósito**: Versión del cliente (app o web build)  
**Tipo**: String  
**Requerido**: Recomendado  
**Ejemplo**:
- App: `2.1.3`
- Web: `web-build-2026.02.15` o `1.0.0-rc1`  
**Longitud máxima**: 32 caracteres

**Formato recomendado**:
- **Mobile apps**: `major.minor.patch` ej: `2.1.3`
- **Web**: `web-build-<date>` ej: `web-build-2026.02.15` o semantic version
- **B2B SDK**: `sdk-<version>` ej: `sdk-1.0.0`

**Validación del servidor**:
- Se registra como está sin validación adicional
- Útil para correlacionar bugs con versiones específicas

---

### 6. X-Device-Name
**Propósito**: Nombre descriptivo del dispositivo (para UX y debugging)  
**Tipo**: String  
**Requerido**: No  
**Ejemplo**:
- `iPhone 14 Pro`
- `Samsung Galaxy S24`
- `Chrome on Mac`  
**Longitud máxima**: 64 caracteres

**Detección automática (si no se proporciona)**:
- iOS: Extrae de User-Agent o patrón conocido
- Android: Extrae modelo del User-Agent
- Web: Extrae browser + OS del User-Agent

**Ejemplos de detección**:
```
User-Agent: "Apple/2.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X)"
→ X-Device-Name: "iPhone 14" (detectado automáticamente)

User-Agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
→ X-Device-Name: "Windows PC" (detectado automáticamente)
```

---

### 7. X-Forwarded-For (HTTP estándar)
**Propósito**: IP del cliente original (cuando hay proxy/load balancer)  
**Tipo**: IP address  
**Requerido**: Automático (si existe proxy)  
**Ejemplo**: `203.0.113.45, 198.51.100.20`  
**Nota**: Se envía automáticamente por proxies. El servidor toma la primera IP

---

### 8. User-Agent (HTTP estándar)
**Propósito**: Información del cliente (navegador, versión, SO)  
**Tipo**: String  
**Requerido**: Automático (enviado por HTTP)  
**Ejemplos**:
```
RealState-Mobile/2.1.3 iOS/17.1
RealState-Mobile/2.1.3 Android/13
Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36
```

**Recomendación para mobile apps**:
```
Format: <AppName>/<version> <OS>/<version>

Examples:
- RealState-Mobile/2.1.3 iOS/17.1
- RealState-Mobile/2.1.3 Android/13
- RealStateApp/1.0.0 iOS/16.5
```

---

### 9. Authorization (HTTP estándar)
**Propósito**: Bearer token JWT para autenticación  
**Tipo**: String  
**Formato**: `Bearer <token>`  
**Requerido**: Para endpoints protegidos  

**Contenido del JWT**:
```json
{
  "sub": "user_12345",  // user_id
  "iat": 1708010400,
  "exp": 1708014000,
  "jti": "session_xyz789"  // session_id para logs
}
```

---

## Header Priority Matrix

Cuando hay múltiples fuentes de información, el servidor aplica esta prioridad:

### Para X-Client-ID
| Prioridad | Fuente | Fallback |
|-----------|--------|----------|
| 1 | Header `X-Client-ID` | ← Si está presente, usar |
| 2 | JWT `sub` (si autenticado) | ← Si no header, extraer del token |
| 3 | Header `X-Device-ID` | ← Si no auth, usar device |
| 4 | IP + UA hash | ← Último recurso |

### Para X-Device-ID
| Prioridad | Fuente | Fallback |
|-----------|--------|----------|
| 1 | Header `X-Device-ID` | ← Si está presente, usar |
| 2 | UA + IP fingerprint | ← Si no header, generar |
| 3 | IP hash | ← Último recurso |

### Para X-Client-Type
| Prioridad | Fuente | Método |
|-----------|--------|--------|
| 1 | Header `X-Client-Type` | ← Más específico |
| 2 | User-Agent pattern | ← Contiene "RealState-Mobile" |
| 3 | Browser keywords | ← Contiene "Chrome", "Safari" |
| 4 | Mobile keywords | ← Contiene "iPhone", "Android" |
| 5 | Default | ← "unknown" |

---

## Complete Request Examples

### Example 1: Mobile iOS App (Authenticated)

```http
POST /properties HTTP/1.1
Host: api.realstate.com
Content-Type: application/json
User-Agent: RealState-Mobile/2.1.3 iOS/17.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Request-ID: req_550e8400-e29b-41d4-a716-446655440000
X-Client-ID: user_12345
X-Device-ID: uuid-iphone-f47ac10b-58cc-4372-a567-0e02b2c3d479
X-Client-Type: mobile_ios
X-Client-Version: 2.1.3
X-Device-Name: iPhone 14 Pro

{
  "name": "Luxury Penthouse",
  "price": 500000,
  "currency": "USD"
}
```

**Resultado en logs**:
```json
{
  "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
  "client_id": "user_12345",
  "client_type": "mobile_ios",
  "device_id": "uuid-iphone-f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "device_name": "iPhone 14 Pro",
  "user_id": "user_12345",
  "session_id": "jti_from_jwt",
  "client_version": "2.1.3",
  "origin": "203.0.113.45"
}
```

---

### Example 2: Web App (No Auth)

```http
POST /login HTTP/1.1
Host: api.realstate.com
Content-Type: application/json
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36
X-Client-Type: web
X-Device-ID: fingerprint_chrome_windows_a1b2c3d4e5f6
X-Client-Version: web-build-2026.02.15

{
  "email": "user@example.com",
  "password": "secret123"
}
```

**Resultado en logs**:
```json
{
  "request_id": "req_auto_generated_uuid",  // ← Generado por servidor
  "client_id": "client_chrome_windows_a1b2c3d4e5f6",  // ← Hash IP+UA
  "client_type": "web",
  "device_id": "fingerprint_chrome_windows_a1b2c3d4e5f6",
  "device_name": "Chrome on Windows",  // ← Detectado
  "client_version": "web-build-2026.02.15",
  "origin": "198.51.100.22"
}
```

---

### Example 3: B2B Integration (API Key)

```http
POST /properties HTTP/1.1
Host: api.realstate.com
Content-Type: application/json
Authorization: Bearer api_key_acme_corp_12345
X-Client-Type: b2b
X-Client-ID: integrator_acme_corp
X-Device-ID: api_key_acme_corp_12345
X-Client-Version: sdk-1.0.0

{
  "name": "Commercial Property",
  "price": 1000000,
  "currency": "USD"
}
```

**Resultado en logs**:
```json
{
  "request_id": "req_auto_generated_uuid",
  "client_id": "integrator_acme_corp",
  "client_type": "b2b",
  "device_id": "api_key_acme_corp_12345",
  "client_version": "sdk-1.0.0",
  "origin": "198.51.100.45"
}
```

---

## Validation Rules

### X-Request-ID
- ✅ Válido: `req_550e8400-e29b-41d4-a716-446655440000`
- ✅ Válido: `550e8400-e29b-41d4-a716-446655440000`
- ❌ Inválido: `<script>alert('xss')</script>` (caracteres especiales)
- ❌ Inválido: `;drop table users` (caracteres especiales)

**Servidor**:
- Si formato inválido, regenera con `req_<new-uuid>`
- Máximo 36 caracteres, alphanumeric + hyphen

### X-Client-ID
- ✅ Válido: `user_12345`, `integrator_acme_corp`, `app_mobile_ios`
- ❌ Inválido: `user@123`, `client#1` (caracteres especiales)
- ❌ Inválido: `" or 1=1` (SQL injection attempt)

**Servidor**:
- Whitelist: `[a-zA-Z0-9_-]` solamente
- Si contiene caracteres inválidos, rechaza o limpia
- Máximo 64 caracteres

### X-Device-ID
- ✅ Válido: `f47ac10b-58cc-4372-a567-0e02b2c3d479` (UUID)
- ✅ Válido: `864507030695220` (IMEI)
- ✅ Válido: `fingerprint_a1b2c3d4e5f6g7h8`
- ❌ Inválido: Excede 128 caracteres

**Servidor**:
- Acepta como está, no valida formato específico
- Máximo 128 caracteres

---

## Response Headers

El servidor retorna estos headers en la respuesta:

```http
HTTP/1.1 200 OK
X-Request-ID: req_550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json

{
  "status": "success",
  "code": 200,
  "data": { ... },
  "meta": {
    "request_id": "req_550e8400-e29b-41d4-a716-446655440000",
    "server_time": "2026-02-15T14:35:22Z"
  }
}
```

**Nota**: El `request_id` se retorna tanto en header como en JSON body para comodidad del cliente

---

## Best Practices

### ✅ DO

1. **Enviar X-Request-ID si es posible** para trazabilidad de lado del cliente
2. **Enviar X-Device-ID para mobile** (persiste durante sesión)
3. **Incluir User-Agent descriptivo** en mobile apps
4. **Mantener X-Device-ID constante** para cada dispositivo
5. **Enviar X-Client-Version** para correlacionar bugs

### ❌ DON'T

1. **No usar caracteres especiales** en X-Client-ID
2. **No cambiar X-Device-ID** entre requests (deve persistir)
3. **No enviar información sensible** en headers (use JWT)
4. **No exceder longitud máxima** de headers
5. **No falsificar X-Client-Type** (el servidor lo detectará)

---

## Implementation Checklist

- [ ] **Mobile iOS**
  - [ ] Set correct User-Agent format
  - [ ] Generate and persist X-Device-ID
  - [ ] Include X-Client-Version
  - [ ] Send X-Client-Type header
  - [ ] Test with staging API

- [ ] **Mobile Android**
  - [ ] Set correct User-Agent format
  - [ ] Generate and persist X-Device-ID
  - [ ] Include X-Client-Version
  - [ ] Send X-Client-Type header
  - [ ] Test with staging API

- [ ] **Web App**
  - [ ] Implement device fingerprinting
  - [ ] Persist X-Device-ID in localStorage
  - [ ] Add X-Client-Type: web header
  - [ ] Include X-Client-Version from build
  - [ ] Test across browsers

- [ ] **B2B Integration**
  - [ ] Set X-Client-Type: b2b
  - [ ] Assign unique X-Client-ID
  - [ ] Assign unique X-Device-ID
  - [ ] Include X-Client-Version
  - [ ] Document in SDK

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-15 | Initial specification |

---

## Contact

Para preguntas sobre esta especificación, contactar al equipo de backend SRE.

**Status**: Active ✅  
**Last Updated**: February 15, 2026
