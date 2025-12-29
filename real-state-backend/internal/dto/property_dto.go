package dto

// CreatePropertyDTO filtra lo que recibimos de Flutter
type CreatePropertyDTO struct {
    Title string  `json:"title"`
    Price float64 `json:"price"`
}