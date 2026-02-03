-- Migration: 000002_iam_system.up.sql
-- Sistema IAM de Alta Seguridad: Usuarios, RBAC, Sesiones y Auditoría

-- Tabla de Usuarios
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    mfa_secret VARCHAR(32), -- Para TOTP (opcional)
    failed_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Permisos
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Relaciones Usuario-Rol
CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Tabla de Relaciones Rol-Permiso
CREATE TABLE role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- Tabla de Sesiones de Usuario
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_jti VARCHAR(36) UNIQUE NOT NULL, -- JWT ID
    refresh_token_hash VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL, -- Fingerprint del dispositivo
    location_data JSONB, -- {ip, country, city, lat, lng}
    user_agent TEXT,
    device_metadata JSONB, -- {os, app_version, etc.}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN DEFAULT FALSE
);

-- Tabla de Auditoría
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL, -- LOGIN_SUCCESS, LOGIN_FAILURE, etc.
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resource VARCHAR(100),
    action VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para rendimiento
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_jti ON user_sessions(token_jti);
CREATE INDEX idx_user_sessions_device_id ON user_sessions(device_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);

-- Datos iniciales: Usuario admin y roles básicos
INSERT INTO users (username, email, password_hash) VALUES ('admin', 'admin@example.com', '$2a$12$3NlU5lfvtVs4QD1w.pq2rOJj/6tBYYTtzWpRPvmAvWaYLHaVuBMae');
INSERT INTO roles (name, description) VALUES ('admin', 'Administrador del sistema');
INSERT INTO roles (name, description) VALUES ('user', 'Usuario estándar');
INSERT INTO permissions (name, resource, action) VALUES ('create_property', 'properties', 'create');
INSERT INTO permissions (name, resource, action) VALUES ('read_property', 'properties', 'read');
INSERT INTO permissions (name, resource, action) VALUES ('update_property', 'properties', 'update');
INSERT INTO permissions (name, resource, action) VALUES ('delete_property', 'properties', 'delete');
INSERT INTO user_roles (user_id, role_id) SELECT u.id, r.id FROM users u, roles r WHERE u.username = 'admin' AND r.name = 'admin';
INSERT INTO role_permissions (role_id, permission_id) SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin' AND p.resource = 'properties';