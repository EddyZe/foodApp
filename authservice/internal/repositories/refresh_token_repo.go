package repositories

import (
	"context"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/jmoiron/sqlx"
)

type RefreshTokenRepository struct {
	*datasourse.PostgresDb
}

func NewRefreshTokenRepository(db *datasourse.PostgresDb) *RefreshTokenRepository {
	return &RefreshTokenRepository{db}
}

func (r *RefreshTokenRepository) Save(token *entity.RefreshToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.refresh_token(user_id, token, expired_at) 
				values (:user_id, :token, :expired_at)
				returning id`,
		token,
	)
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).Scan(&token.Id); err != nil {
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) SaveTx(ctx context.Context, tx *sqlx.Tx, token *entity.RefreshToken) error {
	query, args, err := tx.BindNamed(
		`insert into auth.refresh_token(user_id, token, expired_at) 
				values (:user_id, :token, :expired_at)
				returning id`,
		token,
	)
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).Scan(&token.Id); err != nil {
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) FindByToken(token string) (*entity.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var res entity.RefreshToken

	if err := r.GetContext(
		ctx,
		&res,
		`select * from auth.refresh_token where token = $1`,
		token,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *RefreshTokenRepository) FindByTokenTx(ctx context.Context, tx *sqlx.Tx, token string) (*entity.RefreshToken, error) {
	var res entity.RefreshToken

	if err := tx.GetContext(
		ctx,
		&res,
		`select * from auth.refresh_token where token = $1`,
		token,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *RefreshTokenRepository) DeleteByToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`delete from auth.refresh_token where token = $1`,
		token,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteByTokenTx(ctx context.Context, tx *sqlx.Tx, token string) error {
	if err := tx.QueryRowxContext(
		ctx,
		`delete from auth.refresh_token where token = $1`,
		token,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeToken(token string, isRevoked bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`update auth.refresh_token set is_revoke=$1 where token=$2`,
		isRevoked,
		token,
	).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokenRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *RefreshTokenRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
