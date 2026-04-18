package context

import (
	"context"
	"real-state-backend/pkg/clientid"
)

const (
	// ContextKeyClientIdentity es la clave para almacenar ClientIdentity en context
	ContextKeyClientIdentity = "client_identity"
)

// FromContext extrae ClientIdentity del contexto
// Retorna nil si no existe o no es válido
func FromContext(ctx context.Context) *clientid.ClientIdentity {
	if ctx == nil {
		return nil
	}

	value := ctx.Value(ContextKeyClientIdentity)
	if value == nil {
		return nil
	}

	ci, ok := value.(*clientid.ClientIdentity)
	if !ok {
		return nil
	}

	return ci
}

// WithClientIdentity agrega ClientIdentity al contexto
func WithClientIdentity(ctx context.Context, ci *clientid.ClientIdentity) context.Context {
	if ctx == nil {
		return nil
	}
	if ci == nil {
		return ctx
	}
	return context.WithValue(ctx, ContextKeyClientIdentity, ci)
}

// GetRequestID obtiene el request_id del contexto
func GetRequestID(ctx context.Context) string {
	ci := FromContext(ctx)
	if ci == nil {
		return ""
	}
	return ci.RequestID
}

// GetClientID obtiene el client_id del contexto
func GetClientID(ctx context.Context) string {
	ci := FromContext(ctx)
	if ci == nil {
		return ""
	}
	return ci.ClientID
}

// GetClientType obtiene el client_type del contexto
func GetClientType(ctx context.Context) string {
	ci := FromContext(ctx)
	if ci == nil {
		return ""
	}
	return ci.ClientType
}

// GetDeviceID obtiene el device_id del contexto
func GetDeviceID(ctx context.Context) string {
	ci := FromContext(ctx)
	if ci == nil {
		return ""
	}
	return ci.DeviceID
}

// GetUserID obtiene el user_id del contexto
// Intenta primero desde ClientIdentity, si no, intenta desde "user_id" directo
func GetUserID(ctx context.Context) string {
	if ci := FromContext(ctx); ci != nil && ci.UserID != "" {
		return ci.UserID
	}

	// Fallback a contexto directo (para compatibilidad con JWTMiddleware existente)
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}

	return ""
}

// GetOrigin obtiene la IP del cliente
func GetOrigin(ctx context.Context) string {
	ci := FromContext(ctx)
	if ci == nil {
		return ""
	}
	return ci.Origin
}

// GetClientIdentityOrEmpty retorna ClientIdentity o un struct vacío si no existe
func GetClientIdentityOrEmpty(ctx context.Context) *clientid.ClientIdentity {
	ci := FromContext(ctx)
	if ci == nil {
		return &clientid.ClientIdentity{}
	}
	return ci
}
