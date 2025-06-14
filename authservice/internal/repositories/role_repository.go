package repositories

import (
	"context"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	*datasourse.PostgresDb
}

func NewRoleRepository(db *datasourse.PostgresDb) *RoleRepository {
	return &RoleRepository{db}
}

func (r *RoleRepository) FindByName(name string) (*entity.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var res *entity.Role

	if err := r.GetContext(
		ctx,
		res,
		`select * from auth.role where name=$1`,
		name); err != nil {
		return nil, err
	}
	return res, nil
}

func (r *RoleRepository) FindByNameTx(ctx context.Context, tx *sqlx.Tx, name string) (*entity.Role, error) {
	var res entity.Role
	if err := tx.GetContext(
		ctx,
		&res,
		`select * from auth.role where name =$1`,
		name); err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *RoleRepository) GetRoleByUserId(userId int64) []entity.Role {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var res []entity.Role

	if err := r.SelectContext(
		ctx,
		&res,
		`select r.* from auth.role r join auth.user_roles ur on r.id = ur.role_id
			where ur.user_id=$1`,
		userId,
	); err != nil {
		return make([]entity.Role, 0)
	}

	return res
}

func (r *RoleRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *RoleRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
