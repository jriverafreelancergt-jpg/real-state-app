package domain

import "time"

// Property representa un inmueble en el sistema.
// Se usan etiquetas JSON para la respuesta de la API.
type Property struct {
	ID                   int64      `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Price                float64    `json:"price"`
	Currency             string     `json:"currency"` // USD, GTQ
	Address              string     `json:"address"`
	City                 string     `json:"city"`
	Type                 string     `json:"type"` // Casa, Apartamento, Terreno
	Bedrooms             int        `json:"bedrooms,omitempty"`
	Bathrooms            int        `json:"bathrooms,omitempty"`
	AreaSqM              float64    `json:"area_sqm"`
	Lat                  float64    `json:"lat,omitempty"` // Para mapas en la app móvil
	Lng                  float64    `json:"lng,omitempty"`
	MainImage            string     `json:"main_image"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	LastSuccessfulSyncAt *time.Time `json:"last_successful_sync_at,omitempty"` // Tracking de último sync exitoso
	LastSyncID           string     `json:"last_sync_id,omitempty"`            // UUID del batch de sincronización
	Origin               string     `json:"origin,omitempty"`                  // 'WASI', 'MANUAL', etc
	Status               string     `json:"status,omitempty"`                  // 'ACTIVE', 'INACTIVE'
}
