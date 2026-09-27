package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ChristianDenniss/go-data-model/auth/entity"
	"github.com/ChristianDenniss/go-data-model/auth/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ repository.Repository = (*AuthRepository)(nil)

type AuthRepository struct {
	db *DB
}

func NewAuthRepository(db *DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateAccount(ctx context.Context, in entity.NewAccount) error {
	tx, err := r.db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO accounts (id, name, email, password_hash)
		VALUES ($1, $2, $3, $4)`,
		in.ID, in.Name, in.Email, in.PasswordHash); err != nil {
		if isUniqueViolation(err, "accounts_email_lower_key") {
			return entity.ErrEmailTaken
		}
		return err
	}
	if in.Identity != nil {
		if err := insertIdentity(ctx, tx, *in.Identity); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *AuthRepository) CredentialsByEmail(ctx context.Context, email string) (entity.Credentials, error) {
	var out entity.Credentials
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT id, password_hash
		FROM accounts
		WHERE lower(email) = lower($1)`, email).Scan(&out.AccountID, &out.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Credentials{}, entity.ErrAccountNotFound
	}
	return out, err
}

func (r *AuthRepository) IdentityBySubject(ctx context.Context, provider, subject string) (entity.Identity, error) {
	out := entity.Identity{Provider: provider, Subject: subject}
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT account_id, email
		FROM account_identities
		WHERE provider = $1 AND subject = $2`, provider, subject).Scan(&out.AccountID, &out.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Identity{}, entity.ErrIdentityNotFound
	}
	return out, err
}

func (r *AuthRepository) LinkIdentity(ctx context.Context, identity entity.Identity) error {
	return insertIdentity(ctx, r.db.sql, identity)
}

func (r *AuthRepository) CreateSession(ctx context.Context, session entity.Session) error {
	_, err := r.db.sql.ExecContext(ctx, `
		INSERT INTO auth_sessions (token_hash, account_id, expires_at)
		VALUES ($1, $2, $3)`,
		session.TokenHash, session.AccountID, session.ExpiresAt)
	return err
}

func (r *AuthRepository) SessionByTokenHash(ctx context.Context, tokenHash string, now time.Time) (entity.Session, error) {
	out := entity.Session{TokenHash: tokenHash}
	err := r.db.sql.QueryRowContext(ctx, `
		SELECT account_id, expires_at
		FROM auth_sessions
		WHERE token_hash = $1 AND expires_at > $2`, tokenHash, now).Scan(&out.AccountID, &out.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Session{}, entity.ErrSessionNotFound
	}
	return out, err
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.sql.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func insertIdentity(ctx context.Context, db execer, identity entity.Identity) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO account_identities (provider, subject, account_id, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, subject) DO NOTHING`,
		identity.Provider, identity.Subject, identity.AccountID, identity.Email)
	return err
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraint
}
