package postgres

import (
	"context"
	"time"

	"github.com/ChristianDenniss/go-data-model/merchandising/entity"
	"github.com/ChristianDenniss/go-data-model/merchandising/repository"
)

var _ repository.Repository = (*MerchandisingRepository)(nil)

type MerchandisingRepository struct {
	db *DB
}

func NewMerchandisingRepository(db *DB) *MerchandisingRepository {
	return &MerchandisingRepository{db: db}
}

func (r *MerchandisingRepository) ActivePlacements(ctx context.Context, slot, marketID string, at time.Time) ([]entity.ServedPlacement, error) {
	rows, err := r.db.sql.QueryContext(ctx, `
		SELECT p.id, p.campaign_id, p.slot, p.priority,
			COALESCE(p.legacy_restaurant_id, ''), COALESCE(p.place_id, ''), COALESCE(p.brand_id, ''),
			COALESCE(p.promotion_id, ''), COALESCE(p.category_id, ''), COALESCE(p.cuisine_id, ''),
			p.headline, p.body, p.image_url, p.call_to_action,
			c.bid_cents, a.name
		FROM sponsored_placements p
		JOIN sponsored_campaigns c ON c.id = p.campaign_id
		JOIN advertisers a ON a.id = c.advertiser_id
		WHERE p.slot = $1
			AND c.status = 'active' AND a.status = 'active'
			AND c.starts_at <= $2 AND c.ends_at > $2
			AND ($3::text = '' OR c.market_id IS NULL OR c.market_id = $3::text)
		ORDER BY c.bid_cents DESC, p.priority DESC, p.id`, slot, at, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.ServedPlacement
	for rows.Next() {
		var s entity.ServedPlacement
		p := &s.Placement
		if err := rows.Scan(
			&p.ID, &p.CampaignID, &p.Slot, &p.Priority,
			&p.LegacyRestaurantID, &p.PlaceID, &p.BrandID,
			&p.PromotionID, &p.CategoryID, &p.CuisineID,
			&p.Headline, &p.Body, &p.ImageURL, &p.CallToAction,
			&s.BidCents, &s.AdvertiserName,
		); err != nil {
			return nil, err
		}
		s.CampaignID = p.CampaignID
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *MerchandisingRepository) UpsertAdvertiser(ctx context.Context, a entity.Advertiser) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO advertisers (id, name, brand_id, contact_email, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name, brand_id = EXCLUDED.brand_id, contact_email = EXCLUDED.contact_email,
			status = EXCLUDED.status, updated_at = now()`,
		a.ID, a.Name, nullString(a.BrandID), a.ContactEmail, a.Status)
	return err
}

func (r *MerchandisingRepository) UpsertCampaign(ctx context.Context, c entity.Campaign) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO sponsored_campaigns (id, advertiser_id, market_id, name, status, starts_at, ends_at,
			pricing_model, bid_cents, daily_budget_cents, total_budget_cents, currency)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			advertiser_id = EXCLUDED.advertiser_id, market_id = EXCLUDED.market_id, name = EXCLUDED.name,
			status = EXCLUDED.status, starts_at = EXCLUDED.starts_at, ends_at = EXCLUDED.ends_at,
			pricing_model = EXCLUDED.pricing_model, bid_cents = EXCLUDED.bid_cents,
			daily_budget_cents = EXCLUDED.daily_budget_cents, total_budget_cents = EXCLUDED.total_budget_cents,
			currency = EXCLUDED.currency, updated_at = now()`,
		c.ID, c.AdvertiserID, nullString(c.MarketID), c.Name, c.Status, c.StartsAt, c.EndsAt,
		c.PricingModel, c.BidCents, c.DailyBudgetCents, c.TotalBudgetCents, c.Currency)
	return err
}

func (r *MerchandisingRepository) UpsertPlacement(ctx context.Context, p entity.Placement) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO sponsored_placements (id, campaign_id, slot, priority, legacy_restaurant_id, place_id, brand_id,
			promotion_id, category_id, cuisine_id, headline, body, image_url, call_to_action)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			campaign_id = EXCLUDED.campaign_id, slot = EXCLUDED.slot, priority = EXCLUDED.priority,
			legacy_restaurant_id = EXCLUDED.legacy_restaurant_id, place_id = EXCLUDED.place_id,
			brand_id = EXCLUDED.brand_id, promotion_id = EXCLUDED.promotion_id,
			category_id = EXCLUDED.category_id, cuisine_id = EXCLUDED.cuisine_id,
			headline = EXCLUDED.headline, body = EXCLUDED.body, image_url = EXCLUDED.image_url,
			call_to_action = EXCLUDED.call_to_action, updated_at = now()`,
		p.ID, p.CampaignID, p.Slot, p.Priority, nullString(p.LegacyRestaurantID), nullString(p.PlaceID),
		nullString(p.BrandID), nullString(p.PromotionID), nullString(p.CategoryID), nullString(p.CuisineID),
		p.Headline, p.Body, p.ImageURL, p.CallToAction)
	return err
}

func (r *MerchandisingRepository) RecordEvent(ctx context.Context, e entity.Event) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO sponsored_events (id, placement_id, kind, user_id, surface, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING`,
		e.ID, e.PlacementID, e.Kind, nullString(e.UserID), e.Surface, e.OccurredAt)
	return err
}
