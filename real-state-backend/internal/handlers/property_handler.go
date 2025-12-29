package handlers

import (
	"encoding/json"
	"net/http"
	"realstate_api/internal/dto"
	"realstate_api/internal/models"
	"realstate_api/internal/repository"
)

type PropertyHandler struct {
	Repo *repository.PropertyRepository
}
// internal/handlers/property_handler.go
type PropertyHandler struct {
    service ports.PropertyService
}

func NewPropertyHandler(s ports.PropertyService) *PropertyHandler {
    return &PropertyHandler{service: s}
}

func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var input dto.CreatePropertyDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Seguridad: Validar DTO
	if !input.IsValid() {
		http.Error(w, "Datos insuficientes", http.StatusUnprocessableEntity)
		return
	}

	// Convertir DTO a Modelo (M de MVC)
	property := models.Property{
		Title:       input.Title,
		Price:       input.Price,
		Description: input.Description,
	}

	// Guardar via Repositorio
	if err := h.Repo.Create(&property); err != nil {
		http.Error(w, "Error al guardar en DB", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(property)
}