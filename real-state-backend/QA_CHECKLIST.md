# QA Checklist – Real State API

## Preparación

- [ ] Importar `real-state.postman_environment.json`
- [ ] Importar `real-state.postman_collection.json`
- [ ] Importar `real-state.happy_path.postman_collection.json` (opcional)
- [ ] Seleccionar entorno `Real State API - Local`
- [ ] Confirmar `base_url = http://localhost:8080`
- [ ] Confirmar que la API está levantada y respondiendo

## Autenticación

- [ ] Ejecutar `Auth > POST /v1/login`
- [ ] Verificar que devuelve `200 OK`
- [ ] Verificar que incluye `access_token`
- [ ] Verificar que incluye `refresh_token`
- [ ] Confirmar que `bearer_token` quedó guardado en el entorno
- [ ] Ejecutar `Auth > POST /v1/refresh`
- [ ] Verificar que devuelve `200 OK`
- [ ] Confirmar que se actualiza el `access_token`
- [ ] Ejecutar `Auth > POST /v1/logout`
- [ ] Verificar que devuelve `200 OK`

## Configuración

- [ ] Ejecutar `Config > GET /v1/config`
- [ ] Verificar `200 OK`
- [ ] Verificar estructura del objeto de configuración
- [ ] Ejecutar `Config > PUT /v1/config`
- [ ] Verificar `200 OK`
- [ ] Verificar que la configuración actualizada se aplica correctamente

## Propiedades

- [ ] Ejecutar `Properties > POST /v1/properties`
- [ ] Verificar `201 Created`
- [ ] Verificar que `title`, `price` y `currency` son válidos
- [ ] Guardar `property_id` si la API devuelve un ID
- [ ] Ejecutar `Properties > GET /v1/properties`
- [ ] Verificar `200 OK`
- [ ] Verificar lista de propiedades
- [ ] Ejecutar `Properties > GET /v1/properties/{id}`
- [ ] Verificar `200 OK`
- [ ] Verificar que el ID consultado corresponde a la propiedad creada

## Sincronización

- [ ] Ejecutar `Sync & Monitoring > POST /v1/admin/sync`
- [ ] Verificar `200 OK`
- [ ] Ejecutar `Sync & Monitoring > GET /v1/sync-status`
- [ ] Verificar `200 OK`
- [ ] Verificar que el payload de estado de sincronización tenga sentido

## Validaciones negativas

- [ ] Probar login con credenciales inválidas
- [ ] Verificar `401 Unauthorized`
- [ ] Probar refresh con token inválido
- [ ] Verificar `401 Unauthorized`
- [ ] Probar creación de propiedad con payload inválido
- [ ] Verificar `400 Bad Request` o `422 Unprocessable Entity`
- [ ] Probar acceso sin token a un endpoint protegido
- [ ] Verificar `401 Unauthorized`

## Resultado final esperado

- [ ] Autenticación funciona correctamente
- [ ] Tokens se gestionan bien
- [ ] El flujo principal deja propiedades creadas y consultadas
- [ ] La sincronización responde correctamente
- [ ] Los permisos y validaciones de negocio se cumplen
- [ ] La suite happy path pasa sin errores

## Observaciones

- [ ] Registrar errores, tiempos de respuesta y payloads relevantes
- [ ] Confirmar si las rutas legacy siguen funcionando
- [ ] Confirmar si la versión `/v1` y la compatibilidad sin versión responden igual
