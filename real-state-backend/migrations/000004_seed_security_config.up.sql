-- Insertar keys si no existen (no tocamos timestamps que no existen en la tabla)
INSERT INTO security_config (key, value) VALUES
  ('ACCESS_TOKEN_TTL_MINUTES','15'),
  ('REFRESH_TOKEN_TTL_DAYS','7'),
  ('MAX_FAILED_ATTEMPTS','5'),
  ('LOCKOUT_DURATION_MINUTES','15')
ON CONFLICT (key) DO NOTHING;
