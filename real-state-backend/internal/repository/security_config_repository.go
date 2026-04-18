package repository

import (
	"context"
	"database/sql"
	"real-state-backend/pkg/database"
	"strconv"
	"time"
)

type SecurityConfigRepository struct {
	db *sql.DB
}

func NewSecurityConfigRepository(db *sql.DB) *SecurityConfigRepository {
	return &SecurityConfigRepository{db: db}
}

// GetString obtiene una configuración como string
func (r *SecurityConfigRepository) GetString(ctx context.Context, key string) (string, error) {
	query := `SELECT value FROM security_config WHERE key = $1`
	var value string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		return "", database.HandleError(ctx, err, "GetString", "security_config", map[string]interface{}{"key": key})
	}
	return value, nil
}

// GetInt obtiene una configuración como entero
func (r *SecurityConfigRepository) GetInt(ctx context.Context, key string) (int, error) {
	query := `SELECT value FROM security_config WHERE key = $1`
	var value string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		return 0, database.HandleError(ctx, err, "GetInt", "security_config", map[string]interface{}{"key": key})
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, database.HandleError(ctx, err, "GetInt (conversion)", "security_config", map[string]interface{}{"key": key, "value": value})
	}
	return intVal, nil
}

// GetDuration obtiene una configuración como duración (en minutos por defecto)
func (r *SecurityConfigRepository) GetDuration(ctx context.Context, key string) (time.Duration, error) {
	query := `SELECT value FROM security_config WHERE key = $1`
	var value string
	err := r.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		return 0, database.HandleError(ctx, err, "GetDuration", "security_config", map[string]interface{}{"key": key})
	}
	minutes, err := strconv.Atoi(value)
	if err != nil {
		return 0, database.HandleError(ctx, err, "GetDuration (conversion)", "security_config", map[string]interface{}{"key": key, "value": value})
	}
	return time.Duration(minutes) * time.Minute, nil
}

// UpdateConfig actualiza una configuración
func (r *SecurityConfigRepository) UpdateConfig(ctx context.Context, key string, value string) error {
	query := `UPDATE security_config SET value = $1, updated_at = $2 WHERE key = $3`
	_, err := r.db.ExecContext(ctx, query, value, time.Now(), key)
	if err != nil {
		return database.HandleError(ctx, err, "UpdateConfig", "security_config", map[string]interface{}{"key": key, "value": value})
	}
	return nil
}
