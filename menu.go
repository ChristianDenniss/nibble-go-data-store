package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/menu/entity"
	"github.com/ChristianDenniss/go-data-model/menu/repository"
)

var _ repository.Repository = (*MenuRepository)(nil)

type MenuRepository struct {
	db *DB
}

func NewMenuRepository(db *DB) *MenuRepository {
	return &MenuRepository{db: db}
}

type menuItemRow struct {
	ID           string
	RestaurantID string
	Name         string
	Description  string
	Section      string
	ImageURL     string
}

func (row menuItemRow) toDomain() entity.Item {
	return entity.Item{
		ID:           row.ID,
		RestaurantID: row.RestaurantID,
		Name:         row.Name,
		Description:  row.Description,
		Section:      row.Section,
		ImageURL:     row.ImageURL,
	}
}

func (r *MenuRepository) GetByID(ctx context.Context, id string) (entity.Item, error) {
	var row menuItemRow
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, restaurant_id, name, description, section, image_url
		FROM menu_items
		WHERE id = $1`, id).Scan(&row.ID, &row.RestaurantID, &row.Name, &row.Description, &row.Section, &row.ImageURL)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Item{}, entity.ErrNotFound
	}
	if err != nil {
		return entity.Item{}, err
	}
	return row.toDomain(), nil
}

// Upsert keeps the stored image when the incoming item has none, so an ingest that
// doesn't carry images never wipes one that was set through the API or a seed.
func (r *MenuRepository) Upsert(ctx context.Context, item entity.Item) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO menu_items (id, restaurant_id, name, description, section, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			restaurant_id = EXCLUDED.restaurant_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			section = EXCLUDED.section,
			image_url = COALESCE(NULLIF(EXCLUDED.image_url, ''), menu_items.image_url),
			updated_at = now()`,
		item.ID, item.RestaurantID, item.Name, item.Description, item.Section, item.ImageURL)
	return err
}

func (r *MenuRepository) SetImageURL(ctx context.Context, id, imageURL string) error {
	res, err := r.db.sql.ExecContext(ctx, `
		UPDATE menu_items SET image_url = $2, updated_at = now() WHERE id = $1`, id, imageURL)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}
