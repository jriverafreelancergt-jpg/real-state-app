-- Migration: 000003_security_config.up.sql
-- Tabla para configuración de seguridad dinámica

CREATE TABLE security_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL,
    value VARCHAR(255) NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insertar configuraciones por defecto
INSERT INTO security_config (key, value, description) VALUES
    ('ACCESS_TOKEN_TTL_MINUTES', '15', 'Tiempo de expiración del access token en minutos'),
    ('REFRESH_TOKEN_TTL_DAYS', '7', 'Tiempo de expiración del refresh token en días'),
    ('MAX_FAILED_ATTEMPTS', '5', 'Máximo número de intentos de login fallidos antes de bloqueo'),
    ('LOCKOUT_DURATION_MINUTES', '15', 'Duración del bloqueo de cuenta en minutos');

CREATE INDEX idx_security_config_key ON security_config(key);