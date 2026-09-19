package wasi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"real-state-backend/config"
	"testing"
	"time"
)

func TestUnmarshalWasiPropertyResponse(t *testing.T) {
	// JSON simulando la respuesta típica de Wasi con índices de strings
	rawJSON := `{
		"status": "success",
		"total": "2",
		"0": {
			"id_property": "12345",
			"title": "Hermosa Casa Zona 10",
			"comment": "Descripción de la casa",
			"price": 250000.50,
			"id_currency": "1",
			"address": "Avenida Reforma",
			"city": "Guatemala",
			"name_property_type": "Casa",
			"bedrooms": "3",
			"bathrooms": 2,
			"area": "150.5",
			"latitude": "14.6012",
			"longitude": -90.5213,
			"main_image": {
				"url": "http://img.com/thumb.jpg",
				"original": "http://img.com/original.jpg"
			}
		},
		"1": {
			"id_property": 67890,
			"title": "Apartamento en Zona 14",
			"comment": "Apartamento moderno",
			"price": "300000",
			"id_currency": 2,
			"address": "Avenida Las Americas",
			"city": "Guatemala",
			"name_property_type": "Apartamento",
			"bedrooms": 2,
			"bathrooms": "2.5",
			"area": 120,
			"latitude": 14.5812,
			"longitude": "-90.5113",
			"main_image": {
				"url": "http://img.com/apt.jpg"
			}
		}
	}`

	var resp WasiPropertyResponse
	err := json.Unmarshal([]byte(rawJSON), &resp)
	if err != nil {
		t.Fatalf("Expected no error unmarshaling, got %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected status 'success', got %s", resp.Status)
	}

	if resp.Total != 2 {
		t.Errorf("Expected total 2, got %d", resp.Total)
	}

	if len(resp.Properties) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(resp.Properties))
	}

	// Verificar Propiedad 0 (tipos mixtos, strings/floats)
	prop0 := resp.Properties[0]
	if parseStringValue(prop0.IDProperty) != "12345" {
		t.Errorf("Expected ID 12345, got %v", prop0.IDProperty)
	}
	if parseFloatValue(prop0.Price) != 250000.50 {
		t.Errorf("Expected Price 250000.50, got %v", prop0.Price)
	}
	if parseIntValue(prop0.Bedrooms) != 3 {
		t.Errorf("Expected Bedrooms 3, got %v", prop0.Bedrooms)
	}
	if parseIntValue(prop0.Bathrooms) != 2 {
		t.Errorf("Expected Bathrooms 2, got %v", prop0.Bathrooms)
	}
	if parseFloatValue(prop0.Area) != 150.5 {
		t.Errorf("Expected Area 150.5, got %v", prop0.Area)
	}
	if parseFloatValue(prop0.Latitude) != 14.6012 {
		t.Errorf("Expected Latitude 14.6012, got %v", prop0.Latitude)
	}
	if prop0.MainImage.Original != "http://img.com/original.jpg" {
		t.Errorf("Expected main image original url, got %s", prop0.MainImage.Original)
	}

	// Verificar Propiedad 1 (tipos alternativos, ints/strings invertidos)
	prop1 := resp.Properties[1]
	if parseStringValue(prop1.IDProperty) != "67890" {
		t.Errorf("Expected ID 67890, got %v", prop1.IDProperty)
	}
	if parseFloatValue(prop1.Price) != 300000.0 {
		t.Errorf("Expected Price 300000.0, got %v", prop1.Price)
	}
	if parseIntValue(prop1.Bedrooms) != 2 {
		t.Errorf("Expected Bedrooms 2, got %v", prop1.Bedrooms)
	}
	if parseFloatValue(prop1.Area) != 120.0 {
		t.Errorf("Expected Area 120.0, got %v", prop1.Area)
	}
	if parseFloatValue(prop1.Longitude) != -90.5113 {
		t.Errorf("Expected Longitude -90.5113, got %v", prop1.Longitude)
	}
}

func TestWasiAdapterMapping(t *testing.T) {
	rawJSON := `{
		"status": "success",
		"total": 1,
		"0": {
			"id_property": "999",
			"title": "Terreno Km 15",
			"comment": "Gran oportunidad",
			"price": "75000",
			"id_currency": "2",
			"address": "Carretera al Salvador",
			"city": "Fraijanes",
			"name_property_type": "Terreno",
			"bedrooms": 0,
			"bathrooms": 0,
			"area": "1000",
			"latitude": "14.5122",
			"longitude": "-90.4122",
			"main_image": {
				"url": "http://img.com/land.jpg"
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(rawJSON))
	}))
	defer server.Close()

	client := NewClient(server.URL, "123", config.SecretToken("secret"), 5*time.Second)
	adapter := NewWasiAdapter(client)

	props, totalPages, err := adapter.FetchProperties(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if totalPages != 1 {
		t.Errorf("Expected totalPages = 1, got %d", totalPages)
	}

	if len(props) != 1 {
		t.Fatalf("Expected 1 mapped property, got %d", len(props))
	}

	p := props[0]
	if p.ID != 999 {
		t.Errorf("Expected ID 999, got %d", p.ID)
	}
	if p.Title != "Terreno Km 15" {
		t.Errorf("Expected Title, got %s", p.Title)
	}
	if p.Currency != "GTQ" {
		t.Errorf("Expected currency GTQ (mapped from id_currency=2), got %s", p.Currency)
	}
	if p.AreaSqM != 1000.0 {
		t.Errorf("Expected AreaSqM 1000, got %f", p.AreaSqM)
	}
	if p.MainImage != "http://img.com/land.jpg" {
		t.Errorf("Expected main image fallback to url, got %s", p.MainImage)
	}
}
