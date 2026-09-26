# ADR-002: Riverpod

- Estado: Aceptado
- Fecha: 2026-09-20

## Contexto

Necesitamos gestión de estado y dependencia inyectada con un enfoque moderno y mantenible para Flutter.

## Decisión

Usaremos Riverpod como solución principal de state management.

## Consecuencias

- Ventajas:
  - Inyección declarativa.
  - Testabilidad buena.
  - Integración limpia con providers y auto-dispose.
- Alternativas consideradas:
  - BLoC: más verboso y con más boilerplate.
  - Provider clásico: menos robusto para la escala proyectada.
