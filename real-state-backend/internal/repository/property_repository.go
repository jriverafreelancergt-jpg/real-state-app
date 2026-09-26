package repository

import (
	"context"
	"database/sql"
	"real-state-backend/internal/core/domain"
	"real-state-backend/internal/core/ports"
	"real-state-backend/pkg/database"
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
		return nil, database.HandleError(ctx, err, "GetByID", "properties", map[string]interface{}{"id": id})
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
		return nil, database.HandleError(ctx, err, "GetAll", "properties", map[string]interface{}{"limit": limit, "offset": offset})
	}
	defer rows.Close()

	var properties []domain.Property
	for rows.Next() {
		var p domain.Property
		err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price, &p.Currency,
			&p.Address, &p.City, &p.Type, &p.Bedrooms, &p.Bathrooms,
			&p.AreaSqM, &p.MainImage, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, database.HandleError(ctx, err, "GetAll (scan)", "properties", map[string]interface{}{"limit": limit, "offset": offset})
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

	err := r.db.QueryRowContext(ctx, query,
		property.Title, property.Description, property.Price, property.Currency,
		property.Address, property.City, property.Type, property.Bedrooms,
		property.Bathrooms, property.AreaSqM, property.MainImage).
		Scan(&property.ID, &property.CreatedAt, &property.UpdatedAt)
	if err != nil {
		return database.HandleError(ctx, err, "Create", "properties", map[string]interface{}{"title": property.Title, "currency": property.Currency})
	}
	return nil
}

func (r *propertyRepo) Update(ctx context.Context, id int64, property *domain.Property) (*domain.Property, error) {
	query := `UPDATE properties
			SET title = $1,
			    description = $2,
			    price = $3,
			    currency = $4,
			    address = $5,
			    city = $6,
			    type = $7,
			    bedrooms = $8,
			    bathrooms = $9,
			    area_sqm = $10,
			    main_image = $11,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $12
			RETURNING id, title, description, price, currency, address, city, type,
			          bedrooms, bathrooms, area_sqm, main_image, created_at, updated_at`

	var updated domain.Property
	err := r.db.QueryRowContext(ctx, query,
		property.Title,
		property.Description,
		property.Price,
		property.Currency,
		property.Address,
		property.City,
		property.Type,
		property.Bedrooms,
		property.Bathrooms,
		property.AreaSqM,
		property.MainImage,
		id,
	).Scan(
		&updated.ID,
		&updated.Title,
		&updated.Description,
		&updated.Price,
		&updated.Currency,
		&updated.Address,
		&updated.City,
		&updated.Type,
		&updated.Bedrooms,
		&updated.Bathrooms,
		&updated.AreaSqM,
		&updated.MainImage,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, database.HandleError(ctx, err, "Update", "properties", map[string]interface{}{"id": id})
		}
		return nil, database.HandleError(ctx, err, "Update", "properties", map[string]interface{}{"id": id})
	}

	return &updated, nil
}

func (r *propertyRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM properties WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return database.HandleError(ctx, err, "Delete", "properties", map[string]interface{}{"id": id})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return database.HandleError(ctx, err, "Delete (rows affected)", "properties", map[string]interface{}{"id": id})
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *propertyRepo) UpsertBatch(ctx context.Context, properties []domain.Property, syncBatchID string) error {
	if len(properties) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return database.HandleError(ctx, err, "UpsertBatch (begin tx)", "properties", nil)
	}
	defer tx.Rollback()

	query := `INSERT INTO properties 
		(id, title, description, price, currency, address, city, type, bedrooms, bathrooms, area_sqm, lat, lng, main_image, last_sync_id, origin, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) 
		ON CONFLICT (id) DO UPDATE SET 
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			type = EXCLUDED.type,
			bedrooms = EXCLUDED.bedrooms,
			bathrooms = EXCLUDED.bathrooms,
			area_sqm = EXCLUDED.area_sqm,
			lat = EXCLUDED.lat,
			lng = EXCLUDED.lng,
			main_image = EXCLUDED.main_image,
			last_sync_id = EXCLUDED.last_sync_id,
			origin = EXCLUDED.origin,
			status = EXCLUDED.status,
			updated_at = CURRENT_TIMESTAMP`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return database.HandleError(ctx, err, "UpsertBatch (prepare statement)", "properties", nil)
	}
	defer stmt.Close()

	for _, p := range properties {
		_, err := stmt.ExecContext(ctx,
			p.ID, p.Title, p.Description, p.Price, p.Currency, p.Address, p.City, p.Type,
			p.Bedrooms, p.Bathrooms, p.AreaSqM, p.Lat, p.Lng, p.MainImage, syncBatchID, "WASI", "ACTIVE")
		if err != nil {
			return database.HandleError(ctx, err, "UpsertBatch (exec)", "properties", map[string]interface{}{"property_id": p.ID})
		}
	}

	if err := tx.Commit(); err != nil {
		return database.HandleError(ctx, err, "UpsertBatch (commit)", "properties", nil)
	}

	return nil
}

