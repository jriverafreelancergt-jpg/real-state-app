package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"real-state-backend/internal/core/domain"
	"time"

	"github.com/google/uuid"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, token_jti, refresh_token_hash, device_id, location_data, user_agent, device_metadata, created_at, expires_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	session.ID = uuid.New().String()
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}

	locationJSON, _ := json.Marshal(session.LocationData)
	deviceJSON, _ := json.Marshal(session.DeviceMetadata)

	_, err := r.db.ExecContext(ctx, query, session.ID, session.UserID, session.TokenJTI, session.RefreshTokenHash,
		session.DeviceID, locationJSON, session.UserAgent, deviceJSON, session.CreatedAt, session.ExpiresAt, session.Revoked)
	return err
}

func (r *SessionRepository) GetByTokenJTI(ctx context.Context, jti string) (*domain.UserSession, error) {
	query := `
		SELECT id, user_id, token_jti, refresh_token_hash, device_id, location_data, user_agent, device_metadata, created_at, expires_at, revoked
		FROM user_sessions WHERE token_jti = $1 AND revoked = FALSE`

	session := &domain.UserSession{}
	var locationJSON, deviceJSON []byte
	err := r.db.QueryRowContext(ctx, query, jti).Scan(
		&session.ID, &session.UserID, &session.TokenJTI, &session.RefreshTokenHash,
		&session.DeviceID, &locationJSON, &session.UserAgent, &deviceJSON,
		&session.CreatedAt, &session.ExpiresAt, &session.Revoked,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(locationJSON, &session.LocationData)
	json.Unmarshal(deviceJSON, &session.DeviceMetadata)
	return session, nil
}

func (r *SessionRepository) RevokeByUserID(ctx context.Context, userID string) error {
	query := `UPDATE user_sessions SET revoked = TRUE WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *SessionRepository) RevokeByJTI(ctx context.Context, jti string) error {
	query := `UPDATE user_sessions SET revoked = TRUE WHERE token_jti = $1`
	_, err := r.db.ExecContext(ctx, query, jti)
	return err
}

func (r *SessionRepository) UpdateRefreshToken(ctx context.Context, jti string, newHash string, newExpiry time.Time) error {
	query := `UPDATE user_sessions SET refresh_token_hash = $1, expires_at = $2 WHERE token_jti = $3`
	_, err := r.db.ExecContext(ctx, query, newHash, newExpiry, jti)
	return err
}
