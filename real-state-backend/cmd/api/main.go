package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"realstate_api/internal/handlers"
	"realstate_api/internal/repository"

	_ "github.com/lib/pq" // Driver de Postgres
)

func main() {
	// 1. Conectar a Postgres (usa la variable de entorno de Docker)
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// 2. Inicializar Capas
	repo := &repository.PropertyRepository{DB: db}
	h := &handlers.PropertyHandler{Repo: repo}

	// 3. Rutas con prefijo de versión (Mejor práctica)
	http.HandleFunc("/api/v1/properties", h.CreateProperty)

	log.Println("🚀 Servidor Real State API listo en puerto 8080")
	http.ListenAndServe(":8080", nil)
}