# ADR-001: Feature-First + Clean Architecture

- Estado: Aceptado
- Fecha: 2026-09-20

## Contexto

La aplicación móvil debe crecer con varias áreas funcionales (auth, properties, favorites, leads, appointments) y se requiere mantener separación clara entre negocio, datos y presentación.

## Decisión

Usaremos una estructura `feature-first` con capas `domain`, `data`, `presentation` dentro de cada feature.

## Consecuencias

- Ventajas:
  - Escalabilidad modular.
  - Mejor aislamiento por funcionalidad.
  - Facilita testing y maintenance.
- Costos:
  - Más carpetas y estructura inicial.
  - Requiere disciplina para respetar dependencias.
