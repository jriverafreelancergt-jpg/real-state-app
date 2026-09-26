package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"real-state-backend/internal/dto"
	"strconv"
	"strings"
)

type PropertyHandler struct {
	service ports.PropertyService
}

func NewPropertyHandler(s ports.PropertyService) *PropertyHandler {
	return &PropertyHandler{service: s}
}

const (
	defaultPageSize = 10
	minPageSize     = 1
	maxPageSize     = 100
)

// GetAll lista todas las propiedades con paginación configurable
func (h *PropertyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Validación de página
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Validación de limit (nuevo parámetro configurable)
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < minPageSize {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	properties, err := h.service.ListProperties(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al listar propiedades", "list_properties_error", "property", nil)
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
		writeError(w, http.StatusBadRequest, "ID inválido", "invalid_id", "property", nil)
		return
	}

	property, err := h.service.GetProperty(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Propiedad no encontrada", "property_not_found", "property", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(property)
}

// CreateProperty: Registro de nuevas propiedades desde la App
func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var input dto.CreatePropertyDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json", "property", nil)
		return
	}

	// Usar Validate() en lugar de IsValid()
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error(), "validation_error", "property", nil)
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
		writeError(w, http.StatusInternalServerError, "Error de base de datos", "db_error", "property", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(property)
}

func (h *PropertyHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID inválido", "invalid_id", "property", nil)
		return
	}

	var input dto.CreatePropertyDTO
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json", "property", nil)
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error(), "validation_error", "property", nil)
		return
	}

	property := &domain.Property{
		Title:       input.Title,
		Description: input.Description,
		Price:       input.Price,
		Currency:    input.Currency,
		Address:     input.Location,
		Type:        input.Type,
	}

	updated, err := h.service.UpdateProperty(r.Context(), id, property)
	if err != nil {
		slog.Error("Error updating property", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "Error al actualizar la propiedad", "update_property_error", "property", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *PropertyHandler) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID inválido", "invalid_id", "property", nil)
		return
	}

	if err := h.service.DeleteProperty(r.Context(), id); err != nil {
		slog.Error("Error deleting property", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "Propiedad no encontrada", "property_not_found", "property", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Propiedad eliminada"})
}

func (h *PropertyHandler) UploadPropertyMedia(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "ID inválido", "invalid_id", "property", nil)
		return
	}

	var input struct {
		URL       string `json:"url"`
		ImageURL  string `json:"image_url"`
		MainImage string `json:"main_image"`
		MediaType string `json:"media_type"`
		Type      string `json:"type"`
		IsPrimary bool   `json:"is_primary"`
		Primary   bool   `json:"primary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json", "property", nil)
		return
	}

	imageURL := strings.TrimSpace(input.URL)
	if imageURL == "" {
		imageURL = strings.TrimSpace(input.ImageURL)
	}
	if imageURL == "" {
		imageURL = strings.TrimSpace(input.MainImage)
	}
	if imageURL == "" {
		writeError(w, http.StatusUnprocessableEntity, "la url de la imagen es obligatoria", "validation_error", "property", nil)
		return
	}

	mediaType := strings.TrimSpace(input.MediaType)
	if mediaType == "" {
		mediaType = strings.TrimSpace(input.Type)
	}
	if mediaType == "" {
		mediaType = "IMAGE"
	}

	isPrimary := input.IsPrimary || input.Primary
	media, err := h.service.UploadPropertyMedia(r.Context(), id, imageURL, mediaType, isPrimary)
	if err != nil {
		slog.Error("Error uploading property media", "id", id, "error", err)
		writeError(w, http.StatusBadRequest, err.Error(), "upload_property_media_error", "property", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(media)
}
