package services

import (
	"context"
	"fmt"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
)

type propertyService struct {
	repo ports.PropertyRepository
}

func NewPropertyService(repo ports.PropertyRepository) ports.PropertyService {
	return &propertyService{
		repo: repo,
	}
}

// CORRECCIÓN AQUÍ: Agregamos ctx y el puntero *
func (s *propertyService) CreateProperty(ctx context.Context, p *domain.Property) error {
	// Lógica de negocio (ej: validar que el título no esté vacío)
	if p.Title == "" {
		return fmt.Errorf("el título es obligatorio")
	}

	return s.repo.Create(ctx, p)
}

// Asegúrate de que los otros métodos también tengan el contexto:
func (s *propertyService) GetProperty(ctx context.Context, id int64) (*domain.Property, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *propertyService) ListProperties(ctx context.Context, page, pageSize int) ([]domain.Property, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset)
}
