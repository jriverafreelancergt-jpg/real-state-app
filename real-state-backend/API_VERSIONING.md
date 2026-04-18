# 🔄 Migración de Rutas - /v1/ Versionado

## 📝 Cambio Realizado

Se agregó versionado de API con prefijo `/v1/` en todas las rutas. Sin embargo, se mantiene **backward compatibility** con rutas sin versión.

## ✅ Ambas Funcionan

### Rutas con Versión (Recomendado para nuevos clientes)
```bash
POST   /v1/login
POST   /v1/refresh
POST   /v1/verify-mfa
POST   /v1/logout
GET    /v1/properties
POST   /v1/properties
GET    /v1/properties/{id}
GET    /v1/config
PUT    /v1/config
```

### Rutas sin Versión (Backward Compatibility)
```bash
POST   /login
POST   /refresh
POST   /verify-mfa
POST   /logout
GET    /properties
POST   /properties
GET    /properties/{id}
GET    /config
PUT    /config
```

## 🎯 Recomendación

**Para nuevos desarrollos**: Usa `/v1/` (URLs oficiales)
**Para código existente**: Continúa con rutas sin versión (seguirán funcionando)

## 📊 Ejemplo de Uso

### Con versión (preferido)
```bash
curl -X POST http://localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'
```

### Sin versión (legacy, aún funciona)
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'
```

Ambas retornan la misma respuesta.

## 🚀 Plan Futuro

- **v1.x**: Mantenemos ambas rutas activas
- **v2.0**: Solo `/v2/` estará disponible (cuando haya cambios API)
- **Deprecation**: Se avisará con 6 meses de anticipación

---

**TL;DR**: Las rutas sin `/v1/` todavía funcionan. No hay cambios requeridos en clientes existentes, pero se recomienda migrar a `/v1/` gradualmente.
