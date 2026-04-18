package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	contextpkg "real-state-backend/pkg/context"
)

// DBErrorType define el tipo de error de base de datos
type DBErrorType string

const (
	// Errores de conexión
	ErrTypeConnectionRefused DBErrorType = "connection_refused"
	ErrTypeConnectionTimeout DBErrorType = "connection_timeout"
	ErrTypeConnectionLost    DBErrorType = "connection_lost"
	ErrTypePoolExhausted     DBErrorType = "pool_exhausted"

	// Errores de operación
	ErrTypeNotFound        DBErrorType = "not_found"
	ErrTypeNoRows          DBErrorType = "no_rows"
	ErrTypeTxDone          DBErrorType = "transaction_done"
	ErrTypeTimeout         DBErrorType = "operation_timeout"
	ErrTypeContextCanceled DBErrorType = "context_canceled"

	// Errores de integridad
	ErrTypeUniqueViolation     DBErrorType = "unique_violation"
	ErrTypeForeignKeyViolation DBErrorType = "foreign_key_violation"
	ErrTypeCheckViolation      DBErrorType = "check_violation"

	// Errores generales
	ErrTypeUnknown DBErrorType = "unknown"
)

// DBError encapsula un error de base de datos con metadatos
type DBError struct {
	ErrorType   DBErrorType
	Message     string
	Operation   string                 // ej: "GetByID", "Create", "Update"
	Table       string                 // ej: "users", "properties"
	Err         error                  // Error original
	Fields      map[string]interface{} // Contexto adicional
	Diagnostic  string                 // Diagnóstico detallado del error
	Remediation string                 // Acciones recomendadas
}

func (e *DBError) Error() string {
	return e.Message
}

func (e *DBError) Unwrap() error {
	return e.Err
}

// ClassifyError clasifica el error de BD y devuelve el tipo
func ClassifyError(err error) DBErrorType {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Errores de contexto
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTypeTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrTypeContextCanceled
	}

	// Errores específicos de sql
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTypeNoRows
	}
	if errors.Is(err, sql.ErrTxDone) {
		return ErrTypeTxDone
	}

	// Errores DNS (host lookup failed, server misbehaving, etc.)
	// Estos indican que el host no existe o no se puede resolver
	if strings.Contains(errStr, "server misbehaving") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "lookup") && strings.Contains(errStr, "refused") ||
		strings.Contains(errStr, "lookup") && strings.Contains(errStr, "no such") {
		return ErrTypeConnectionRefused
	}

	// Errores PostgreSQL
	if strings.Contains(errStr, "connection refused") {
		return ErrTypeConnectionRefused
	}
	if strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "timeout") {
		return ErrTypeConnectionTimeout
	}
	if strings.Contains(errStr, "connection lost") || strings.Contains(errStr, "broken pipe") {
		return ErrTypeConnectionLost
	}

	// Errores de integridad (PostgreSQL)
	if strings.Contains(errStr, "unique constraint") {
		return ErrTypeUniqueViolation
	}
	if strings.Contains(errStr, "foreign key constraint") {
		return ErrTypeForeignKeyViolation
	}
	if strings.Contains(errStr, "check constraint") {
		return ErrTypeCheckViolation
	}

	// Errores de pool
	if strings.Contains(errStr, "pool") || strings.Contains(errStr, "no more connections available") {
		return ErrTypePoolExhausted
	}

	return ErrTypeUnknown
}

// HandleError maneja y loguea un error de BD de manera estructurada
func HandleError(ctx context.Context, err error, operation, table string, fields map[string]interface{}) error {
	if err == nil {
		return nil
	}

	errorType := ClassifyError(err)
	errStr := err.Error()

	dbErr := &DBError{
		ErrorType: errorType,
		Operation: operation,
		Table:     table,
		Err:       err,
		Fields:    fields,
	}

	// Construir mensaje basado en el tipo de error
	dbErr.Message = buildErrorMessage(errorType, operation, table)

	// Generar diagnóstico detallado basado en el tipo de error
	dbErr.Diagnostic = generateDiagnostic(errorType, errStr)
	dbErr.Remediation = generateRemediation(errorType)

	// Loguear error estructurado
	logDBError(ctx, dbErr)

	return dbErr
}

