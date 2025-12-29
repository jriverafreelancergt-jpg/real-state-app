package repository

import (
	"database/sql"
	"realstate_api/internal/models"
)

type PropertyRepository struct {
	DB *sql.DB
}

func (r *PropertyRepository) Create(p *models.Property) error {
	query := `INSERT INTO properties (title, price, description) VALUES ($1, $2, $3) RETURNING id`
	return r.DB.QueryRow(query, p.Title, p.Price, p.Description).Scan(&p.ID)
}