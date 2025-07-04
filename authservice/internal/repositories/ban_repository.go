package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"time"

	"github.com/jmoiron/sqlx"
)

type BanRepository struct {
	*postgre.PostgresDb
}

func NewBanRepository(db *postgre.PostgresDb) *BanRepository {
	return &BanRepository{db}
}

func (r *BanRepository) Save(ban *entity.Ban) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.users_ban(user_id, cause, expired_at, is_forever) 
			values (:user_id, :cause, :expired_at, :is_forever)
			returning id, created_at`,
		ban,
	)

	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(
		ctx,
		query,
		args,
	).Scan(&ban.Id, &ban.CreatedAt); err != nil {
		return err
	}

	return nil
}

func (r *BanRepository) FindActiveUserBans(userId int64) (*entity.Ban, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var ban entity.Ban

	if err := r.GetContext(
		ctx,
		&ban,
		`select ub.* from auth.users_ban ub join auth.users u on ub.user_id = u.id
         where u.id = $1 and (now() < expired_at or is_forever = true)
         order by created_at desc`,
		userId,
	); err != nil {
		return nil, err
	}

	return &ban, nil
}

func (r *BanRepository) SetBanTx(ctx context.Context, tx *sqlx.Tx, ban *entity.Ban) error {
	query, args, err := tx.BindNamed(
		"insert into auth.users_ban (user_id, cause, expired_at, is_forever) values (:user_id, :cause, :expired_at, :is_forever) returning id, created_at",
		ban,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(
		ctx,
		query,
		args...,
	).Scan(&ban.Id, &ban.CreatedAt); err != nil {
		return err
	}
	return nil
}

func (r *BanRepository) DeleteByUserId(userId int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`delete from auth.users_ban where user_id = $1`,
		userId,
	).Err(); err != nil {
		return err
	}

	return nil
}

func (r *BanRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *BanRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
