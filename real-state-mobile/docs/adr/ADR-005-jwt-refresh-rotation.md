# ADR-005: JWT + Refresh Token Rotation

- Estado: Aceptado
- Fecha: 2026-09-20

## Contexto

La aplicación requiere sesiones seguras, expiración controlada y posibilidad de revocar tokens sin romper la UX.

## Decisión

Usaremos JWT para Access Token y Refresh Token con rotación y revocación controlada en backend. El mobile solo almacenará tokens en `flutter_secure_storage` y no los escribirá en almacenamiento inseguro.

## Consecuencias

- Ventajas:
  - Mejor seguridad.
  - Menor riesgo de abuso de tokens reutilizados.
  - Soporte para logout seguro y expiración.
- Costos:
  - Requiere flujo de refresh y manejo de race conditions.
  - El mobile debe manejar expiración 401/403 con reintento inteligente.
