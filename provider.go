package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ChristianDenniss/go-data-model/provider"
)

var _ provider.Repository = (*ProviderRepository)(nil)

type ProviderRepository struct {
	db *DB
}

func NewProviderRepository(db *DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

type providerRow struct {
	ID   string
	Name string
}

func (row providerRow) toDomain() provider.Provider {
	return provider.Provider{ID: row.ID, Name: row.Name}
}

func (r *ProviderRepository) GetByID(ctx context.Context, id string) (provider.Provider, error) {
	var row providerRow
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, name
		FROM providers
		WHERE id = $1`, id).Scan(&row.ID, &row.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return provider.Provider{}, provider.ErrNotFound
	}
	if err != nil {
		return provider.Provider{}, err
	}
	return row.toDomain(), nil
}

func (r *ProviderRepository) Upsert(ctx context.Context, p provider.Provider) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO providers (id, name)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = now()`,
		p.ID, p.Name)
	return err
}
