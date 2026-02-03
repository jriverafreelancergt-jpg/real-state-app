# Instrucciones para Agentes de Codificación AI en el Backend de Bienes Raíces

## Resumen de Arquitectura
Este backend en Go sigue **Arquitectura Limpia** con separación estricta de capas:
- `internal/core/domain/`: Entidades de negocio (ej. struct `Property`)
- `internal/core/ports/`: Interfaces que definen contratos (`PropertyRepository`, `PropertyService`)
- `internal/repository/`: Implementaciones PostgreSQL
- `internal/services/`: Capa de lógica de negocio
- `internal/handlers/`: Manejo de solicitudes/respuestas HTTP
- `internal/dto/`: Objetos de Transferencia de Datos con validación (ej. `CreatePropertyDTO.Validate()`)

**¿Por qué esta estructura?** Permite testabilidad, inversión de dependencias e integración con app móvil con límites claros.

## Patrones y Convenciones Clave
- **Uso de contexto**: Todos los métodos de servicio/repositorio aceptan `context.Context` para timeouts y cancelación
- **Validación de DTO**: Usar método `Validate()` en DTOs (no `IsValid()`) - retorna mensajes de error específicos
- **Manejo de moneda**: Solo permitir `["USD", "GTQ"]` en `CreatePropertyDTO`
- **Respuestas de error**: Formato JSON con `{"error": "mensaje"}` para errores amigables al cliente
- **Logging**: Usar `slog` con salida estructurada JSON (ej. `slog.Info("mensaje", "clave", valor)`)
- **Base de datos**: PostgreSQL con consultas parametrizadas; usar `QueryRowContext`/`QueryContext` para conciencia de contexto

## Flujo de Desarrollo
- **Desarrollo local**: Ejecutar `air` (recarga en caliente) vía Docker: `docker-compose up --build`
- **Configuración de BD**: Ejecutar migraciones con `migrate -path ./migrations -database $DATABASE_URL up`
- **Build**: `go build -o tmp/main cmd/api/main.go`
- **Config**: Variables de entorno (ver `config/config.go`); defaults en `.env` para dev local

## Puntos de Integración
- **Base de datos**: Pool de conexiones PostgreSQL configurado en `main.go` (máx 25 conexiones)
- **Servidor HTTP**: Librería estándar con middleware de seguridad (`pkg/middleware/security.go`)
- **App móvil**: Espera respuestas JSON; DTOs mapean input móvil a objetos de dominio

## Tareas Comunes
- **Agregar nuevo endpoint**: Crear método handler, agregar ruta en `main.go`, implementar service/repo si es necesario
- **Cambios en BD**: Crear archivos de migración en `migrations/` (up/down SQL)
- **Validación**: Extender `Validate()` en DTOs; verificar `slices.Contains` para valores permitidos
- **Testing**: Enfocarse en tests unitarios de capa service; mockear repositorios usando interfaces

## Ejemplos de Archivos
- **Patrón handler**: Ver `internal/handlers/property_handler.go` - decodificar JSON, validar DTO, llamar service
- **Lógica service**: `internal/services/property_service.go` - reglas de negocio antes de llamadas repo
- **Consultas repository**: `internal/repository/property_repository.go` - SQL parametrizado con contexto</content>
<parameter name="filePath">/home/jrivera/work/real-state-app/real-state-backend/.github/copilot-instructions.md