package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"real-state-backend/internal/core/ports"

	"github.com/golang-jwt/jwt/v5"
)

// SecurityConfig contiene configuraciones de seguridad
type SecurityConfig struct {
	AllowedOrigins []string
}

// SecurityMiddleware agrupa los middlewares de seguridad
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. CORS (Cross-Origin Resource Sharing)
		// Ajusta "*" al dominio específico de tu app en producción
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 2. Security Headers (Protección básica contra ataques comunes)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// 3. Aquí iría la validación de Rate Limiting (usando librería externa como tollbooth)

		// Log de la petición (Auditoría básica)
		slog.Info("Request received", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}

// JWTMiddleware valida el token JWT con JTI y consistencia de dispositivo
func JWTMiddleware(authService ports.AuthService, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			// Verificar formato "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "Invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parsear y validar token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verificar método de firma
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Extraer claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error": "Invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			jti, ok := claims["jti"].(string)
			if !ok {
				http.Error(w, `{"error": "Missing JTI"}`, http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, `{"error": "Missing subject"}`, http.StatusUnauthorized)
				return
			}

			// Validar sesión
			session, err := authService.ValidateSession(r.Context(), jti)
			if err != nil {
				http.Error(w, `{"error": "Session invalid"}`, http.StatusUnauthorized)
				return
			}

			// Consistencia de dispositivo
			deviceFingerprint := r.Header.Get("X-Device-Fingerprint")
			if deviceFingerprint == "" {
				deviceFingerprint = "default-device"
			}
			if session.DeviceID != deviceFingerprint {
				authService.Logout(r.Context(), jti) // Revocar sesión
				http.Error(w, `{"error": "Device mismatch"}`, http.StatusUnauthorized)
				return
			}

			// Detección de viaje imposible (simplificada)
			currentIP := r.RemoteAddr
			if session.LocationData != nil {
				if prevIP, ok := session.LocationData["ip"].(string); ok && prevIP != currentIP {
					slog.Warn("Possible impossible travel", "user_id", userID, "prev_ip", prevIP, "current_ip", currentIP)
					// Podrías agregar lógica adicional aquí
				}
			}

			// Agregar al contexto
			slog.Info("JWT validated", "user_id", userID, "jti", jti)
			ctx := context.WithValue(r.Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "jti", jti)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// RBACMiddleware verifica permisos específicos
func RBACMiddleware(authService ports.AuthService, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDValue := r.Context().Value("user_id")
			if userIDValue == nil {
				http.Error(w, `{"error": "User not authenticated"}`, http.StatusUnauthorized)
				return
			}
			userID, ok := userIDValue.(string)
			if !ok {
				http.Error(w, `{"error": "Invalid user ID"}`, http.StatusUnauthorized)
				return
			}

			permissions, err := authService.GetUserPermissions(r.Context(), userID)
			if err != nil {
				http.Error(w, `{"error": "Failed to get permissions"}`, http.StatusInternalServerError)
				return
			}

			hasPermission := false
			for _, perm := range permissions {
				if perm.Name == requiredPermission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				http.Error(w, `{"error": "Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
