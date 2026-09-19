# Arquitectura Técnica: Backend Realty Astarac

Este documento presenta la visión arquitectónica y técnica del backend para la aplicación inmobiliaria Realty Astarac. Diseñado bajo estrictos principios de Ingeniería de Software, prioriza el rendimiento, la escalabilidad y una **seguridad de grado bancario**.

---

## 🏗️ 1. Patrón Arquitectónico: Clean Architecture (Hexagonal)

El sistema está estructurado utilizando **Clean Architecture** (Arquitectura Hexagonal o de Puertos y Adaptadores). El objetivo principal es mantener el núcleo del negocio (dominio) completamente agnóstico de tecnologías externas (frameworks, bases de datos o APIs de terceros).

### Capas del Sistema:

1. **Dominio (Entities/Models)**: 
   - Contiene la lógica de negocio pura y los modelos de datos (Ej: `User`, `Property`, `Session`).
   - No tiene dependencias hacia ninguna otra capa.

2. **Casos de Uso (Services)**:
   - Orquesta la lógica de la aplicación.
   - Define interfaces (Puertos) para comunicarse con el exterior, como la base de datos o APIs de terceros.

3. **Adaptadores (Repositories & Handlers)**:
   - **Capa de Entrada (Handlers/Controllers)**: Recibe peticiones HTTP, parsea JSON y pasa la ejecución a los Casos de Uso.
   - **Capa de Salida (Repositories)**: Implementaciones concretas de las interfaces del dominio para acceder a PostgreSQL. Si la DB cambia, el dominio no se ve afectado.

---

## 🛡️ 2. Seguridad de Grado Bancario

La aplicación procesa datos sensibles de usuarios e inmobiliarias, por lo cual implementa múltiples anillos de seguridad:

### A. Autenticación y Criptografía
- **Hashing Robusto**: Uso de algoritmos resistentes a ataques de fuerza bruta y GPUs, específicamente **Argon2id** (estándar OWASP actual) o **Bcrypt** con work factors altos.
- **Gestión de Sesiones (JWT)**: 
  - Manejo de **Access Tokens** de corta duración y **Refresh Tokens** de larga duración.
  - Implementación de **JTI (JWT ID)** almacenado en la tabla de sesiones para permitir la **revocación inmediata** de tokens en caso de compromisos o cierres de sesión remotos.
  - Almacenamiento seguro en el cliente (Flutter) vía Secure Storage.

### B. Identificación y Control de Dispositivos
- **Fingerprinting de Clientes**: Rastreo avanzado de dispositivos y clientes para detectar anomalías (ubicaciones inusuales, cambios drásticos de dispositivo).
- Bloqueo preventivo tras múltiples intentos de login fallidos (`UpdateFailedAttempts`) y flujos para MFA (Multi-Factor Authentication).

### C. Auditoría y Trazabilidad Estricta
- **Logs Estructurados y Seguros**: Registro de logs en formato JSON (`JSON_LOG_STRUCTURE`). 
- **Prevención de Fuga de Datos**: El sistema de logging captura el contexto completo de errores en la base de datos (queries lentas, timeouts, etc.) pero **jamás** expone esta información sensible hacia el cliente final (ver `pkg/database/errors.go`).

---

## 🔄 3. Integración con Wasi.co

Dado que `wasi.co` es el núcleo operativo de las propiedades de la inmobiliaria, la integración está **completamente implementada** y operativa, con resiliencia de nivel enterprise:

### Estado Actual de Implementación:

1. **Sincronización Concurrente (100% operativa)**
   - **Worker Pool**: 5 workers configurables descargando propiedades en paralelo (ver `PropertySyncService`)
   - **Paginación Inteligente**: Descubre automáticamente el total de páginas en Wasi y distribuye la carga de trabajo
   - **Batch Tracking**: Cada sincronización genera un `batch_id` (UUID) para auditoría y trazabilidad
   - **Cancelación Dinámica**: Si un worker falla, todos los demás se cancelan automáticamente mediante `context.WithCancel()`
   - Tests unitarios: ✅ `TestSyncAllProperties_Success`, `TestSyncAllProperties_FetcherFailureAbortsSweep` (PASS)

2. **Resiliencia Empresarial (100% operativa)**
   - **Circuit Breaker**: Protege contra cascadas de fallos con 3 estados (Closed, Open, HalfOpen)
     - Abre después de N fallos consecutivos transitorios
     - Reabre automáticamente después de cooldown (configurable)
   - **Backoff Exponencial**: Reintentos con espera creciente en caso de errores transitorios (429, 5xx, timeouts)
   - **Detección de Errores**: Diferencia errores transitorios (redes, límites de tasa) de permanentes
   - Tests unitarios: ✅ `TestCircuitBreakerStates`, `TestClientExponentialBackoffAndTimeout` (PASS)

