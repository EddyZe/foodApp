package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"time"

	"github.com/jmoiron/sqlx"
)

type UserRoleRepository struct {
	*postgre.PostgresDb
}

func NewUserRoleRepository(db *postgre.PostgresDb) *UserRoleRepository {
	return &UserRoleRepository{db}
}

func (r *UserRoleRepository) SetUserRole(role entity.UserRole) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.user_roles (user_id, role_id) values (:user_id, :role_id)`,
		role,
	)

	if err != nil {
		return err
	}
	if err := r.QueryRowxContext(ctx, query, args...).Err(); err != nil {
		return err
	}

	return nil
}

func (r *UserRoleRepository) SetUserRoleTx(ctx context.Context, tx *sqlx.Tx, role *entity.UserRole) error {
	query, atgs, err := tx.BindNamed(
		`insert into auth.user_roles (user_id, role_id) values (:user_id, :role_id)`,
		role,
	)
	if err != nil {
		return err
	}

	if err := tx.QueryRowxContext(ctx, query, atgs...).Err(); err != nil {
		return err
	}

	return nil
}

func (r *UserRoleRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *UserRoleRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
