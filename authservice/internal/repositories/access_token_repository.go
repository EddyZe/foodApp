package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"time"
)

type AccessTokenRepository struct {
	*postgre.PostgresDb
}

func NewAccessTokenRepository(db *postgre.PostgresDb) *AccessTokenRepository {
	return &AccessTokenRepository{db}
}

func (r *AccessTokenRepository) SaveTx(ctx context.Context, tx *sqlx.Tx, token *entity.AccessToken) error {
	query, args, err := tx.BindNamed(
		`insert into auth.access_token(token, expired_at) values (:token, :expired_at) returning id`,
		token,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(
		ctx,
		query,
		args...,
	).Scan(&token.Id); err != nil {
		return err
	}

	return nil
}

func (r *AccessTokenRepository) DeleteByRefreshTokenByTokenTx(ctx context.Context, tx *sqlx.Tx, refreshToken string) error {
	if _, err := tx.ExecContext(
		ctx,
		`delete from auth.access_token 
       where id in (select ac.id from auth.access_token ac 
           join auth.refresh_token rt on rt.access_token_id=ac.id where rt.token=$1)`,
		refreshToken,
	); err != nil {
		return err
	}

	return nil
}

func (r *AccessTokenRepository) DeleteByIdsTx(ctx context.Context, tx *sqlx.Tx, ids ...int64) error {
	query, args, err := sqlx.In(`delete from auth.access_token where id in (?)`, ids)
	if err != nil {
		return err
	}

	query = tx.Rebind(query)

	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (r *AccessTokenRepository) DeleteByTokenTx(ctx context.Context, tx *sqlx.Tx, token string) error {
	if err := tx.QueryRowxContext(
		ctx,
		`delete from auth.access_token where token=$1`,
		token,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *AccessTokenRepository) FindByToken(token string) (*entity.AccessToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var res entity.AccessToken

	if err := r.GetContext(
		ctx,
		&res,
		"select * from auth.access_token where token = $1",
		token,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *AccessTokenRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *AccessTokenRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