3. **Mapeo Anti-Corrupción (100% operativa)**
   - **WasiAdapter**: Aísla el contrato de Wasi.co del dominio mediante clase adaptadora
   - **Transformación de Datos**: 
     - Mapeo de campos: `IDProperty` → `ID`, `NamePropertyType` → `Type`, `Comment` → `Description`
     - Normalización de monedas: `IDCurrency` (código numérico Wasi) → `["USD", "GTQ"]` (estándar app)
     - Selección inteligente de imágenes: Prioriza `MainImage.Original`, fallback a `MainImage.URL`
   - **Protección del Dominio**: Si Wasi.co cambia su API, solo el adaptador necesita actualización
   - Tests unitarios: ✅ `TestUnmarshalWasiPropertyResponse`, `TestWasiAdapterMapping` (PASS)

4. **Endpoint de Sincronización Manual**
   - **Ruta**: `POST /v1/admin/sync` (protegida por JWT + RBAC)
   - **Uso**: Trigger manual de sincronización en caso de desfase o auditoría
   - Implementación: `SyncHandler.TriggerSync()`

5. **Caché Local - Estrategia de Persistencia (100% operativa)**
   
   **Arquitectura de Almacenamiento:**
   - **BD PostgreSQL como Caché Primaria**: Todas las propiedades de Wasi se persisten via `PropertyRepository.UpsertBatch()` usando UPSERT (INSERT ... ON CONFLICT)
   - **Rastreo de Sincronización**: Campo `last_sync_id` (batch UUID) permite auditar qué propiedades vinieron en cada sincronización
   - **Versionamiento de Estado**: Campo `origin = 'WASI'` + `status = 'ACTIVE'/'INACTIVE'` para diferenciar propiedades vivas vs eliminadas
   
   **Patrón de Reconciliación (Sweep):**
   - **Operación Atómica**: Después de cada sincronización exitosa, `PropertyRepository.Sweep()` desactiva propiedades no presentes en lote actual
   - **Transacciones DB**: `UpsertBatch()` envuelve operaciones en transacción para garantizar consistencia (rollback automático en fallos)
   - **Queryable History**: Propiedades `INACTIVE` se conservan en BD para auditoría; no se eliminan
   
   **Modo Degradado (Fallback Automático):**
   - **Sin Lógica de Switch Manual**: Los handlers (`PropertyHandler.GetAll()`) siempre consultan BD local primero
   - **Insensibilidad a Fallos de Wasi**: Si `SyncAllProperties` falla, app móvil sigue operativa con datos del último sync exitoso
   - **Recuperación Automática**: Próximo trigger manual (`POST /v1/admin/sync`) o worker automático reintentar sincronización con circuit breaker reabierto
   
   **Limitaciones Arquitectónicas Actuales (Considerar en Evolución):**
   - ⚠️ **Sin Caché en Memoria (Redis)**: Solo BD PostgreSQL como caché (latencia ~5-15ms por query)
   - ⚠️ **Sin Timestamp de Última Sincronización**: No se trackea `last_successful_sync_at` globalmente (necesario para UX de "sync status")
   - ⚠️ **Sin Validación de Datos Obsoletos**: La app móvil ignora si datos tienen >N horas sin sincronizar (silently stale)
   
   **Mejoras Recomendadas (Roadmap P2):**
   - Agregar columna `last_successful_sync_at` a tabla `properties` y tabla de sistema para tracking global
   - Implementar Redis como caché L1 en frente a BD (queries frecuentes a `GetAll()` benefician latencia)
   - Endpoint `GET /v1/sync-status` que exponga timestamp de última sincronización para UX de "data freshness"

### Integración en el Sistema:
- **Cliente HTTP**: `internal/wasi/client.go` - Manejo de autenticación (Token), timeouts, reintentos
- **Almacenamiento**: `PropertyRepository.SaveBatchProperties()` - Persist local de propiedades descargadas
- **Orquestación**: `main.go` - Inicializa WASI client con credenciales de `config.WasiBaseURL`, `config.WasiToken`, `config.WasiCompanyID`
- **Observabilidad**: Logs estructurados JSON en cada fase (inicio sync, páginas descubiertas, desactivaciones, fallos)

---

## 🚀 4. Infraestructura y Despliegue (DevOps)

- **Contenedores (Docker)**: La aplicación y sus dependencias (PostgreSQL, Redis si aplica) corren sobre Docker, garantizando paridad total entre desarrollo, staging y producción (`Dockerfile`, `docker-compose.yml`, scripts de control).
- **Migraciones Controladas**: Gestión versionada de la base de datos (`golang-migrate/migrate`), permitiendo auditar la evolución de las tablas y realizar rollbacks exactos.
- **Observabilidad**: Sistema preparado para conectarse a Prometheus/Grafana (métricas) y OpenTelemetry (trazas distribuidas), gracias al uso avanzado del paquete `context` y logging centralizado.

---

## 🎯 Conclusión

El backend de **Realty Astarac** no es un CRUD convencional. Es un microservicio robusto en Golang diseñado para ser:
- **Altamente Testeable** (por la inyección de dependencias).
- **Inquebrantable** (manejo inteligente de errores y seguridad por diseño).
- **Escalable** (preparado para soportar miles de peticiones móviles concurrentes con un mínimo consumo de RAM).
