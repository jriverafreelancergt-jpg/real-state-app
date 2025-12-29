package models

type Property struct {
    ID          uint    `json:"id"`
    Title       string  `json:"title"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
    SecretNote  string  `json:"-"` // El "-" oculta este campo del JSON (Seguridad)
}