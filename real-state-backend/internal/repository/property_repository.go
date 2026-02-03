package repository

import (
	"context"
	"database/sql"
	"errors"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
)

type propertyRepo struct {
	db *sql.DB
}

// NewPropertyRepository crea una instancia del repositorio.
func NewPropertyRepository(db *sql.DB) ports.PropertyRepository {
	return &propertyRepo{db: db}
}

func (r *propertyRepo) GetByID(ctx context.Context, id int64) (*domain.Property, error) {
	// Query parametrizada: INMUNE a SQL Injection
	query := `SELECT id, title, price, address, type, created_at FROM properties WHERE id = $1`

	var p domain.Property
	// Usamos QueryRowContext para respetar el timeout del contexto
	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Title, &p.Price, &p.Address, &p.Type, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("property not found")
		}
		return nil, err
	}
	return &p, nil
}

func (r *propertyRepo) GetAll(ctx context.Context, limit, offset int) ([]domain.Property, error) {
	query := `SELECT id, title, description, price, currency, address, city, type, 
                     bedrooms, bathrooms, area_sqm, main_image, created_at, updated_at 
              FROM properties 
              ORDER BY created_at DESC 
              LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []domain.Property
	for rows.Next() {
		var p domain.Property
		err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Currency,
			&p.Address, &p.City, &p.Type, &p.Bedrooms, &p.Bathrooms,
			&p.AreaSqM, &p.MainImage, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}

	return properties, nil
}

func (r *propertyRepo) Create(ctx context.Context, property *domain.Property) error {
	query := `INSERT INTO properties 
              (title, description, price, currency, address, city, type, 
               bedrooms, bathrooms, area_sqm, main_image) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
              RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		property.Title, property.Description, property.Price, property.Currency,
		property.Address, property.City, property.Type, property.Bedrooms,
		property.Bathrooms, property.AreaSqM, property.MainImage).
		Scan(&property.ID, &property.CreatedAt, &property.UpdatedAt)
}
