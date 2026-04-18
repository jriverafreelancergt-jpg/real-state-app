package services

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"real-state-backend/internal/dto"
	"real-state-backend/pkg/audit"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService maneja la lógica de autenticación avanzada
type AuthService struct {
	userRepo    ports.UserRepository
	sessionRepo ports.SessionRepository
	auditRepo   ports.AuditRepository
	jwtSecret   []byte
	jwtPepper   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	maxAttempts int
	lockoutDur  time.Duration
}

// NewAuthService crea una nueva instancia de AuthService
func NewAuthService(userRepo ports.UserRepository, sessionRepo ports.SessionRepository, auditRepo ports.AuditRepository, secret, pepper string, accessTTL, refreshTTL time.Duration, maxAttempts int, lockoutDur time.Duration) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		jwtSecret:   []byte(secret),
		jwtPepper:   []byte(pepper),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		maxAttempts: maxAttempts,
		lockoutDur:  lockoutDur,
	}
}

// Login maneja el proceso de login con MFA opcional
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequestDTO, deviceFingerprint string, locationData map[string]interface{}, userAgent string, deviceMetadata map[string]interface{}) (dto.LoginResponseDTO, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Usuario no encontrado - es un error de seguridad auditable
		auditErr := audit.ErrInvalidCredentials()
		s.LogAuditError(ctx, "LOGIN_FAILURE", auditErr, nil)
		return dto.LoginResponseDTO{}, auditErr
	}

	// Debug log temporal (desactivar en producción)
	slog.Info("Login attempt", "username", req.Username, "user_id", user.ID)

	// Verificar bloqueo
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		auditErr := audit.ErrAccountLocked()
		s.LogAuditError(ctx, "LOGIN_FAILURE", auditErr, &user.ID)
		return dto.LoginResponseDTO{}, auditErr
	}

	// Verificar contraseña
	pwOk := s.checkPassword(req.Password, user.PasswordHash)
	slog.Info("Password check result", "username", req.Username, "ok", pwOk) // temporal
	if !pwOk {
		s.handleFailedAttempt(ctx, user, userAgent)
		auditErr := audit.ErrInvalidCredentials()
		return dto.LoginResponseDTO{}, auditErr
	}

	// Reset failed attempts on success
	s.userRepo.UpdateFailedAttempts(ctx, user.ID, 0, nil)

	// Generar tokens
	accessToken, refreshToken, jti, err := s.generateTokens(user.ID, deviceFingerprint)
	if err != nil {
		return dto.LoginResponseDTO{}, err
	}

	// Crear sesión
	session := &domain.UserSession{
		UserID:           user.ID,
		TokenJTI:         jti,
		RefreshTokenHash: s.hashRefreshToken(refreshToken),
		DeviceID:         deviceFingerprint,
		LocationData:     locationData,
		UserAgent:        userAgent,
		DeviceMetadata:   deviceMetadata,
		ExpiresAt:        time.Now().Add(s.refreshTTL),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return dto.LoginResponseDTO{}, err
	}

	// Verificar si MFA está habilitado
	mfaRequired := user.MFASecret != nil && *user.MFASecret != ""

	s.logAudit(ctx, "LOGIN_SUCCESS", &user.ID, "auth", "login", nil, map[string]interface{}{"mfa_required": mfaRequired}, "", userAgent)

	return dto.LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.accessTTL.Seconds()),
		MFARequired:  mfaRequired,
		User: dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

// VerifyMFA verifica el código TOTP
func (s *AuthService) VerifyMFA(ctx context.Context, userID string, code string) error {
	// Implementar verificación TOTP (usar librería como github.com/pquerna/otp)
	// Por simplicidad, asumir verificación exitosa si code == "123456"
	if code != "123456" {
		s.logAudit(ctx, "MFA_FAILURE", &userID, "auth", "mfa_verify", nil, nil, "", "")
		return errors.New("invalid MFA code")
	}
	s.logAudit(ctx, "MFA_SUCCESS", &userID, "auth", "mfa_verify", nil, nil, "", "")
	return nil
}

// RefreshToken rota el refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, deviceFingerprint string) (dto.TokenResponseDTO, error) {
	// Parsear refresh token para obtener JTI
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		auditErr := audit.ErrInvalidToken("invalid refresh token signature")
		return dto.TokenResponseDTO{}, auditErr
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		auditErr := audit.ErrInvalidToken("invalid token claims")
		return dto.TokenResponseDTO{}, auditErr
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		auditErr := audit.ErrInvalidJTI()
		return dto.TokenResponseDTO{}, auditErr
	}

	session, err := s.sessionRepo.GetByTokenJTI(ctx, jti)
	if err != nil || session.Revoked {
		auditErr := audit.ErrSessionRevoked()
		return dto.TokenResponseDTO{}, auditErr
	}

	// Verificar device consistency
	if subtle.ConstantTimeCompare([]byte(session.DeviceID), []byte(deviceFingerprint)) != 1 {
		s.sessionRepo.RevokeByJTI(ctx, jti)
		auditErr := audit.ErrDeviceMismatch()
		s.LogAuditError(ctx, "SESSION_HIJACK_ATTEMPT", auditErr, &session.UserID)
		return dto.TokenResponseDTO{}, auditErr
	}

	// Generar nuevo access token
	newAccessToken, err := s.generateAccessToken(session.UserID, jti)
	if err != nil {
		return dto.TokenResponseDTO{}, err
	}

	s.logAudit(ctx, "TOKEN_REFRESH", &session.UserID, "auth", "refresh", nil, nil, "", "")

	return dto.TokenResponseDTO{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.accessTTL.Seconds()),
	}, nil
}

