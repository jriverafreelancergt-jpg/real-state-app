package dto

import (
	"errors"
	"slices"
)

// CreatePropertyDTO define la estructura de datos que esperamos del móvil
// Usamos etiquetas `json` para que Go sepa cómo mapear el cuerpo del request.
type CreatePropertyDTO struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Location    string  `json:"location"`
	Type        string  `json:"type"`
}

// IsValid realiza una validación básica de seguridad de los datos de entrada
func (d *CreatePropertyDTO) IsValid() bool {
	// Reglas de negocio básicas: Título no vacío y precio positivo
	if d.Title == "" || d.Price <= 0 || len(d.Description) < 10 {
		return false
	}
	return true
}
func (d *CreatePropertyDTO) Validate() error {
	if d.Title == "" {
		return errors.New("title is required")
	}
	if len(d.Title) < 3 {
		return errors.New("title must be at least 3 characters")
	}
	if d.Price <= 0 {
		return errors.New("price must be positive")
	}
	// Validar moneda
	allowedCurrencies := []string{"USD", "GTQ"}
	if !slices.Contains(allowedCurrencies, d.Currency) {
		return errors.New("invalid currency")
	}
	// Validar tipo de propiedad
	allowedTypes := []string{"Casa", "Apartamento", "Terreno", "Oficina"}
	if !slices.Contains(allowedTypes, d.Type) {
		return errors.New("invalid type")
	}
	return nil

}
