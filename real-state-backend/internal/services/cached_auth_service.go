package services

import (
	"context"
	"log/slog"
	"time"

	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"real-state-backend/internal/dto"
	"real-state-backend/pkg/cache"
)

// CachedAuthService encapsula AuthService con cacheo de permisos
type CachedAuthService struct {
	baseService ports.AuthService
	cache       cache.Cache
	cacheTTL    time.Duration
}

// NewCachedAuthService crea un nuevo servicio de autenticación con cacheo
func NewCachedAuthService(baseService ports.AuthService, c cache.Cache, ttl time.Duration) ports.AuthService {
	return &CachedAuthService{
		baseService: baseService,
		cache:       c,
		cacheTTL:    ttl,
	}
}

// GetUserPermissions obtiene permisos del usuario, intentando cacheo primero
func (cas *CachedAuthService) GetUserPermissions(ctx context.Context, userID string) ([]domain.Permission, error) {
	// Intentar obtener del cache
	if cas.cache != nil {
		if permissions, err := cas.cache.(*cache.RedisCache).GetCachedPermissions(ctx, userID); err == nil {
			slog.Debug("Permissions loaded from cache", "user_id", userID)
			return permissions, nil
		}
	}

	// Si no está en cache, obtener de BD
	permissions, err := cas.baseService.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Almacenar en cache para futuras solicitudes
	if cas.cache != nil {
		if err := cas.cache.(*cache.RedisCache).CachePermissions(ctx, userID, permissions, cas.cacheTTL); err != nil {
			slog.Warn("Failed to cache permissions", "user_id", userID, "error", err)
		}
	}

	return permissions, nil
}

// Proxy todos los otros métodos al servicio base
func (cas *CachedAuthService) Login(ctx context.Context, req dto.LoginRequestDTO, deviceFingerprint string, locationData map[string]interface{}, userAgent string, deviceMetadata map[string]interface{}) (dto.LoginResponseDTO, error) {
	return cas.baseService.Login(ctx, req, deviceFingerprint, locationData, userAgent, deviceMetadata)
}

func (cas *CachedAuthService) VerifyMFA(ctx context.Context, userID string, code string) error {
	return cas.baseService.VerifyMFA(ctx, userID, code)
}

func (cas *CachedAuthService) RefreshToken(ctx context.Context, refreshToken string, deviceFingerprint string) (dto.TokenResponseDTO, error) {
	return cas.baseService.RefreshToken(ctx, refreshToken, deviceFingerprint)
}

func (cas *CachedAuthService) Logout(ctx context.Context, tokenJTI string) error {
	return cas.baseService.Logout(ctx, tokenJTI)
}

func (cas *CachedAuthService) ValidateSession(ctx context.Context, tokenJTI string) (*domain.UserSession, error) {
	return cas.baseService.ValidateSession(ctx, tokenJTI)
}

func (cas *CachedAuthService) LogAuditError(ctx context.Context, eventType string, err error, userID *string) error {
	return cas.baseService.LogAuditError(ctx, eventType, err, userID)
}
