package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"real-state-backend/internal/dto"
	"strconv"
	//"strings"
)

type PropertyHandler struct {
	service ports.PropertyService
}

func NewPropertyHandler(s ports.PropertyService) *PropertyHandler {
	return &PropertyHandler{service: s}
}

// GetAll: Resuelve el error de "undefined GetAll" en main.go
func (h *PropertyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Seguridad: Validamos y sanitizamos parámetros de paginación
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	properties, err := h.service.ListProperties(r.Context(), page, 10)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error al listar propiedades"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(properties)
}

// GetByID: Resuelve el error de "undefined GetByID" en main.go
func (h *PropertyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extracción segura del ID desde la URL
	idStr := r.PathValue("id") // Go 1.22 feature
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	property, err := h.service.GetProperty(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Propiedad no encontrada"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(property)
}

// CreateProperty: Registro de nuevas propiedades desde la App
func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var input dto.CreatePropertyDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Usar Validate() en lugar de IsValid()
	if err := input.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	property := &domain.Property{
		Title:       input.Title,
		Price:       input.Price,
		Description: input.Description,
		Currency:    input.Currency,
		Address:     input.Location,
		Type:        input.Type,
	}

	slog.Info("Creating property", "title", property.Title, "price", property.Price, "currency", property.Currency, "address", property.Address)

	if err := h.service.CreateProperty(r.Context(), property); err != nil {
		slog.Error("Error creating property", "error", err)
		http.Error(w, "Error de base de datos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(property)
}
