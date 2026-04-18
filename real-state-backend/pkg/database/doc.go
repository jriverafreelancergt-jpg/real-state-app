// Package database proporciona utilitarios para manejo estructurado y logging de errores de base de datos.
//
// # Características principales
//
// - Clasificación automática de errores de BD (conexión, timeout, integridad, etc.)
// - Logging estructurado con contexto relevante
// - Funciones helper para verificar tipos de errores específicos
//
// # Ejemplo de uso en repositorios
//
// Antes (sin logging mejorado):
//
//	func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
//		user := &domain.User{}
//		err := r.db.QueryRowContext(ctx, query, id).Scan(...)
//		if err != nil {
//			return nil, err  // ❌ Sin contexto, sin clasificación
//		}
//		return user, nil
//	}
//
// Después (con logging mejorado):
//
//	func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
//		user := &domain.User{}
//		err := r.db.QueryRowContext(ctx, query, id).Scan(...)
//		if err != nil {
//			// ✅ Con contexto, clasificación y logging automático
//			return nil, database.HandleError(ctx, err, "GetByID", "users", map[string]interface{}{"id": id})
//		}
//		return user, nil
//	}
//
// # Salida de logs
//
// Con contexto, el sistema genera logs como estos:
//
// Error de conexión (CRITICAL):
//
//	{
//	  "time": "2026-02-14T20:31:14Z",
//	  "level": "ERROR",
//	  "msg": "Database error in GetByID on table 'users': connection refused (database server not accessible)",
//	  "error_type": "connection_refused",
//	  "operation": "GetByID",
//	  "table": "users",
//	  "original_error": "connection refused",
//	  "context_fields": {"id": "user123"}
//	}
//
// Error de no encontrado (INFO):
//
//	{
//	  "time": "2026-02-14T20:31:14Z",
//	  "level": "INFO",
//	  "msg": "Database error in GetByID on table 'users': no rows found",
//	  "error_type": "no_rows",
//	  "operation": "GetByID",
//	  "table": "users",
//	  "original_error": "sql: no rows in result set",
//	  "context_fields": {"id": "user123"}
//	}
//
// Error de timeout (WARN):
//
//	{
//	  "time": "2026-02-14T20:31:14Z",
//	  "level": "WARN",
//	  "msg": "Database error in GetByID on table 'users': operation timeout (query took too long)",
//	  "error_type": "operation_timeout",
//	  "operation": "GetByID",
//	  "table": "users",
//	  "original_error": "context deadline exceeded",
//	  "context_fields": {"id": "user123"}
//	}
//
// # Funciones helper
//
// Para verificar tipos específicos de errores en niveles superiores:
//
//	if err != nil {
//		if database.IsConnectionError(err) {
//			// Reintentar con backoff
//			return retryWithBackoff()
//		}
//		if database.IsNotFound(err) {
//			// Retornar 404
//			return http.StatusNotFound
//		}
//		if database.IsTimeout(err) {
//			// Retornar 504 (Service Unavailable)
//			return http.StatusServiceUnavailable
//		}
//	}
//
// # Tipos de errores clasificados
//
// Errores de conexión:
//   - ErrTypeConnectionRefused: Servidor BD no accesible
//   - ErrTypeConnectionTimeout: Timeout en la conexión
//   - ErrTypeConnectionLost: Conexión perdida durante operación
//   - ErrTypePoolExhausted: Pool de conexiones agotado
//
// Errores de operación:
//   - ErrTypeNoRows: Sin resultados (no_rows)
//   - ErrTypeTimeout: Timeout de operación (query lenta)
//   - ErrTypeContextCanceled: Contexto cancelado
//   - ErrTypeTxDone: Transacción ya completada
//
// Errores de integridad:
//   - ErrTypeUniqueViolation: Violación de constraint único
//   - ErrTypeForeignKeyViolation: Violación de FK
//   - ErrTypeCheckViolation: Violación de check constraint
//
// # Best Practices
//
// 1. Siempre pasar contexto: HandleError requiere contexto para logs precisos
// 2. Incluir campos relevantes: Pasar datos útiles en el mapa (id, username, etc.)
// 3. Especificar operación: Indicar el nombre del método (GetByID, Create, etc.)
// 4. Pasar tabla correcta: Nombre de la tabla para mejor debugging
// 5. En handlers HTTP: Usar IsConnectionError, IsTimeout, IsNotFound para mapear a códigos HTTP
//
// # Integración en handlers HTTP
//
// Ejemplo de handler que usa el nuevo sistema:
//
//	func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
//		userID := mux.Vars(r)["id"]
//
//		user, err := h.service.GetByID(r.Context(), userID)
//		if err != nil {
//			// El error ya tiene logging automático de BD
//			if database.IsNotFound(err) {
//				writeError(w, http.StatusNotFound, "User not found", "not_found")
//				return
//			}
//			if database.IsConnectionError(err) {
//				writeError(w, http.StatusServiceUnavailable, "Database unavailable", "service_unavailable")
//				return
//			}
//			writeError(w, http.StatusInternalServerError, "Internal server error", "internal_error")
//			return
//		}
//
//		writeJSON(w, http.StatusOK, user)
//	}
//
// # Testing
//
// Ejemplo de test unitario verificando manejo de errores:
//
//	func TestGetUserDatabaseError(t *testing.T) {
//		ctx := context.Background()
//		// Mock repository que retorna error de conexión
//		repo := &MockUserRepository{
//			GetByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
//				return nil, database.HandleError(ctx, sql.ErrNoRows, "GetByID", "users", map[string]interface{}{"id": id})
//			},
//		}
//
//		user, err := repo.GetByID(ctx, "user123")
//		assert.Nil(t, user)
//		assert.NotNil(t, err)
//		assert.True(t, database.IsNotFound(err))
//	}
//
// # Evolución futura
//
// Consideraciones para mejoras:
//   - Agregar métricas de Prometheus para errores por tipo
//   - Implementar circuit breaker automático para errores de conexión
//   - Rastreo distribuido (distributed tracing) con IDs de correlación
//   - Alertas automáticas para errores críticos (conexión rechazada, pool agotado)
//   - Rollback automático en transacciones fallidas
package database
