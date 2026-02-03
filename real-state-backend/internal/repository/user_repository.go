package repository

import (
	"context"
	"database/sql"
	"real-state-backend/internal/core/domain"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, mfa_secret, failed_attempts, locked_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	user.ID = uuid.New().String()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash, user.MFASecret, user.FailedAttempts, user.LockedUntil, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, mfa_secret, failed_attempts, locked_until, created_at, updated_at
		FROM users WHERE username = $1`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.MFASecret,
		&user.FailedAttempts, &user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, mfa_secret, failed_attempts, locked_until, created_at, updated_at
		FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.MFASecret,
		&user.FailedAttempts, &user.LockedUntil, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateFailedAttempts(ctx context.Context, id string, attempts int, lockedUntil *time.Time) error {
	query := `UPDATE users SET failed_attempts = $1, locked_until = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, attempts, lockedUntil, time.Now(), id)
	return err
}

func (r *UserRepository) UpdateMFASecret(ctx context.Context, id string, secret string) error {
	query := `UPDATE users SET mfa_secret = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, secret, time.Now(), id)
	return err
}

// GetPermissions retorna los permisos efectivos del usuario (vía roles)
func (r *UserRepository) GetPermissions(ctx context.Context, userID string) ([]domain.Permission, error) {
	query := `
	SELECT DISTINCT p.id, p.name, p.resource, p.action, p.created_at
	FROM permissions p
	JOIN role_permissions rp ON p.id = rp.permission_id
	JOIN roles ro ON rp.role_id = ro.id
	JOIN user_roles ur ON ro.id = ur.role_id
	WHERE ur.user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := make([]domain.Permission, 0)
	for rows.Next() {
		var p domain.Permission
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt
		perms = append(perms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return perms, nil
}
