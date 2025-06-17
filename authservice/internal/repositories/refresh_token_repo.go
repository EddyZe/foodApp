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
		`insert into auth.refresh_token(user_id, token, expired_at, access_token_id) 
				values (:user_id, :token, :expired_at, :access_token_id)
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
		`insert into auth.refresh_token(user_id, token, expired_at, access_token_id) 
				values (:user_id, :token, :expired_at, :access_token_id)
				returning id`,
		token,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(ctx, query, args...).Scan(&token.Id); err != nil {
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

func (r *RefreshTokenRepository) Update(token *entity.RefreshToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`update auth.refresh_token set is_revoke = :is_revoke where token = :token`,
		token,
	)
	if err != nil {
		return err
	}
	if err := r.QueryRowxContext(ctx, query, args...).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RefreshTokenRepository) RemoveAllRefreshTokensUser(userId int64) ([]entity.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	rows, err := r.QueryxContext(
		ctx,
		"DELETE FROM auth.refresh_token rt WHERE user_id = $1 RETURNING rt.*",
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var tokens []entity.RefreshToken
	for rows.Next() {
		var token entity.RefreshToken
		if err := rows.StructScan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

func (r *RefreshTokenRepository) RemoveAllRefreshTokensUserTx(ctx context.Context, tx *sqlx.Tx, userId int64) ([]entity.RefreshToken, error) {
	rows, err := tx.QueryxContext(
		ctx,
		"DELETE FROM auth.refresh_token rt WHERE user_id = $1 RETURNING rt.*",
		userId,
	)
	if err != nil {
		return nil, err
	}
	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var tokens []entity.RefreshToken
	for rows.Next() {
		var token entity.RefreshToken
		if err := rows.StructScan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

func (r *RefreshTokenRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *RefreshTokenRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