// buildErrorMessage construye un mensaje descriptivo del error
func buildErrorMessage(errType DBErrorType, operation, table string) string {
	baseMsg := fmt.Sprintf("Database error in %s on table '%s'", operation, table)

	switch errType {
	case ErrTypeConnectionRefused:
		return fmt.Sprintf("%s: connection refused (database server not accessible)", baseMsg)
	case ErrTypeConnectionTimeout:
		return fmt.Sprintf("%s: connection timeout (database server unreachable)", baseMsg)
	case ErrTypeConnectionLost:
		return fmt.Sprintf("%s: connection lost (network issue)", baseMsg)
	case ErrTypePoolExhausted:
		return fmt.Sprintf("%s: connection pool exhausted (too many concurrent requests)", baseMsg)
	case ErrTypeNoRows:
		return fmt.Sprintf("%s: no rows found", baseMsg)
	case ErrTypeTimeout:
		return fmt.Sprintf("%s: operation timeout (query took too long)", baseMsg)
	case ErrTypeContextCanceled:
		return fmt.Sprintf("%s: operation canceled", baseMsg)
	case ErrTypeUniqueViolation:
		return fmt.Sprintf("%s: unique constraint violation", baseMsg)
	case ErrTypeForeignKeyViolation:
		return fmt.Sprintf("%s: foreign key constraint violation", baseMsg)
	case ErrTypeCheckViolation:
		return fmt.Sprintf("%s: check constraint violation", baseMsg)
	default:
		return fmt.Sprintf("%s: unknown error", baseMsg)
	}
}

// generateDiagnostic genera un diagnóstico detallado del error
func generateDiagnostic(errType DBErrorType, originalError string) string {
	switch errType {
	case ErrTypeConnectionRefused:
		if strings.Contains(originalError, "server misbehaving") {
			return "DNS resolution failed: the database hostname could not be resolved. " +
				"The container might be down, the hostname is incorrect, or there's a network issue."
		}
		if strings.Contains(originalError, "no such host") {
			return "DNS error: hostname does not exist. Verify that the DATABASE_URL has the correct hostname and that the database server is accessible."
		}
		return "Connection refused: the database server is not accepting connections. " +
			"Possible causes: server is down, wrong port, firewall blocking connection."

	case ErrTypeConnectionTimeout:
		return "Connection timeout: unable to reach the database server within the configured timeout. " +
			"The server might be overloaded, network is slow, or server is temporarily unavailable."

	case ErrTypeConnectionLost:
		return "Connection lost: the connection to the database was lost during operation. " +
			"This can happen due to network issues, server restart, or resource constraints."

	case ErrTypePoolExhausted:
		return "Connection pool exhausted: all available connections are in use. " +
			"Too many concurrent requests are trying to access the database simultaneously."

	case ErrTypeTimeout:
		return "Query execution timeout: the query took longer than expected to complete. " +
			"The query might be inefficient, or the database is under heavy load."

	case ErrTypeUniqueViolation:
		return "Unique constraint violation: trying to insert or update a value that must be unique. " +
			"A record with this value already exists in the database."

	case ErrTypeForeignKeyViolation:
		return "Foreign key constraint violation: trying to reference a record that doesn't exist. " +
			"The parent record in the referenced table must exist before creating this record."

	case ErrTypeCheckViolation:
		return "Check constraint violation: the data doesn't meet the validation requirements defined in the table. " +
			"Verify that the data matches the column constraints."

	case ErrTypeNoRows:
		return "No records found: the query executed successfully but returned no results. " +
			"This is normal when searching for non-existent data."

	case ErrTypeContextCanceled:
		return "Operation canceled: the request was canceled before completion. " +
			"This usually happens when the client disconnects or the operation times out."

	default:
		return fmt.Sprintf("Unknown error: %s", originalError)
	}
}

