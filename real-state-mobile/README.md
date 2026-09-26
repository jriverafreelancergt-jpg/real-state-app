# Realty Astarac Mobile App

Proyecto Flutter separado del backend `real-state-backend`.

## Recomendación de arquitectura

Se recomienda mantener este proyecto en un repositorio independiente del backend por estas razones:

- Ciclo de desarrollo diferente: mobile y backend no avanzan siempre en paralelo.
- Diferentes pipelines CI/CD, reviewers y despliegues.
- Separación clara de responsabilidades: API contract y DTOs entre equipos.
- Publicación independiente en App Store / Google Play.

## Estructura Hito 0

```text
lib/
  core/
    config/
    error/
    network/
    router/
    theme/
    utils/
  features/
    auth/
    property/
      data/
        dto/
        mappers/
        repositories/
      domain/
        entities/
        repositories/
      presentation/
        providers/
        screens/
        widgets/
```

## Dependencias iniciales

- flutter_riverpod
- go_router
- dio
- dartz
- flutter_secure_storage
- intl
- json_annotation
- freezed_annotation

## Hito 0 actual

- Arquitectura base definida
- Modelos de dominio y DTOs iniciales
- Mapper para propiedad
- Estructura feature-first con Clean Architecture

## Requisitos para ejecutar

1. Instalar Flutter SDK.
2. Ejecutar:

```bash
flutter pub get
flutter analyze
```

> Este proyecto está preparado como estructura base. La compilación real requiere Flutter instalado localmente o en CI.
