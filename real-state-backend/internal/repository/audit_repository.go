package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"real-state-backend/internal/core/domain"
	contextpkg "real-state-backend/pkg/context"
	"real-state-backend/pkg/database"

	"github.com/google/uuid"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) LogEvent(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, event_type, user_id, resource, action, old_values, new_values, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	log.ID = uuid.New().String()

	oldJSON, _ := json.Marshal(log.OldValues)
	newJSON, _ := json.Marshal(log.NewValues)

	// Si IPAddress está vacío, intentar obtenerla del contexto ClientIdentity
	ipAddress := log.IPAddress
	if ipAddress == "" {
		if clientIdentity := contextpkg.FromContext(ctx); clientIdentity != nil {
			ipAddress = clientIdentity.Origin
		}
	}
	// Si aún está vacío, usar NULL (en lugar de cadena vacía) para el tipo inet
	var ipAddressParam interface{} = ipAddress
	if ipAddress == "" {
		ipAddressParam = nil
	}

	_, err := r.db.ExecContext(ctx, query, log.ID, log.EventType, log.UserID, log.Resource, log.Action,
		oldJSON, newJSON, ipAddressParam, log.UserAgent, log.Timestamp)
	if err != nil {
		return database.HandleError(ctx, err, "LogEvent", "audit_logs", map[string]interface{}{"event_type": log.EventType, "resource": log.Resource, "action": log.Action})
	}
	return nil
}
