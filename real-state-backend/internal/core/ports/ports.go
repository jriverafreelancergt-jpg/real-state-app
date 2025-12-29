package ports

import "backend/internal/core/domain"

// El repositorio es un "puerto de salida" (hacia la DB)
type PropertyRepository interface {
    Save(p *domain.Property) error
}

// El servicio es un "puerto de entrada" (hacia la lógica)
type PropertyService interface {
    CreateProperty(p domain.Property) error
}