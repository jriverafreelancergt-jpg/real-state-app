package ports

import (
	"context"
	"real-state-backend/internal/core/domain"
)

// WasiProvider define la interfaz para interactuar con el proveedor externo de inmuebles (Wasi).
type WasiProvider interface {
	// FetchProperties obtiene propiedades paginadas de Wasi.
	// Retorna la lista de propiedades del dominio, el número total de páginas disponibles, y un error si ocurre.
	FetchProperties(ctx context.Context, page int) ([]domain.Property, int, error)
}

// PropertySyncService define el caso de uso para sincronizar propiedades.
type PropertySyncService interface {
	// SyncAllProperties orquesta la sincronización concurrente con Wasi.
	SyncAllProperties(ctx context.Context) error
}

