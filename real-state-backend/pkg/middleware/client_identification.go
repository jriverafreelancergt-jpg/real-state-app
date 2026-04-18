package middleware

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"real-state-backend/pkg/clientid"
	contextpkg "real-state-backend/pkg/context"

	"github.com/google/uuid"
)

// RequestIDMiddleware genera o valida un request ID único para cada petición
// Debe ejecutarse ANTES de ClientIdentificationMiddleware
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Intentar obtener request_id del header
		requestID := r.Header.Get("X-Request-ID")

		// 2. Si no existe, generar uno
		if requestID == "" {
			requestID = "req_" + uuid.New().String()
		}

		// 3. Validar formato (debe empezar con "req_" o ser UUID válido)
		if !strings.HasPrefix(requestID, "req_") && !isValidUUID(requestID) {
			requestID = "req_" + uuid.New().String()
		}

		// 4. Establecer en header de respuesta para que el cliente lo reciba
		w.Header().Set("X-Request-ID", requestID)

		// 5. Agregar al contexto (para acceso posterior)
		ctx := r.Context()
		ctx = context.WithValue(ctx, contextValue("request_id"), requestID)
		slog.Debug("Request ID generated", "request_id", requestID)

		next.ServeHTTP(w, r)
	})
}

// ClientIdentificationMiddleware enriquece la petición con información de cliente/dispositivo
// Debe ejecutarse DESPUÉS de RequestIDMiddleware
func ClientIdentificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extraer request_id que fue generado por RequestIDMiddleware
		var requestID string
		if rid := r.Context().Value(contextValue("request_id")); rid != nil {
			requestID = rid.(string)
		} else {
			requestID = "req_" + uuid.New().String()
		}

		// Extraer información del cliente
		clientType := detectClientType(r)
		clientID := extractClientID(r)
		deviceID := extractDeviceID(r)
		deviceName := extractDeviceName(r)
		origin := extractOrigin(r)
		clientVersion := r.Header.Get("X-Client-Version")

		// Crear ClientIdentity
		clientIdentity := &clientid.ClientIdentity{
			RequestID:     requestID,
			ClientID:      clientID,
			ClientType:    clientType,
			DeviceID:      deviceID,
			DeviceName:    deviceName,
			Origin:        origin,
			ClientVersion: clientVersion,
			Timestamp:     time.Now().UTC(),
		}

		// Agregar al contexto
		ctx := contextpkg.WithClientIdentity(r.Context(), clientIdentity)
		r = r.WithContext(ctx)

		// Log de auditoría
		slog.Debug(
			"Client identified",
			"request_id", clientIdentity.RequestID,
			"client_type", clientIdentity.ClientType,
			"client_id", clientIdentity.ClientID,
			"device_id", clientIdentity.DeviceID,
			"origin", clientIdentity.Origin,
		)

		next.ServeHTTP(w, r)
	})
}

// detectClientType identifica el tipo de cliente basado en headers y User-Agent
func detectClientType(r *http.Request) string {
	// 1. Header explícito (PRIORIDAD 1)
	if clientType := r.Header.Get("X-Client-Type"); clientType != "" {
		if isValidClientType(clientType) {
			return clientType
		}
	}

	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return "unknown"
	}

	// 2. Headers y User-Agent de cliente (PRIORIDAD 2)
	if strings.Contains(ua, "RealState-Mobile") || strings.Contains(ua, "RealStateApp") {
		if strings.Contains(ua, "iOS") {
			return "mobile_ios"
		}
		if strings.Contains(ua, "Android") {
			return "mobile_android"
		}
		return "mobile_android" // Default si no especifica
	}

	if strings.Contains(ua, "RealState-Web") || strings.Contains(ua, "RealStateWeb") {
		return "web"
	}

	// 3. SDK/Integración patterns (PRIORIDAD 3)
	if strings.Contains(ua, "gRPC") || strings.Contains(ua, "grpc") {
		return "b2b"
	}
	if strings.Contains(ua, "curl") || strings.Contains(ua, "Postman") ||
		strings.Contains(ua, "insomnia") || strings.Contains(ua, "Thunder Client") {
		// Testing tools pero reportar como b2b
		return "b2b"
	}

	// 4. Browser/Web detection (PRIORIDAD 4)
	browserKeywords := []string{
		"Chrome", "Firefox", "Safari", "Edge", "Opera",
		"Mozilla", "AppleWebKit", "Gecko", "Trident",
	}
	for _, keyword := range browserKeywords {
		if strings.Contains(ua, keyword) {
			return "web"
		}
	}

	// 5. Mobile detection por User-Agent (PRIORIDAD 5)
	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPad", "iPod",
		"Windows Phone", "IEMobile", "BlackBerry",
	}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			if strings.Contains(ua, "iPad") {
				return "mobile_ios"
			}
			if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPod") {
				return "mobile_ios"
			}
			if strings.Contains(ua, "Android") {
				return "mobile_android"
			}
		}
	}

	// 6. Default
	return "unknown"
}

