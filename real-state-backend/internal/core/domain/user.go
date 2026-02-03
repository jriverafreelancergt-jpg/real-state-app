package domain

import (
	"time"
)

// User representa un usuario del sistema
type User struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	MFASecret      *string    `json:"-"`
	FailedAttempts int        `json:"-"`
	LockedUntil    *time.Time `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Role representa un rol
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Permission representa un permiso
type Permission struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Resource  string    `json:"resource"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

// UserSession representa una sesión de usuario
type UserSession struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id"`
	TokenJTI         string                 `json:"-"`
	RefreshTokenHash string                 `json:"-"`
	DeviceID         string                 `json:"device_id"`
	LocationData     map[string]interface{} `json:"location_data"`
	UserAgent        string                 `json:"user_agent"`
	DeviceMetadata   map[string]interface{} `json:"device_metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	Revoked          bool                   `json:"revoked"`
}

// AuditLog representa un log de auditoría
type AuditLog struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"event_type"`
	UserID    *string                `json:"user_id"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	OldValues map[string]interface{} `json:"old_values"`
	NewValues map[string]interface{} `json:"new_values"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Timestamp time.Time              `json:"timestamp"`
}
