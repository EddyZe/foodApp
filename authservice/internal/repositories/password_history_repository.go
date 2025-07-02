package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"time"
)

type PasswordHistoryRepository struct {
	*postgre.PostgresDb
}

func NewPasswordHistoryRepository(db *postgre.PostgresDb) *PasswordHistoryRepository {
	return &PasswordHistoryRepository{db}
}

func (r *PasswordHistoryRepository) Save(pass *entity.PasswordHistory) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.password_history (user_id, old_password) 
			values (:user_id, :old_password) returning id`, pass,
	)
	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(
		ctx,
		query,
		args...,
	).Scan(&pass.Id); err != nil {
		return err
	}

	return nil
}

func (r *PasswordHistoryRepository) GetLastPasswords(userId int64, limit int) []entity.PasswordHistory {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res := make([]entity.PasswordHistory, 0)

	if err := r.SelectContext(
		ctx,
		&res,
		`select * from auth.password_history where user_id = $1 order by changed_at desc limit $2`,
		userId,
		limit,
	); err != nil {
		return res
	}

	return res
}

func (r *PasswordHistoryRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *PasswordHistoryRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
