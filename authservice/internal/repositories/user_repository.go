package repositories

import (
	"context"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	*datasourse.PostgresDb
}

func NewUserRepository(db *datasourse.PostgresDb) *UserRepository {
	return &UserRepository{
		db,
	}
}

func (r *UserRepository) Save(u *entity.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.users (email, password) values (:email, :password) returning id`,
		u,
	)
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).Scan(&u.Id); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) SaveTx(ctx context.Context, tx *sqlx.Tx, u *entity.User) error {
	query, args, err := tx.BindNamed(
		`insert into auth.users (email, password) values (:email, :password) returning id`,
		u,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(ctx, query, args...).Scan(&u.Id); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) FindByEmail(email string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var u entity.User
	if err := r.GetContext(
		ctx,
		&u,
		`select * from auth.users where email = $1`,
		email,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByEmailTx(ctx context.Context, tx *sqlx.Tx, email string) (*entity.User, error) {
	var u entity.User

	if err := tx.GetContext(
		ctx,
		&u,
		`select * from auth.users where email = $1`,
		email,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *UserRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
