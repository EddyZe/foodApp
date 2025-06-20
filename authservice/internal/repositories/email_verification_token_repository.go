package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/jmoiron/sqlx"
	"time"
)

type EmailVerificationTokenRepository struct {
	*datasourse.PostgresDb
}

func NewEmailVerificationTokenRepository(db *datasourse.PostgresDb) *EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepository{db}
}

func (r *EmailVerificationTokenRepository) Save(tok *entity.EmailVerificationToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		"insert into auth.email_verification_token(code_id, token, expired_at) values (:code_id, :token, :expired_at) returning id",
		tok,
	)
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).Scan(&tok.Id); err != nil {
		return err
	}

	return nil
}

func (r *EmailVerificationTokenRepository) SetIsActive(token string, b bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		"update auth.email_verification_token set is_active = $1 where token = $2",
		b,
		token,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *EmailVerificationTokenRepository) FindToken(token string) (*entity.EmailVerificationToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	var res entity.EmailVerificationToken

	if err := r.GetContext(
		ctx,
		&res,
		"select * from auth.email_verification_token where token=$1",
		token,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *EmailVerificationTokenRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *EmailVerificationTokenRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