// extractClientID obtiene el ID del cliente
func extractClientID(r *http.Request) string {
	// 1. Header explícito X-Client-ID
	if clientID := r.Header.Get("X-Client-ID"); clientID != "" {
		return clientID
	}

	// 2. Device ID como fallback (será generado por extractDeviceID)
	if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
		return deviceID
	}

	// 3. IP + User-Agent hash como último recurso
	origin := extractOrigin(r)
	ua := r.Header.Get("User-Agent")
	if origin != "" && ua != "" {
		hash := md5.Sum([]byte(origin + ua))
		return "client_" + fmt.Sprintf("%x", hash)[:12]
	}

	return "unknown_client"
}

// extractDeviceID obtiene el ID único del dispositivo
func extractDeviceID(r *http.Request) string {
	// 1. Header explícito X-Device-ID (IMEI, UUID, fingerprint del cliente)
	if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
		return deviceID
	}

	// 2. Para web: usar fingerprint basado en User-Agent + IP
	ua := r.Header.Get("User-Agent")
	origin := extractOrigin(r)

	if ua != "" && origin != "" {
		hash := md5.Sum([]byte(ua + origin))
		return fmt.Sprintf("%x", hash)[:16]
	}

	// 3. Fallback: IP + timestamp hash
	if origin != "" {
		hash := md5.Sum([]byte(origin))
		return fmt.Sprintf("%x", hash)[:16]
	}

	return "unknown_device"
}

// extractDeviceName obtiene el nombre/descripción del dispositivo
func extractDeviceName(r *http.Request) string {
	// 1. Header explícito
	if deviceName := r.Header.Get("X-Device-Name"); deviceName != "" {
		return deviceName
	}

	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return "unknown"
	}

	// 2. Extractar info de User-Agent
	// iPhone
	if strings.Contains(ua, "iPhone") {
		if strings.Contains(ua, "iPhone15") {
			return "iPhone 15"
		}
		if strings.Contains(ua, "iPhone14") {
			return "iPhone 14"
		}
		if strings.Contains(ua, "iPhone13") {
			return "iPhone 13"
		}
		return "iPhone"
	}

	// iPad
	if strings.Contains(ua, "iPad") {
		if strings.Contains(ua, "iPad Pro") {
			return "iPad Pro"
		}
		return "iPad"
	}

	// Android
	if strings.Contains(ua, "Android") {
		// Intentar extraer modelo
		parts := strings.Split(ua, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "SM-") || strings.Contains(part, "GT-") {
				// Samsung
				if idx := strings.Index(part, " Build"); idx > 0 {
					return part[:idx]
				}
				return part
			}
		}
		return "Android Device"
	}

	// Browsers
	if strings.Contains(ua, "Chrome") {
		if strings.Contains(ua, "Mobile") {
			return "Chrome Mobile"
		}
		return "Chrome"
	}
	if strings.Contains(ua, "Safari") {
		if strings.Contains(ua, "Mobile") {
			return "Mobile Safari"
		}
		return "Safari"
	}
	if strings.Contains(ua, "Firefox") {
		return "Firefox"
	}
	if strings.Contains(ua, "Edge") {
		return "Microsoft Edge"
	}

	// Platform info
	if strings.Contains(ua, "Windows") {
		return "Windows PC"
	}
	if strings.Contains(ua, "Macintosh") {
		return "Mac"
	}
	if strings.Contains(ua, "Linux") {
		return "Linux"
	}

	return "unknown"
}

// extractOrigin obtiene la IP del cliente
func extractOrigin(r *http.Request) string {
	// 1. X-Forwarded-For (si hay proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Tomar la primera IP (cliente original)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. X-Real-IP (alternativa común)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 3. RemoteAddr directo
	if r.RemoteAddr != "" {
		// Eliminar puerto
		if idx := strings.LastIndex(r.RemoteAddr, ":"); idx > 0 {
			return r.RemoteAddr[:idx]
		}
		return r.RemoteAddr
	}

	return "unknown"
}

// isValidClientType verifica si el client_type es válido
func isValidClientType(clientType string) bool {
	validTypes := map[string]bool{
		"mobile_ios":     true,
		"mobile_android": true,
		"web":            true,
		"b2b":            true,
		"unknown":        true,
	}
	return validTypes[clientType]
}

// isValidUUID verifica si es un UUID válido
func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// contextValue es una clave privada para evitar colisiones
type contextValue string

// WithValue agrega un valor al contexto (helper local)
func WithValue(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, contextValue(key), value)
}
