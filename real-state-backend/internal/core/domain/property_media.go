package domain

import "time"

// PropertyMedia representa un recurso multimedia asociado a una propiedad.
// Se usa para registrar la URL principal de la imagen y mantener el contrato
// del cliente móvil sin introducir una tabla adicional en la base de datos.
type PropertyMedia struct {
	ID         string    `json:"id"`
	PropertyID int64     `json:"property_id"`
	URL        string    `json:"url"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsPrimary  bool      `json:"is_primary"`
}
