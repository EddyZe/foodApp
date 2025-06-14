package repositories

import (
	"context"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/jmoiron/sqlx"
)

type BanRepository struct {
	*datasourse.PostgresDb
}

func NewBanRepository(db *datasourse.PostgresDb) *BanRepository {
	return &BanRepository{db}
}

func (r *BanRepository) Save(ban *entity.Ban) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.users_ban(user_id, cause, expired_at, is_forever) 
			values (:user_id, :cause, :expired_at, :is_forever)
			returning id`,
		ban,
	)

	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(
		ctx,
		query,
		args,
	).Scan(&ban.Id); err != nil {
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

func (r *BanRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *BanRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
