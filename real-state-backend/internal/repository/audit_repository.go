package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"real-state-backend/internal/core/domain"

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

	_, err := r.db.ExecContext(ctx, query, log.ID, log.EventType, log.UserID, log.Resource, log.Action,
		oldJSON, newJSON, log.IPAddress, log.UserAgent, log.Timestamp)
	return err
}
