package services

import (
	"context"
	"fmt"
	"net/url"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"strings"

	"github.com/google/uuid"
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

func (s *propertyService) UpdateProperty(ctx context.Context, id int64, p *domain.Property) (*domain.Property, error) {
	if p == nil {
		return nil, fmt.Errorf("la propiedad es obligatoria")
	}
	if p.Title == "" {
		return nil, fmt.Errorf("el título es obligatorio")
	}
	if p.Price <= 0 {
		return nil, fmt.Errorf("el precio debe ser mayor a cero")
	}

	return s.repo.Update(ctx, id, p)
}

func (s *propertyService) DeleteProperty(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("id de propiedad inválido")
	}

	return s.repo.Delete(ctx, id)
}

// Asegúrate de que los otros métodos también tengan el contexto:
func (s *propertyService) GetProperty(ctx context.Context, id int64) (*domain.Property, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *propertyService) ListProperties(ctx context.Context, page, pageSize int) ([]domain.Property, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset)
}

func (s *propertyService) UploadPropertyMedia(ctx context.Context, propertyID int64, imageURL string, mediaType string, isPrimary bool) (*domain.PropertyMedia, error) {
	if propertyID <= 0 {
		return nil, fmt.Errorf("id de propiedad inválido")
	}
	cleanURL := strings.TrimSpace(imageURL)
	if cleanURL == "" {
		return nil, fmt.Errorf("la url de la imagen es obligatoria")
	}
	parsed, err := url.ParseRequestURI(cleanURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("la url de la imagen no es válida")
	}
	if strings.TrimSpace(mediaType) == "" {
		mediaType = "IMAGE"
	}

	if _, err := s.repo.GetByID(ctx, propertyID); err != nil {
		return nil, err
	}

	if isPrimary {
		if err := s.repo.SetMainImage(ctx, propertyID, cleanURL); err != nil {
			return nil, err
		}
	}

	return &domain.PropertyMedia{
		ID:         uuid.NewString(),
		PropertyID: propertyID,
		URL:        cleanURL,
		Type:       strings.ToUpper(mediaType),
		IsPrimary:  isPrimary,
	}, nil
}
