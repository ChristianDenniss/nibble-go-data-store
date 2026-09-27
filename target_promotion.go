package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ChristianDenniss/go-data-model/promotion/entity"
	"github.com/ChristianDenniss/go-data-model/promotion/repository"
)

var (
	_ repository.PromotionRepository         = (*PromotionRepository)(nil)
	_ repository.MembershipProductRepository = (*MembershipProductRepository)(nil)
)

const promotionColumns = `id, channel_id, name, description, kind, fulfillment_mode, value_cents, currency, value_bps, starts_at, ends_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPromotion(row rowScanner) (entity.Promotion, error) {
	var p entity.Promotion
	err := row.Scan(&p.ID, &p.ChannelID, &p.Name, &p.Description, &p.Kind, &p.FulfillmentMode,
		&p.Value.AmountCents, &p.Value.Currency, &p.ValueBPS, &p.StartsAt, &p.EndsAt)
	return p, err
}

type PromotionRepository struct {
	db *DB
}

func NewPromotionRepository(db *DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

func (r *PromotionRepository) GetByID(ctx context.Context, id string) (entity.Promotion, error) {
	out, err := scanPromotion(r.db.sql.QueryRowContext(ctx, `SELECT `+promotionColumns+` FROM promotions WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Promotion{}, entity.ErrNotFound
	}
	return out, err
}

func (r *PromotionRepository) ListActiveByChannel(ctx context.Context, channelID string, at time.Time) ([]entity.Promotion, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT `+promotionColumns+`
		FROM promotions WHERE channel_id = $1 AND starts_at <= $2 AND ends_at >= $2`, channelID, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []entity.Promotion
	for rows.Next() {
		p, err := scanPromotion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PromotionRepository) ListActiveWithTargets(ctx context.Context, at time.Time) ([]entity.ActivePromotion, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT `+promotionColumns+`
		FROM promotions WHERE starts_at <= $1 AND ends_at >= $1
		ORDER BY starts_at DESC, id`, at)
	if err != nil {
		return nil, err
	}
	var out []entity.ActivePromotion
	index := map[string]int{}
	for rows.Next() {
		p, err := scanPromotion(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		index[p.ID] = len(out)
		out = append(out, entity.ActivePromotion{Promotion: p})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if len(out) == 0 {
		return out, nil
	}

	targets, err := r.db.sql.QueryContext(ctx, `
		SELECT t.id, t.promotion_id, COALESCE(t.place_id, ''), COALESCE(t.source_store_id, ''),
			COALESCE(t.source_item_id, ''), COALESCE(t.dish_id, ''), COALESCE(t.brand_id, ''),
			COALESCE(t.legacy_restaurant_id, '')
		FROM promotion_targets t
		JOIN promotions p ON p.id = t.promotion_id
		WHERE p.starts_at <= $1 AND p.ends_at >= $1
		ORDER BY t.id`, at)
	if err != nil {
		return nil, err
	}
	defer targets.Close()
	for targets.Next() {
		var t entity.Target
		if err := targets.Scan(&t.ID, &t.PromotionID, &t.PlaceID, &t.SourceStoreID, &t.SourceItemID, &t.DishID, &t.BrandID, &t.LegacyRestaurantID); err != nil {
			return nil, err
		}
		if i, ok := index[t.PromotionID]; ok {
			out[i].Targets = append(out[i].Targets, t)
		}
	}
	return out, targets.Err()
}

func (r *PromotionRepository) Upsert(ctx context.Context, p entity.Promotion) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO promotions (id, channel_id, name, description, kind, fulfillment_mode, value_cents, currency, value_bps, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			channel_id = EXCLUDED.channel_id, name = EXCLUDED.name, description = EXCLUDED.description,
			kind = EXCLUDED.kind, fulfillment_mode = EXCLUDED.fulfillment_mode,
			value_cents = EXCLUDED.value_cents, currency = EXCLUDED.currency, value_bps = EXCLUDED.value_bps,
			starts_at = EXCLUDED.starts_at, ends_at = EXCLUDED.ends_at, updated_at = now()`,
		p.ID, p.ChannelID, p.Name, p.Description, p.Kind, p.FulfillmentMode, p.Value.AmountCents, p.Value.Currency, p.ValueBPS, p.StartsAt, p.EndsAt)
	return err
}

func (r *PromotionRepository) UpsertConstraint(ctx context.Context, c entity.Constraint) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO promotion_constraints (id, promotion_id, min_subtotal_cents, code, membership_required, max_discount_cents)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			promotion_id = EXCLUDED.promotion_id, min_subtotal_cents = EXCLUDED.min_subtotal_cents,
			code = EXCLUDED.code, membership_required = EXCLUDED.membership_required,
			max_discount_cents = EXCLUDED.max_discount_cents`,
		c.ID, c.PromotionID, c.MinSubtotalCents, c.Code, c.MembershipRequired, c.MaxDiscountCents)
	return err
}

func (r *PromotionRepository) UpsertTarget(ctx context.Context, t entity.Target) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO promotion_targets (id, promotion_id, place_id, source_store_id, source_item_id, dish_id, brand_id, legacy_restaurant_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			promotion_id = EXCLUDED.promotion_id, place_id = EXCLUDED.place_id,
			source_store_id = EXCLUDED.source_store_id, source_item_id = EXCLUDED.source_item_id,
			dish_id = EXCLUDED.dish_id, brand_id = EXCLUDED.brand_id,
			legacy_restaurant_id = EXCLUDED.legacy_restaurant_id`,
		t.ID, t.PromotionID, nullString(t.PlaceID), nullString(t.SourceStoreID), nullString(t.SourceItemID),
		nullString(t.DishID), nullString(t.BrandID), nullString(t.LegacyRestaurantID))
	return err
}

type MembershipProductRepository struct {
	db *DB
}

func NewMembershipProductRepository(db *DB) *MembershipProductRepository {
	return &MembershipProductRepository{db: db}
}

func (r *MembershipProductRepository) GetByID(ctx context.Context, id string) (entity.MembershipProduct, error) {
	var out entity.MembershipProduct
	err := r.db.sql.QueryRowContext(ctx, `SELECT id, channel_id, name, slug FROM membership_products WHERE id = $1`, id).
		Scan(&out.ID, &out.ChannelID, &out.Name, &out.Slug)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.MembershipProduct{}, entity.ErrNotFound
	}
	return out, err
}

func (r *MembershipProductRepository) Upsert(ctx context.Context, p entity.MembershipProduct) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO membership_products (id, channel_id, name, slug) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET channel_id = EXCLUDED.channel_id, name = EXCLUDED.name, slug = EXCLUDED.slug`,
		p.ID, p.ChannelID, p.Name, p.Slug)
	return err
}
