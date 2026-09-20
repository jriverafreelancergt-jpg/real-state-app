# Real State API – Guía de QA con Postman y Newman

Este documento describe cómo importar, ejecutar y validar la API REST de Real State usando Postman y Newman.

## Archivos de la colección

- `real-state.postman_collection.json`: colección principal con todos los endpoints
- `real-state.postman_environment.json`: entorno local listo para importar
- `real-state.happy_path.postman_collection.json`: suite secuencial de flujo principal

## Requisitos previos

- Postman instalado
- Node.js y npm instalados si se usa Newman
- La API ejecutándose en local

URL base esperada:

```text
http://localhost:8080
```

## Credenciales por defecto

- Usuario: `admin`
- Password: `admin123`

## Importación en Postman

1. Abre Postman.
2. Importa el archivo `real-state.postman_environment.json`.
3. Importa la colección principal `real-state.postman_collection.json`.
4. Opcional: importa también `real-state.happy_path.postman_collection.json`.
5. Selecciona el entorno `Real State API - Local`.
6. Verifica que `base_url` sea `http://localhost:8080`.

## Flujo recomendado de validación

Se recomienda ejecutar los endpoints en este orden:

1. `Auth > POST /v1/login`
2. `Auth > POST /v1/refresh`
3. `Config > GET /v1/config`
4. `Properties > POST /v1/properties`
5. `Properties > GET /v1/properties`
6. `Properties > GET /v1/properties/{id}`
7. `Sync & Monitoring > POST /v1/admin/sync`
8. `Sync & Monitoring > GET /v1/sync-status`
9. `Auth > POST /v1/logout`

## Auto-login

La colección incluye scripts de `pre-request` para:

- comprobar si hay un token activo
- hacer login automático si `bearer_token` está vacío
- inyectar `Authorization: Bearer <token>`
- enviar el header `X-Device-Fingerprint`

Esto elimina la necesidad de copiar manualmente el token entre cada request protegida.

## Happy Path

La suite `real-state.happy_path.postman_collection.json` ejecuta el flujo principal de forma secuencial con `postman.setNextRequest(...)`.

Ejecuta esta suite desde Postman Runner para validar lo siguiente:

- login exitoso
- refresh token
- consulta de configuración
- creación de propiedad
- listado de propiedades
- detalle por ID
- sincronización manual
- estado de sincronización
- logout

## Validaciones esperadas

- Login: `200 OK`
- Refresh: `200 OK`
- Config: `200 OK`
- Create property: `201 Created`
- List properties: `200 OK`
- Get property by ID: `200 OK`
- Trigger sync: `200 OK`
- Get sync status: `200 OK`
- Logout: `200 OK`

## Diagnóstico rápido

Si una petición falla:

- Asegúrate de que la API esté levantada en `http://localhost:8080`
- Verifica que hayas seleccionado el entorno correcto
- Comprueba que `base_url` no tenga una barra final extra
- Confirma que el servidor está respondiendo con JSON
- Revisa el header `X-Device-Fingerprint`

## Uso con Newman

Instala Newman:

```bash
npm install -g newman
```

Ejecuta la suite happy path:

```bash
newman run "real-state.happy_path.postman_collection.json" \
  -e "real-state.postman_environment.json" \
  --reporters cli,json \
  --reporter-json-export "newman-report.json"
```

Ejecuta la colección completa:

```bash
newman run "real-state.postman_collection.json" \
  -e "real-state.postman_environment.json" \
  --reporters cli,json \
  --reporter-json-export "newman-full-report.json"
```

## Integración continua

Puedes ejecutar esta validación desde GitHub Actions o GitLab CI usando Newman. Los ejemplos incluyen un workflow de GitHub Actions y un pipeline de GitLab.

## Recomendación final

Para QA y validación rápida, usa la suite `Happy Path` y la colección principal para comprobar rutas legacy y versionadas. La colección principal es útil para pruebas manuales; la suite happy path es ideal para ejecuciones automáticas y pipelines CI.