// Logout revoca la sesión
func (s *AuthService) Logout(ctx context.Context, tokenJTI string) error {
	err := s.sessionRepo.RevokeByJTI(ctx, tokenJTI)
	if err != nil {
		return err
	}
	s.logAudit(ctx, "LOGOUT", nil, "auth", "logout", nil, nil, "", "")
	return nil
}

// ValidateSession valida la sesión para middlewares
func (s *AuthService) ValidateSession(ctx context.Context, tokenJTI string) (*domain.UserSession, error) {
	session, err := s.sessionRepo.GetByTokenJTI(ctx, tokenJTI)
	if err != nil {
		return nil, err
	}
	if session.Revoked || time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session invalid")
	}
	return session, nil
}

// GetUserPermissions obtiene permisos del usuario
func (s *AuthService) GetUserPermissions(ctx context.Context, userID string) ([]domain.Permission, error) {
	// Obtener permisos desde la BD a través de UserRepository
	perms, err := s.userRepo.GetPermissions(ctx, userID)
	if err != nil {
		slog.Warn("Failed to load user permissions from DB", "user_id", userID, "error", err)
		return nil, err
	}
	return perms, nil
}

// generateTokens genera access y refresh tokens
func (s *AuthService) generateTokens(userID, deviceFingerprint string) (string, string, string, error) {
	jti := uuid.New().String()

	accessToken, err := s.generateAccessToken(userID, jti)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := s.generateRefreshToken(userID, jti)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, jti, nil
}

// generateAccessToken genera JWT de acceso
func (s *AuthService) generateAccessToken(userID, jti string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"jti":  jti,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.accessTTL).Unix(),
		"type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateRefreshToken genera JWT de refresh
func (s *AuthService) generateRefreshToken(userID, jti string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"jti":  jti,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.refreshTTL).Unix(),
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// checkPassword verifica contraseña con bcrypt + pepper
func (s *AuthService) checkPassword(password, hash string) bool {
	peppered := s.pepperPassword(password)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(peppered))
	return err == nil
}

// pepperPassword aplica pepper a la contraseña
func (s *AuthService) pepperPassword(password string) string {
	return fmt.Sprintf("%s%s", password, s.jwtPepper)
}

// hashRefreshToken hashea el refresh token para BD
func (s *AuthService) hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// handleFailedAttempt maneja intentos fallidos
func (s *AuthService) handleFailedAttempt(ctx context.Context, user *domain.User, userAgent string) {
	attempts := user.FailedAttempts + 1
	var lockedUntil *time.Time
	if attempts >= s.maxAttempts {
		lockTime := time.Now().Add(s.lockoutDur)
		lockedUntil = &lockTime
	}
	s.userRepo.UpdateFailedAttempts(ctx, user.ID, attempts, lockedUntil)
	s.logAudit(ctx, "LOGIN_FAILURE", &user.ID, "auth", "login", nil, map[string]interface{}{"attempts": attempts}, "", userAgent)
}

// logAudit registra evento de auditoría
func (s *AuthService) logAudit(ctx context.Context, eventType string, userID *string, resource, action string, oldValues, newValues map[string]interface{}, ip, userAgent string) {
	log := &domain.AuditLog{
		EventType: eventType,
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		OldValues: oldValues,
		NewValues: newValues,
		IPAddress: ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}
	s.auditRepo.LogEvent(ctx, log)
}

// LogAuditError registra un error auditable en la tabla de auditoría
// Usado para registrar errores de seguridad significativos
func (s *AuthService) LogAuditError(ctx context.Context, eventType string, err error, userID *string) error {
	auditErr, ok := err.(*audit.AuditableError)
	if !ok {
		// Si no es un AuditableError, loguear y retornar
		slog.Warn("LogAuditError: error no es AuditableError",
			"event_type", eventType,
			"error_type", fmt.Sprintf("%T", err),
		)
		return nil
	}

	// Construir newValues con la información del error
	errorData := map[string]interface{}{
		"error_code":    string(auditErr.Code),
		"error_message": auditErr.Message,
		"status_code":   auditErr.StatusCode,
	}

	// Agregar detalles adicionales si existen
	if auditErr.Details != nil {
		for k, v := range auditErr.Details {
			errorData[k] = v
		}
	}

	log := &domain.AuditLog{
		EventType: eventType,
		UserID:    userID,
		Resource:  "auth",
		Action:    "error_occurred",
		NewValues: errorData,
		Timestamp: time.Now(),
	}

	// Registrar el intento de guardar en BD
	slog.Debug("LogAuditError: intentando guardar en BD",
		"event_type", eventType,
		"error_code", string(auditErr.Code),
		"user_id", userID,
	)

	errorRepo := s.auditRepo.LogEvent(ctx, log)
	if errorRepo != nil {
		slog.Error("LogAuditError: fallo al guardar en BD",
			"event_type", eventType,
			"error_code", string(auditErr.Code),
			"error", errorRepo.Error(),
		)
		return errorRepo
	}

	slog.Info("LogAuditError: guardado exitosamente en BD",
		"event_type", eventType,
		"error_code", string(auditErr.Code),
	)
	return nil
}
