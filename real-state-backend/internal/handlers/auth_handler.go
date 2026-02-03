package handlers

import (
	"encoding/json"
	"net/http"
	"real-state-backend/internal/core/ports"
	"real-state-backend/internal/dto"
	"time"

	"github.com/google/uuid"
)

type AuthHandler struct {
	service ports.AuthService
}

func NewAuthHandler(s ports.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

// Login maneja el endpoint de login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDTO

	// Decodificar JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json", "auth", nil)
		return
	}

	// Validar DTO
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), "validation_error", "auth", nil)
		return
	}

	// Extraer metadatos del dispositivo (simplificado)
	deviceFingerprint := r.Header.Get("X-Device-Fingerprint") // Debería generarse en cliente
	if deviceFingerprint == "" {
		deviceFingerprint = "default-device"
	}
	locationData := map[string]interface{}{
		"ip": r.RemoteAddr,
	}
	userAgent := r.Header.Get("User-Agent")
	deviceMetadata := map[string]interface{}{
		"os": "unknown",
	}

	// Llamar al servicio
	resp, err := h.service.Login(r.Context(), req, deviceFingerprint, locationData, userAgent, deviceMetadata)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error(), "invalid_credentials", "auth", nil)
		return
	}

	// Construir respuesta con envoltorio (status, meta, request id)
	requestID := "req_auth_" + uuid.New().String()
	serverTime := time.Now().UTC().Format(time.RFC3339)

	wrapper := dto.LoginResponseWrapper{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Autenticación completada con éxito",
		Data: dto.LoginResponseData{
			User: resp.User,
			Authentication: dto.LoginResponseDTO{
				AccessToken:  resp.AccessToken,
				RefreshToken: resp.RefreshToken,
				TokenType:    resp.TokenType,
				ExpiresIn:    resp.ExpiresIn,
				MFARequired:  resp.MFARequired,
			},
		},
		Meta: dto.ResponseMeta{
			ServerTime: serverTime,
			RequestID:  requestID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wrapper)
}

// VerifyMFA maneja verificación de MFA
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string) // De middleware
	var req dto.MFAVerifyRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json", "auth", nil)
		return
	}

	err := h.service.VerifyMFA(r.Context(), userID, req.Code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error(), "mfa_invalid", "auth", nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "MFA verified"})
}

// RefreshToken maneja refresh de tokens
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json", "auth", nil)
		return
	}

	deviceFingerprint := r.Header.Get("X-Device-Fingerprint")
	if deviceFingerprint == "" {
		deviceFingerprint = "default-device"
	}

	resp, err := h.service.RefreshToken(r.Context(), req.RefreshToken, deviceFingerprint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout maneja logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	jti := r.Context().Value("jti").(string) // De middleware
	err := h.service.Logout(r.Context(), jti)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Logout failed", "logout_error", "auth", nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}
