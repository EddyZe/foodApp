package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"time"
)

type ResetPasswordRepository struct {
	*postgre.PostgresDb
}

func NewResetPasswordRepository(db *postgre.PostgresDb) *ResetPasswordRepository {
	return &ResetPasswordRepository{db}
}

func (r *ResetPasswordRepository) Save(reset *entity.ResetPasswordCode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.reset_password_codes (user_id, code, expired_at) 
			values (:user_id, :code, :expired_at) returning id`,
		reset,
	)

	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(
		ctx,
		query,
		args...,
	).Scan(&reset.Id); err != nil {
		return err
	}

	return nil
}

func (r *ResetPasswordRepository) SaveTx(ctx context.Context, tx *sqlx.Tx, reset *entity.ResetPasswordCode) error {
	query, args, err := tx.BindNamed(
		`insert into auth.reset_password_codes (user_id, code, expired_at) 
		values (:user_id, :code, :expired_at) returning id`,
		reset,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(
		ctx,
		query,
		args...,
	).Scan(&reset.Id); err != nil {
		return err
	}

	return nil
}

func (r *ResetPasswordRepository) FindByCode(code string) (*entity.ResetPasswordCode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var res entity.ResetPasswordCode

	if err := r.GetContext(
		ctx,
		&res,
		`select * from auth.reset_password_codes where code = $1`,
		code,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *ResetPasswordRepository) DeleteByCode(code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`delete from auth.reset_password_codes where code = $1`,
		code,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *ResetPasswordRepository) DeleteByCodeTx(ctx context.Context, tx *sqlx.Tx, code string) error {
	if err := tx.QueryRowxContext(
		ctx,
		`delete from auth.reset_password_codes where code = $1`,
		code,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *ResetPasswordRepository) SetIsValid(code string, b bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`update auth.reset_password_codes set is_valid=$1 where code = $2`,
		b,
		code,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *ResetPasswordRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *ResetPasswordRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
