package postgres

import (
	"context"
	"time"
)

type GmailConnection struct {
	AccountID    string
	Email        string
	RefreshToken string
	Active       bool
	LastSyncAt   *time.Time
}

type GmailOAuthState struct {
	State     string
	AccountID string
	ExpiresAt time.Time
}

type DealEmailMessage struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"-"`
	ExternalID  string    `json:"externalId"`
	Sender      string    `json:"sender"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ReceivedAt  time.Time `json:"receivedAt"`
	Status      string    `json:"status"`
	PromotionID string    `json:"promotionId"`
	ParseError  string    `json:"parseError"`
}

func (r *DB) UpsertGmailConnection(ctx context.Context, connection GmailConnection) error {
	_, err := r.sql.ExecContext(ctx, `INSERT INTO gmail_connections (account_id, email, refresh_token, active) VALUES ($1,$2,$3,$4)
		ON CONFLICT (account_id) DO UPDATE SET email=EXCLUDED.email, refresh_token=EXCLUDED.refresh_token, active=EXCLUDED.active, updated_at=now()`,
		connection.AccountID, connection.Email, connection.RefreshToken, connection.Active)
	return err
}

func (r *DB) GmailConnection(ctx context.Context, accountID string) (GmailConnection, error) {
	var out GmailConnection
	err := r.sql.QueryRowContext(ctx, `SELECT account_id, email, refresh_token, active, last_sync_at FROM gmail_connections WHERE account_id=$1`, accountID).
		Scan(&out.AccountID, &out.Email, &out.RefreshToken, &out.Active, &out.LastSyncAt)
	return out, err
}

func (r *DB) DeleteGmailConnection(ctx context.Context, accountID string) error {
	_, err := r.sql.ExecContext(ctx, `DELETE FROM gmail_connections WHERE account_id=$1`, accountID)
	return err
}

func (r *DB) SetGmailLastSync(ctx context.Context, accountID string, at time.Time) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE gmail_connections SET last_sync_at=$2, updated_at=now() WHERE account_id=$1`, accountID, at)
	return err
}

func (r *DB) CreateGmailOAuthState(ctx context.Context, state GmailOAuthState) error {
	_, err := r.sql.ExecContext(ctx, `INSERT INTO gmail_oauth_states (state, account_id, expires_at) VALUES ($1,$2,$3)`, state.State, state.AccountID, state.ExpiresAt)
	return err
}

func (r *DB) ConsumeGmailOAuthState(ctx context.Context, state string, now time.Time) (GmailOAuthState, error) {
	tx, err := r.sql.BeginTx(ctx, nil)
	if err != nil {
		return GmailOAuthState{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var out GmailOAuthState
	err = tx.QueryRowContext(ctx, `SELECT state, account_id, expires_at FROM gmail_oauth_states WHERE state=$1 AND expires_at>$2`, state, now).Scan(&out.State, &out.AccountID, &out.ExpiresAt)
	if err != nil {
		return GmailOAuthState{}, err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM gmail_oauth_states WHERE state=$1`, state); err != nil {
		return GmailOAuthState{}, err
	}
	return out, tx.Commit()
}

func (r *DB) InsertDealEmailMessage(ctx context.Context, message DealEmailMessage) (bool, error) {
	result, err := r.sql.ExecContext(ctx, `INSERT INTO deal_email_messages (id, account_id, external_id, sender, subject, body, received_at, status, promotion_id, parse_error) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (account_id, external_id) DO NOTHING`, message.ID, message.AccountID, message.ExternalID, message.Sender, message.Subject, message.Body, message.ReceivedAt, message.Status, message.PromotionID, message.ParseError)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (r *DB) ListDealEmailMessages(ctx context.Context, accountID string, limit int) ([]DealEmailMessage, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT id, external_id, sender, subject, body, received_at, status, promotion_id, parse_error FROM deal_email_messages WHERE account_id=$1 ORDER BY received_at DESC LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DealEmailMessage{}
	for rows.Next() {
		var message DealEmailMessage
		if err := rows.Scan(&message.ID, &message.ExternalID, &message.Sender, &message.Subject, &message.Body, &message.ReceivedAt, &message.Status, &message.PromotionID, &message.ParseError); err != nil {
			return nil, err
		}
		out = append(out, message)
	}
	return out, rows.Err()
}

func (r *DB) UpdateDealEmailMessage(ctx context.Context, accountID, externalID, status, promotionID, parseError string) error {
	_, err := r.sql.ExecContext(ctx, `UPDATE deal_email_messages SET status=$3, promotion_id=$4, parse_error=$5 WHERE account_id=$1 AND external_id=$2`, accountID, externalID, status, promotionID, parseError)
	return err
}
