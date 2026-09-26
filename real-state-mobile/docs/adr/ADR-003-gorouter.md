# ADR-003: GoRouter

- Estado: Aceptado
- Fecha: 2026-09-20

## Contexto

La app necesita navegación declarativa y soporte para deep links en futuro.

## Decisión

Usaremos `go_router` como router principal.

## Consecuencias

- Ventajas:
  - Navegación declarativa.
  - Deep link + web route friendly.
  - Integración simple con MaterialApp.router.
- Alternativas:
  - Navigator 1.0: adecuado para MVP muy simple, no ideal para crecimiento.
