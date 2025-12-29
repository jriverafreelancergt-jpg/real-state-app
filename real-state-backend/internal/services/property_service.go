package services

import (
    "backend/internal/core/domain"
    "backend/internal/core/ports"
)

type propertyService struct {
    repo ports.PropertyRepository // El servicio depende de la interfaz, no de la DB real
}

func NewPropertyService(r ports.PropertyRepository) ports.PropertyService {
    return &propertyService{repo: r}
}

func (s *propertyService) CreateProperty(p domain.Property) error {
    // Aquí irían validaciones de negocio
    return s.repo.Save(&p)
}