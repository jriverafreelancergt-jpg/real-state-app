package dto

import (
	"errors"
)

// LoginRequestDTO representa las credenciales de login
type LoginRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate valida los campos del DTO
func (dto *LoginRequestDTO) Validate() error {
	if dto.Username == "" {
		return errors.New("username is required")
	}
	if dto.Password == "" {
		return errors.New("password is required")
	}
	// Aquí podrías agregar validaciones adicionales, como longitud mínima
	return nil
}

// LoginResponseDTO representa la respuesta de login con el token
type LoginResponseDTO struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	MFARequired  bool     `json:"mfa_required"`
	User         UserInfo `json:"user"`
}

// UserInfo representa información básica del usuario
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// TokenResponseDTO para refresh token
type TokenResponseDTO struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// MFAVerifyRequestDTO para verificación de MFA
type MFAVerifyRequestDTO struct {
	Code string `json:"code"`
}

// RefreshTokenRequestDTO para refresh token
type RefreshTokenRequestDTO struct {
	RefreshToken string `json:"refresh_token"`
}
