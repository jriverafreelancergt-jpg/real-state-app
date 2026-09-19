package wasi

import (
	"encoding/json"
	"strconv"
)

// WasiPropertyResponse envuelve la respuesta de búsqueda/listado de Wasi.
// Wasi API retorna un JSON de tipo clave-valor dinámico en vez de un array normal.
type WasiPropertyResponse struct {
	Status     string         `json:"status"`
	Total      int            `json:"total"`
	Properties []WasiProperty `json:"-"`
}

// WasiMainImage representa la imagen principal de un inmueble en Wasi.
type WasiMainImage struct {
	URL      string `json:"url"`
	Original string `json:"original"`
}

// WasiProperty contiene los campos crudos retornados por la API de Wasi.
// Usamos interfaces (interface{}) para los tipos numéricos debido a que la API de Wasi
// es inconsistente y puede retornar números como strings o como enteros/flotantes reales.
type WasiProperty struct {
	IDProperty       interface{}   `json:"id_property"`
	Title            string        `json:"title"`
	Comment          string        `json:"comment"` // Descripción o comentario
	Price            interface{}   `json:"price"`
	IDCurrency       interface{}   `json:"id_currency"` // Identificador de la moneda (USD, GTQ, etc)
	Address          string        `json:"address"`
	City             string        `json:"city"`
	NamePropertyType string        `json:"name_property_type"`
	Bedrooms         interface{}   `json:"bedrooms"`
	Bathrooms        interface{}   `json:"bathrooms"`
	Area             interface{}   `json:"area"`
	Latitude         interface{}   `json:"latitude"`
	Longitude        interface{}   `json:"longitude"`
	MainImage        WasiMainImage `json:"main_image"`
}

// UnmarshalJSON implementa json.Unmarshaler para manejar las inconsistencias de la API de Wasi
// (como retornar listas de propiedades como objetos con índices numéricos en las llaves).
func (r *WasiPropertyResponse) UnmarshalJSON(data []byte) error {
	// Primero parseamos como un mapa genérico de json.RawMessage
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	// Extraer status
	if statusRaw, exists := rawMap["status"]; exists {
		var status string
		if err := json.Unmarshal(statusRaw, &status); err == nil {
			r.Status = status
		}
	}

	// Extraer total (puede venir como int o como string)
	if totalRaw, exists := rawMap["total"]; exists {
		var totalInt int
		if err := json.Unmarshal(totalRaw, &totalInt); err != nil {
			var totalStr string
			if err := json.Unmarshal(totalRaw, &totalStr); err == nil {
				totalInt, _ = strconv.Atoi(totalStr)
			}
		}
		r.Total = totalInt
	}

	// El resto de llaves que representan índices numéricos son las propiedades
	r.Properties = make([]WasiProperty, 0)
	for key, valRaw := range rawMap {
		// Ignorar campos no dinámicos conocidos de la respuesta
		if key == "status" || key == "total" || key == "message" || key == "code" {
			continue
		}

		// Si la llave es un número (índices que usa Wasi como "0", "1", "24"), es una propiedad.
		if _, err := strconv.Atoi(key); err == nil {
			var prop WasiProperty
			if err := json.Unmarshal(valRaw, &prop); err == nil {
				r.Properties = append(r.Properties, prop)
			}
		}
	}

	return nil
}

// Helpers defensivos para parsear los campos altamente dinámicos e inconsistentes de Wasi

func parseFloatValue(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func parseIntValue(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case float32:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	}
	return 0
}

func parseInt64Value(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	}
	return 0
}

func parseStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	}
	return ""
}
