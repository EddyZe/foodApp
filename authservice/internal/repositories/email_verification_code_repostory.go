package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/jmoiron/sqlx"
	"time"
)

type EmailVerificationCodeRepository struct {
	*datasourse.PostgresDb
}

func NewEmailVerificationCodeRepository(db *datasourse.PostgresDb) *EmailVerificationCodeRepository {
	return &EmailVerificationCodeRepository{db}
}

func (r *EmailVerificationCodeRepository) Save(code *entity.EmailVerificationCode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.email_verification_codes (code, user_id, expired_at) 
		values (:code, :user_id, :expired_at)
		returning id;`,
		code,
	)
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).Scan(&code.Id); err != nil {
		return err
	}

	return nil
}

func (r *EmailVerificationCodeRepository) SaveTx(ctx context.Context, tx *sqlx.Tx, code *entity.EmailVerificationCode) error {
	query, args, err := tx.BindNamed(
		`insert into auth.email_verification_codes (code, user_id, expired_at) 
		values (:code, :user_id, :expired_at)
		returning id;`,
		code,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(ctx, query, args...).Scan(&code.Id); err != nil {
		return err
	}

	return nil
}

func (r *EmailVerificationCodeRepository) FindByCode(codeString string) (*entity.EmailVerificationCode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var code entity.EmailVerificationCode

	if err := r.GetContext(
		ctx,
		&code,
		`select * from auth.email_verification_codes where code=$1`,
		codeString,
	); err != nil {
		return nil, err
	}

	return &code, nil
}

func (r *EmailVerificationCodeRepository) FindByCodeTx(ctx context.Context, tx *sqlx.Tx, codeString string) (*entity.EmailVerificationCode, error) {

	var code entity.EmailVerificationCode

	if err := tx.GetContext(
		ctx,
		&code,
		`select * from auth.email_verification_codes where code=$1`,
		codeString,
	); err != nil {
		return nil, err
	}

	return &code, nil
}

func (r *EmailVerificationCodeRepository) DeleteByCode(codeString string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`delete from auth.email_verification_codes where code = $1`,
		codeString,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *EmailVerificationCodeRepository) SetVerified(codeString string, b bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`update auth.email_verification_codes set is_verified=$1 where code=$2`,
		b,
		codeString,
	).Err(); err != nil {
		return err
	}
	return nil
}

func (r *EmailVerificationCodeRepository) FindCodeByVerifiedToken(token string) (*entity.EmailVerificationCode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var res entity.EmailVerificationCode

	if err := r.GetContext(
		ctx,
		&res,
		`select c.* 
	from auth.email_verification_codes c
    join auth.email_verification_token t
        on t.code_id=c.id 
	where t.token = $1`,
		token,
	); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *EmailVerificationCodeRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *EmailVerificationCodeRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
