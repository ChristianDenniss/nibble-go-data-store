package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/location"
	"github.com/ChristianDenniss/go-data-model/restaurant"
)

var _ restaurant.Repository = (*RestaurantRepository)(nil)

type RestaurantRepository struct {
	db *DB
}

func NewRestaurantRepository(db *DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

type restaurantRow struct {
	ID        string
	Name      string
	Latitude  float64
	Longitude float64
	Address   string
}

func (row restaurantRow) toDomain() restaurant.Restaurant {
	return restaurant.Restaurant{
		ID:   row.ID,
		Name: row.Name,
		Location: location.Location{
			Latitude:  row.Latitude,
			Longitude: row.Longitude,
			Address:   row.Address,
		},
	}
}

func (r *RestaurantRepository) GetByID(ctx context.Context, id string) (restaurant.Restaurant, error) {
	var row restaurantRow
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, name, latitude, longitude, address
		FROM restaurants
		WHERE id = $1`, id).Scan(&row.ID, &row.Name, &row.Latitude, &row.Longitude, &row.Address)
	if errors.Is(err, sql.ErrNoRows) {
		return restaurant.Restaurant{}, restaurant.ErrNotFound
	}
	if err != nil {
		return restaurant.Restaurant{}, err
	}
	return row.toDomain(), nil
}

func (r *RestaurantRepository) Upsert(ctx context.Context, in restaurant.Restaurant) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO restaurants (id, name, latitude, longitude, address)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			address = EXCLUDED.address,
			updated_at = now()`,
		in.ID, in.Name, in.Location.Latitude, in.Location.Longitude, in.Location.Address)
	return err
}
