package wasi

import (
	"context"
	"encoding/json"
	"fmt"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"time"
)

// WasiAdapter implementa el puerto ports.WasiProvider.
type WasiAdapter struct {
	client *Client
	take   int
}

// Asegurar que WasiAdapter implementa la interfaz en tiempo de compilación.
var _ ports.WasiProvider = (*WasiAdapter)(nil)

// NewWasiAdapter crea un adaptador para la API de Wasi.
func NewWasiAdapter(client *Client) *WasiAdapter {
	return &WasiAdapter{
		client: client,
		take:   20, // Cantidad por defecto por página
	}
}

// FetchProperties obtiene y mapea propiedades de la API de Wasi.
func (a *WasiAdapter) FetchProperties(ctx context.Context, page int) ([]domain.Property, int, error) {
	// Consumir el endpoint de búsqueda paginada
	path := fmt.Sprintf("property/search?page=%d&take=%d", page, a.take)
	
	respData, err := a.client.Do(ctx, "GET", path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed fetching from Wasi: %w", err)
	}

	var response WasiPropertyResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal Wasi response: %w", err)
	}

	// Calcular páginas totales
	totalPages := 0
	if response.Total > 0 {
		totalPages = (response.Total + a.take - 1) / a.take
	}

	// Mapear DTOs a entidades de dominio (Capa Anti-Corrupción)
	properties := make([]domain.Property, 0, len(response.Properties))
	for _, wasiProp := range response.Properties {
		properties = append(properties, a.toDomainProperty(wasiProp))
	}

	return properties, totalPages, nil
}

// toDomainProperty convierte una propiedad de Wasi al modelo de dominio propio del sistema.
func (a *WasiAdapter) toDomainProperty(wasiProp WasiProperty) domain.Property {
	imageURL := wasiProp.MainImage.Original
	if imageURL == "" {
		imageURL = wasiProp.MainImage.URL
	}

	return domain.Property{
		ID:          parseInt64Value(wasiProp.IDProperty),
		Title:       wasiProp.Title,
		Description: wasiProp.Comment,
		Price:       parseFloatValue(wasiProp.Price),
		Currency:    a.parseCurrency(wasiProp.IDCurrency),
		Address:     wasiProp.Address,
		City:        wasiProp.City,
		Type:        wasiProp.NamePropertyType,
		Bedrooms:    parseIntValue(wasiProp.Bedrooms),
		Bathrooms:   parseIntValue(wasiProp.Bathrooms),
		AreaSqM:     parseFloatValue(wasiProp.Area),
		Lat:         parseFloatValue(wasiProp.Latitude),
		Lng:         parseFloatValue(wasiProp.Longitude),
		MainImage:   imageURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// parseCurrency traduce los IDs de moneda de Wasi a códigos de 3 caracteres estándar.
func (a *WasiAdapter) parseCurrency(v interface{}) string {
	val := parseStringValue(v)
	switch val {
	case "1":
		return "USD"
	case "2":
		return "GTQ"
	default:
		// Si Wasi ya devuelve un código de 3 letras (USD/GTQ)
		if len(val) == 3 {
			return val
		}
		return "USD" // Fallback seguro
	}
}
