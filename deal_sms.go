package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type DealSMSSource struct {
	PhoneNumber      string `json:"phoneNumber"`
	ProviderID       string `json:"providerId"`
	Label            string `json:"label"`
	ExpiryPolicy     string `json:"expiryPolicy"`
	FixedExpiryHours int    `json:"fixedExpiryHours"`
	Active           bool   `json:"active"`
}
type DealSMSMessage struct {
	ID          string    `json:"id"`
	ExternalID  string    `json:"externalId"`
	FromNumber  string    `json:"fromNumber"`
	ToNumber    string    `json:"toNumber"`
	ProviderID  string    `json:"providerId"`
	Body        string    `json:"body"`
	Status      string    `json:"status"`
	PromotionID string    `json:"promotionId"`
	ParseError  string    `json:"parseError"`
	ReceivedAt  time.Time `json:"receivedAt"`
}

func (r *DB) UpsertDealSMSSource(ctx context.Context, source DealSMSSource) error {
	_, err := r.sql.ExecContext(ctx, `INSERT INTO deal_sms_sources (phone_number, provider_id, label, expiry_policy, fixed_expiry_hours, active) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (phone_number) DO UPDATE SET provider_id=EXCLUDED.provider_id,label=EXCLUDED.label,expiry_policy=EXCLUDED.expiry_policy,fixed_expiry_hours=EXCLUDED.fixed_expiry_hours,active=EXCLUDED.active`, source.PhoneNumber, source.ProviderID, source.Label, source.ExpiryPolicy, source.FixedExpiryHours, source.Active)
	return err
}

func (r *DB) GetDealSMSSource(ctx context.Context, phone string) (DealSMSSource, error) {
	var out DealSMSSource
	err := r.sql.QueryRowContext(ctx, `SELECT phone_number, provider_id, label, expiry_policy, fixed_expiry_hours, active FROM deal_sms_sources WHERE phone_number=$1`, phone).Scan(&out.PhoneNumber, &out.ProviderID, &out.Label, &out.ExpiryPolicy, &out.FixedExpiryHours, &out.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return out, sql.ErrNoRows
	}
	return out, err
}

func (r *DB) ListDealSMSSources(ctx context.Context) ([]DealSMSSource, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT phone_number, provider_id, label, expiry_policy, fixed_expiry_hours, active FROM deal_sms_sources ORDER BY provider_id, phone_number`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DealSMSSource
	for rows.Next() {
		var source DealSMSSource
		if err := rows.Scan(&source.PhoneNumber, &source.ProviderID, &source.Label, &source.ExpiryPolicy, &source.FixedExpiryHours, &source.Active); err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	return out, rows.Err()
}

func (r *DB) InsertDealSMSMessage(ctx context.Context, message DealSMSMessage) error {
	_, err := r.sql.ExecContext(ctx, `INSERT INTO deal_sms_messages (id, external_id, from_number, to_number, provider_id, body, received_at, status, promotion_id, parse_error) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (external_id) DO NOTHING`, message.ID, message.ExternalID, message.FromNumber, message.ToNumber, message.ProviderID, message.Body, message.ReceivedAt, message.Status, message.PromotionID, message.ParseError)
	return err
}

func (r *DB) GetDealSMSMessage(ctx context.Context, externalID string) (DealSMSMessage, error) {
	var out DealSMSMessage
	err := r.sql.QueryRowContext(ctx, `SELECT id, external_id, from_number, to_number, provider_id, body, status, promotion_id, parse_error, received_at FROM deal_sms_messages WHERE external_id=$1`, externalID).Scan(&out.ID, &out.ExternalID, &out.FromNumber, &out.ToNumber, &out.ProviderID, &out.Body, &out.Status, &out.PromotionID, &out.ParseError, &out.ReceivedAt)
	return out, err
}

func (r *DB) ListDealSMSMessages(ctx context.Context, limit int) ([]DealSMSMessage, error) {
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := r.sql.QueryContext(ctx, `SELECT id, external_id, from_number, to_number, provider_id, body, status, promotion_id, parse_error, received_at FROM deal_sms_messages ORDER BY received_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DealSMSMessage
	for rows.Next() {
		var message DealSMSMessage
		if err := rows.Scan(&message.ID, &message.ExternalID, &message.FromNumber, &message.ToNumber, &message.ProviderID, &message.Body, &message.Status, &message.PromotionID, &message.ParseError, &message.ReceivedAt); err != nil {
			return nil, err
		}
		out = append(out, message)
	}
	return out, rows.Err()
}

func (r *DB) UpdateDealSMSMessage(ctx context.Context, externalID, status, promotionID, parseError string) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE deal_sms_messages SET status=$2, promotion_id=$3, parse_error=$4 WHERE external_id=$1`, externalID, status, promotionID, parseError)
	return err
}
