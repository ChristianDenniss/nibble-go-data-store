package postgres

import (
	"context"
	"strconv"
	"strings"

	"github.com/ChristianDenniss/go-data-model/serviceability/entity"
	"github.com/ChristianDenniss/go-data-model/serviceability/repository"
)

var _ repository.RestaurantStorePathRepository = (*RestaurantStorePathRepository)(nil)

type RestaurantStorePathRepository struct {
	db *DB
}

func NewRestaurantStorePathRepository(db *DB) *RestaurantStorePathRepository {
	return &RestaurantStorePathRepository{db: db}
}

func (r *RestaurantStorePathRepository) ListByRestaurants(ctx context.Context, restaurantIDs []string) ([]entity.RestaurantStorePath, error) {
	if len(restaurantIDs) == 0 {
		return nil, nil
	}
	args := make([]any, len(restaurantIDs))
	placeholders := make([]string, len(restaurantIDs))
	for i, id := range restaurantIDs {
		args[i] = id
		placeholders[i] = "$" + strconv.Itoa(i+1)
	}
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT restaurant_id, provider_id, source_store_id
		FROM restaurant_source_store_paths
		WHERE restaurant_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY restaurant_id, provider_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.RestaurantStorePath
	for rows.Next() {
		var path entity.RestaurantStorePath
		if err := rows.Scan(&path.RestaurantID, &path.ProviderID, &path.SourceStoreID); err != nil {
			return nil, err
		}
		out = append(out, path)
	}
	return out, rows.Err()
}

func (r *RestaurantStorePathRepository) Upsert(ctx context.Context, path entity.RestaurantStorePath) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO restaurant_source_store_paths (restaurant_id, provider_id, source_store_id, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (restaurant_id, provider_id) DO UPDATE SET source_store_id = EXCLUDED.source_store_id, updated_at = now()`,
		path.RestaurantID, path.ProviderID, path.SourceStoreID)
	return err
}
