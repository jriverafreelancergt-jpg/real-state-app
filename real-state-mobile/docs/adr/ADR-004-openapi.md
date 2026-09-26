# ADR-004: OpenAPI como contrato

- Estado: Aceptado
- Fecha: 2026-09-20

## Contexto

La API y el mobile deben evolucionar de manera sincronizada sin acoplarse a cambios improvisados.

## Decisión

Documentaremos la API con OpenAPI y usaremos ese contract como fuente de verdad para DTOs y endpoints.

## Consecuencias

- Ventajas:
  - Más clara colaboración backend-mobile.
  - Menos drift entre contrato y código.
  - Facilita SDK generation futuro.
- Costos:
  - Requiere disciplina al mantener el spec actualizado.
