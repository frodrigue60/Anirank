package postgres

import (
	"context"
	"anirank/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type authTokenRepository struct {
	db *sqlx.DB
}

func NewAuthTokenRepository(db *sqlx.DB) domain.AuthTokenRepository {
	return &authTokenRepository{db: db}
}

func (r *authTokenRepository) Create(ctx context.Context, token *domain.AuthToken) error {
	query := `
		INSERT INTO auth_tokens (user_id, token, type, expires_at, created_at)
		VALUES (:user_id, :token, :type, :expires_at, NOW())
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, token)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&token.ID)
	}
	return err
}

func (r *authTokenRepository) GetByToken(ctx context.Context, token string, tokenType string) (*domain.AuthToken, error) {
	var t domain.AuthToken
	// Quitamos el filtro de expires_at > NOW() para manejarlo en el usecase y poder debuguear mejor
	query := `SELECT * FROM auth_tokens WHERE token = $1 AND type = $2`
	err := r.db.GetContext(ctx, &t, query, token, tokenType)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *authTokenRepository) DeleteByUser(ctx context.Context, userID uint64, tokenType string) error {
	query := `DELETE FROM auth_tokens WHERE user_id = $1 AND type = $2`
	_, err := r.db.ExecContext(ctx, query, userID, tokenType)
	return err
}

func (r *authTokenRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM auth_tokens WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *authTokenRepository) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM auth_tokens WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
