package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"time"

	"github.com/jmoiron/sqlx"
)

type AuditRepository struct {
	*postgre.PostgresDb
}

func NewAuditRepository(db *postgre.PostgresDb) *AuditRepository {
	return &AuditRepository{db}
}

func (r *AuditRepository) Save(audit *entity.Audit) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args, err := r.BindNamed(
		`insert into auth.audit_log(user_id, action, ip_address, user_agent)
			VALUES (:user_id, :action, :ip_address, :user_agent)`,
		audit,
	)

	if err != nil {
		return err
	}

	if err := r.QueryRowxContext(ctx, query, args...).Err(); err != nil {
		return err
	}
	return nil
}

func (r *AuditRepository) SaveTx(ctx context.Context, tx *sqlx.Tx, audit *entity.Audit) error {
	query, args, err := tx.BindNamed(
		`insert into auth.audit_log(user_id, action, ip_address, user_agent)
			VALUES (:user_id, :action, :ip_address, :user_agent)`,
		audit,
	)

	if err != nil {
		return err
	}
	if err := r.QueryRowxContext(ctx, query, args...).Err(); err != nil {
		return err
	}

	return nil
}

func (r *AuditRepository) CreateTx() (*sqlx.Tx, error) {
	return createTx(r.DB)
}

func (r *AuditRepository) CommitTx(tx *sqlx.Tx) error {
	return commitTx(tx)
}
