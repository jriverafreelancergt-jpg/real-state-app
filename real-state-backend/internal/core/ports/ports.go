package ports

import (
	"context"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/dto"
	"time"
)

// PropertyRepository define las operaciones de base de datos.
type PropertyRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Property, error)
	GetAll(ctx context.Context, limit, offset int) ([]domain.Property, error)
	Create(ctx context.Context, property *domain.Property) error
	// Aquí agregarías métodos de filtro avanzados más adelante
}

// PropertyService define la lógica de negocio.
type PropertyService interface {
	GetProperty(ctx context.Context, id int64) (*domain.Property, error)
	ListProperties(ctx context.Context, page, pageSize int) ([]domain.Property, error)
	CreateProperty(ctx context.Context, property *domain.Property) error
}

// AuthService define la lógica de autenticación.
type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequestDTO, deviceFingerprint string, locationData map[string]interface{}, userAgent string, deviceMetadata map[string]interface{}) (dto.LoginResponseDTO, error)
	VerifyMFA(ctx context.Context, userID string, code string) error
	RefreshToken(ctx context.Context, refreshToken string, deviceFingerprint string) (dto.TokenResponseDTO, error)
	Logout(ctx context.Context, tokenJTI string) error
	ValidateSession(ctx context.Context, tokenJTI string) (*domain.UserSession, error)
	GetUserPermissions(ctx context.Context, userID string) ([]domain.Permission, error)
}

// UserRepository define operaciones de BD para usuarios.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateFailedAttempts(ctx context.Context, id string, attempts int, lockedUntil *time.Time) error
	UpdateMFASecret(ctx context.Context, id string, secret string) error
	// GetPermissions retorna los permisos efectivos de un usuario (a través de roles)
	GetPermissions(ctx context.Context, userID string) ([]domain.Permission, error)
}

// SessionRepository define operaciones de BD para sesiones.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.UserSession) error
	GetByTokenJTI(ctx context.Context, jti string) (*domain.UserSession, error)
	RevokeByUserID(ctx context.Context, userID string) error
	RevokeByJTI(ctx context.Context, jti string) error
	UpdateRefreshToken(ctx context.Context, jti string, newHash string, newExpiry time.Time) error
}

// AuditRepository define operaciones de BD para auditoría.
type AuditRepository interface {
	LogEvent(ctx context.Context, log *domain.AuditLog) error
}

// SecurityConfigRepository define operaciones de BD para configuración de seguridad.
type SecurityConfigRepository interface {
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetDuration(ctx context.Context, key string) (time.Duration, error)
	UpdateConfig(ctx context.Context, key string, value string) error
}
