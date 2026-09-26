package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/serviceability/entity"
	"github.com/ChristianDenniss/go-data-model/serviceability/repository"
)

var _ repository.StoreStatusRepository = (*SourceStoreStatusRepository)(nil)

type SourceStoreStatusRepository struct {
	db *DB
}

func NewSourceStoreStatusRepository(db *DB) *SourceStoreStatusRepository {
	return &SourceStoreStatusRepository{db: db}
}

func (r *SourceStoreStatusRepository) Get(ctx context.Context, sourceStoreID string) (entity.StoreStatus, error) {
	var out entity.StoreStatus
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT source_store_id, open_now, paused, observed_at FROM source_store_status WHERE source_store_id = $1`, sourceStoreID).
		Scan(&out.SourceStoreID, &out.OpenNow, &out.Paused, &out.ObservedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.StoreStatus{}, entity.ErrNotFound
	}
	return out, err
}

func (r *SourceStoreStatusRepository) Upsert(ctx context.Context, st entity.StoreStatus) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO source_store_status (source_store_id, open_now, paused, observed_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (source_store_id) DO UPDATE SET
			open_now = EXCLUDED.open_now, paused = EXCLUDED.paused, observed_at = EXCLUDED.observed_at`,
		st.SourceStoreID, st.OpenNow, st.Paused, st.ObservedAt)
	return err
}