// generateRemediation genera acciones recomendadas para resolver el error
func generateRemediation(errType DBErrorType) string {
	switch errType {
	case ErrTypeConnectionRefused:
		return "1. Check if PostgreSQL container is running: docker ps | grep postgres\n" +
			"2. If not running, start it: docker start postgres_server_realstate\n" +
			"3. Verify DATABASE_URL in .env has correct hostname and port\n" +
			"4. Check network connectivity between application and database\n" +
			"5. Review database server logs for errors"

	case ErrTypeConnectionTimeout:
		return "1. Check database server status and resources (CPU, memory)\n" +
			"2. Verify network connectivity and latency\n" +
			"3. Increase connection timeout if appropriate\n" +
			"4. Check if server is overloaded with too many queries\n" +
			"5. Review slow query logs if available"

	case ErrTypeConnectionLost:
		return "1. Check network connectivity\n" +
			"2. Verify database server is still running\n" +
			"3. Review database logs for crashes or restarts\n" +
			"4. Implement retry logic with exponential backoff\n" +
			"5. Check for resource constraints on the server"

	case ErrTypePoolExhausted:
		return "1. Increase MaxOpenConns in db.SetMaxOpenConns()\n" +
			"2. Review application code for connection leaks\n" +
			"3. Monitor active connections with metrics\n" +
			"4. Implement connection pooling best practices\n" +
			"5. Consider database connection limits"

	case ErrTypeTimeout:
		return "1. Add indexes to frequently queried columns\n" +
			"2. Run EXPLAIN ANALYZE to review query plan\n" +
			"3. Optimize the query (reduce data transferred)\n" +
			"4. Increase query timeout if appropriate\n" +
			"5. Check database server resources (CPU, disk I/O)"

	case ErrTypeUniqueViolation:
		return "1. Validate data uniqueness before insert/update\n" +
			"2. Check for duplicate entries in database\n" +
			"3. Use INSERT ... ON CONFLICT for handling duplicates\n" +
			"4. Review application business logic\n" +
			"5. Implement proper error handling for this constraint"

	case ErrTypeForeignKeyViolation:
		return "1. Verify parent record exists before creating this record\n" +
			"2. Check foreign key references are correct\n" +
			"3. Review data relationships and dependencies\n" +
			"4. Use transactions to ensure data consistency\n" +
			"5. Add validation before database operations"

	case ErrTypeCheckViolation:
		return "1. Review check constraints defined in table schema\n" +
			"2. Validate data against constraints before insert/update\n" +
			"3. Check data types and value ranges\n" +
			"4. Add proper input validation in application\n" +
			"5. Review migration files for constraint definitions"

	case ErrTypeContextCanceled:
		return "1. This is usually normal behavior\n" +
			"2. Implement timeout handling in requests\n" +
			"3. Add context cancellation callbacks if needed\n" +
			"4. Monitor if this happens frequently\n" +
			"5. Increase timeout if legitimate operations are being canceled"

	default:
		return "1. Check database connectivity\n" +
			"2. Review error details carefully\n" +
			"3. Check application and database logs\n" +
			"4. Verify database server is running\n" +
			"5. Retry the operation after verifying the issue is resolved"
	}
}

// logDBError loguea un error de BD con contexto estructurado
// Incluye identificación de cliente/dispositivo si está disponible en el contexto
func logDBError(ctx context.Context, dbErr *DBError) {
	logArgs := []interface{}{
		"error_type", string(dbErr.ErrorType),
		"operation", dbErr.Operation,
		"table", dbErr.Table,
		"original_error", dbErr.Err.Error(),
		"diagnostic", dbErr.Diagnostic,
		"remediation", dbErr.Remediation,
	}

	// Agregar campos adicionales si existen
	if dbErr.Fields != nil && len(dbErr.Fields) > 0 {
		logArgs = append(logArgs, "context_fields", dbErr.Fields)
	}

	// Agregar información de cliente/dispositivo si está disponible en el contexto
	// Esto permite trazabilidad completa de errores
	if clientIdentity := contextpkg.FromContext(ctx); clientIdentity != nil && clientIdentity.IsValid() {
		clientFields := clientIdentity.LogFields()
		for key, value := range clientFields {
			logArgs = append(logArgs, key, value)
		}
	}

	// Determinar nivel de log según el tipo de error
	switch dbErr.ErrorType {
	case ErrTypeConnectionRefused, ErrTypeConnectionTimeout, ErrTypeConnectionLost, ErrTypePoolExhausted:
		// Errores críticos de conexión
		slog.Error(dbErr.Message, logArgs...)
	case ErrTypeTimeout:
		// Errores de timeout (warning, pueden ser transitorios)
		slog.Warn(dbErr.Message, logArgs...)
	case ErrTypeNoRows:
		// No es realmente un error, pero lo registramos como info
		slog.Info(dbErr.Message, logArgs...)
	default:
		// Otros errores como error
		slog.Error(dbErr.Message, logArgs...)
	}
}

// HandleErrorWithContext es una versión simplificada sin contexto personalizado
func HandleErrorSimple(err error, operation, table string) error {
	if err == nil {
		return nil
	}
	return HandleError(context.Background(), err, operation, table, nil)
}

// IsConnectionError verifica si es un error de conexión
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errType := ClassifyError(err)
	return errType == ErrTypeConnectionRefused ||
		errType == ErrTypeConnectionTimeout ||
		errType == ErrTypeConnectionLost ||
		errType == ErrTypePoolExhausted
}

// IsNotFound verifica si es un error de "no encontrado"
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	errType := ClassifyError(err)
	return errType == ErrTypeNoRows || errType == ErrTypeNotFound
}

// IsTimeout verifica si es un error de timeout
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	errType := ClassifyError(err)
	return errType == ErrTypeTimeout || errType == ErrTypeConnectionTimeout
}

// IsIntegrityError verifica si es un error de integridad
func IsIntegrityError(err error) bool {
	if err == nil {
		return false
	}
	errType := ClassifyError(err)
	return errType == ErrTypeUniqueViolation ||
		errType == ErrTypeForeignKeyViolation ||
		errType == ErrTypeCheckViolation
}
