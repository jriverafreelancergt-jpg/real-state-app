package clientid

import (
	"encoding/json"
	"fmt"
	"time"
)

// ClientIdentity contiene información unívoca del cliente que realizó la petición
type ClientIdentity struct {
	RequestID     string    `json:"request_id"`     // UUID único para la petición (e.g., "req_abc123")
	ClientID      string    `json:"client_id"`      // ID del cliente (user_id, app_id, integrator_id)
	ClientType    string    `json:"client_type"`    // mobile_ios, mobile_android, web, b2b, unknown
	DeviceID      string    `json:"device_id"`      // ID único del dispositivo (IMEI, fingerprint, etc)
	DeviceName    string    `json:"device_name"`    // iPhone 14, Samsung S24, Chrome on Windows, etc
	Origin        string    `json:"origin"`         // IP del cliente (192.168.1.100)
	UserID        string    `json:"user_id"`        // User ID del JWT (si autenticado)
	SessionID     string    `json:"session_id"`     // JTI del token (si autenticado)
	ClientVersion string    `json:"client_version"` // Versión: 2.1.3, web-build-456, etc
	Timestamp     time.Time `json:"timestamp"`      // Hora de inicio de la petición
}

// MarshalJSON personaliza la serialización para que solo incluya campos no-vacíos
func (ci *ClientIdentity) MarshalJSON() ([]byte, error) {
	type Alias ClientIdentity
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: ci.Timestamp.UTC().Format(time.RFC3339),
		Alias:     (*Alias)(ci),
	})
}

// String proporciona una representación legible
func (ci *ClientIdentity) String() string {
	if ci == nil {
		return "[UNKNOWN CLIENT]"
	}
	return fmt.Sprintf(
		"[req_id=%s | client_id=%s | type=%s | device=%s | origin=%s]",
		ci.RequestID, ci.ClientID, ci.ClientType, ci.DeviceID, ci.Origin,
	)
}

// LogFields retorna un mapa de campos para incluir en logs slog
// Solo incluye campos no-vacíos para no contaminar los logs
func (ci *ClientIdentity) LogFields() map[string]interface{} {
	if ci == nil {
		return map[string]interface{}{}
	}

	fields := make(map[string]interface{})

	if ci.RequestID != "" {
		fields["request_id"] = ci.RequestID
	}
	if ci.ClientID != "" {
		fields["client_id"] = ci.ClientID
	}
	if ci.ClientType != "" {
		fields["client_type"] = ci.ClientType
	}
	if ci.DeviceID != "" {
		fields["device_id"] = ci.DeviceID
	}
	if ci.DeviceName != "" {
		fields["device_name"] = ci.DeviceName
	}
	if ci.Origin != "" {
		fields["origin"] = ci.Origin
	}
	if ci.UserID != "" {
		fields["user_id"] = ci.UserID
	}
	if ci.SessionID != "" {
		fields["session_id"] = ci.SessionID
	}
	if ci.ClientVersion != "" {
		fields["client_version"] = ci.ClientVersion
	}

	return fields
}

// IsValid verifica que al menos request_id y client_type estén presentes
func (ci *ClientIdentity) IsValid() bool {
	if ci == nil {
		return false
	}
	return ci.RequestID != "" && ci.ClientType != ""
}

// IsAuthenticated verifica si el cliente está autenticado (tiene UserID)
func (ci *ClientIdentity) IsAuthenticated() bool {
	if ci == nil {
		return false
	}
	return ci.UserID != ""
}

// IsMobile verifica si es un cliente móvil
func (ci *ClientIdentity) IsMobile() bool {
	if ci == nil {
		return false
	}
	return ci.ClientType == "mobile_ios" || ci.ClientType == "mobile_android"
}

// IsWeb verifica si es un cliente web
func (ci *ClientIdentity) IsWeb() bool {
	if ci == nil {
		return false
	}
	return ci.ClientType == "web"
}

// IsB2B verifica si es un cliente B2B
func (ci *ClientIdentity) IsB2B() bool {
	if ci == nil {
		return false
	}
	return ci.ClientType == "b2b"
}