func (r *propertyRepo) Sweep(ctx context.Context, syncBatchID string, origin string) (int64, error) {
	query := `UPDATE properties 
		SET status = 'INACTIVE', updated_at = CURRENT_TIMESTAMP 
		WHERE origin = $1 AND (last_sync_id IS NULL OR last_sync_id != $2) AND status = 'ACTIVE'`

	res, err := r.db.ExecContext(ctx, query, origin, syncBatchID)
	if err != nil {
		return 0, database.HandleError(ctx, err, "Sweep", "properties", map[string]interface{}{"syncBatchID": syncBatchID, "origin": origin})
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, database.HandleError(ctx, err, "Sweep (rows affected)", "properties", nil)
	}

	return rowsAffected, nil
}

// GetLastSyncMetadata obtiene los metadatos del último sync completado exitosamente
func (r *propertyRepo) GetLastSyncMetadata(ctx context.Context) (*domain.SyncMetadata, error) {
	query := `SELECT id, batch_id, sync_started_at, sync_completed_at, properties_synced, 
	                properties_deactivated, status, error_message, executed_by, created_at, updated_at
	          FROM sync_metadata 
	          WHERE status = 'COMPLETED' 
	          ORDER BY sync_completed_at DESC LIMIT 1`

	var meta domain.SyncMetadata
	err := r.db.QueryRowContext(ctx, query).Scan(
		&meta.ID, &meta.BatchID, &meta.SyncStartedAt, &meta.SyncCompletedAt,
		&meta.PropertiesSynced, &meta.PropertiesDeactivated, &meta.Status,
		&meta.ErrorMessage, &meta.ExecutedBy, &meta.CreatedAt, &meta.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No hay sync previo
		}
		return nil, database.HandleError(ctx, err, "GetLastSyncMetadata", "sync_metadata", nil)
	}
	return &meta, nil
}

// CreateSyncMetadata registra el inicio de una nueva sincronización
func (r *propertyRepo) CreateSyncMetadata(ctx context.Context, batchID string, executedBy *string) error {
	query := `INSERT INTO sync_metadata (batch_id, status, executed_by) 
	          VALUES ($1, 'IN_PROGRESS', $2)`

	_, err := r.db.ExecContext(ctx, query, batchID, executedBy)
	if err != nil {
		return database.HandleError(ctx, err, "CreateSyncMetadata", "sync_metadata", map[string]interface{}{"batch_id": batchID})
	}
	return nil
}

// CompleteSyncMetadata marca una sincronización como completada exitosamente
func (r *propertyRepo) CompleteSyncMetadata(ctx context.Context, batchID string, propertiesSynced, propertiesDeactivated int) error {
	query := `UPDATE sync_metadata 
	          SET status = 'COMPLETED', sync_completed_at = CURRENT_TIMESTAMP, 
	              properties_synced = $1, properties_deactivated = $2
	          WHERE batch_id = $3`

	_, err := r.db.ExecContext(ctx, query, propertiesSynced, propertiesDeactivated, batchID)
	if err != nil {
		return database.HandleError(ctx, err, "CompleteSyncMetadata", "sync_metadata", map[string]interface{}{"batch_id": batchID})
	}
	return nil
}

// FailSyncMetadata marca una sincronización como fallida
func (r *propertyRepo) FailSyncMetadata(ctx context.Context, batchID string, errorMessage string) error {
	query := `UPDATE sync_metadata 
	          SET status = 'FAILED', sync_completed_at = CURRENT_TIMESTAMP, error_message = $1
	          WHERE batch_id = $2`

	_, err := r.db.ExecContext(ctx, query, errorMessage, batchID)
	if err != nil {
		return database.HandleError(ctx, err, "FailSyncMetadata", "sync_metadata", map[string]interface{}{"batch_id": batchID})
	}
	return nil
}

// GetActivePropertiesCount retorna el total de propiedades activas
func (r *propertyRepo) GetActivePropertiesCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM properties WHERE status = 'ACTIVE'`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, database.HandleError(ctx, err, "GetActivePropertiesCount", "properties", nil)
	}
	return count, nil
}

func (r *propertyRepo) SetMainImage(ctx context.Context, id int64, url string) error {
	query := `UPDATE properties
			SET main_image = $1,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, url, id)
	if err != nil {
		return database.HandleError(ctx, err, "SetMainImage", "properties", map[string]interface{}{"id": id, "url": url})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return database.HandleError(ctx, err, "SetMainImage (rows affected)", "properties", map[string]interface{}{"id": id})
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *propertyRepo) UploadPropertyMedia(ctx context.Context, propertyID int64, url string, mediaType string, isPrimary bool) (*domain.PropertyMedia, error) {
	if _, err := r.GetByID(ctx, propertyID); err != nil {
		return nil, err
	}

	if isPrimary {
		if err := r.SetMainImage(ctx, propertyID, url); err != nil {
			return nil, err
		}
	}

	return &domain.PropertyMedia{
		PropertyID: propertyID,
		URL:        url,
		Type:       mediaType,
		IsPrimary:  isPrimary,
	}, nil
}
