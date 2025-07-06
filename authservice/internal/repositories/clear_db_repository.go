package repositories

import (
	"context"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"time"
)

type ClearDBRepository struct {
	*postgre.PostgresDb
}

func NewClearDBRepository(db *postgre.PostgresDb) *ClearDBRepository {
	return &ClearDBRepository{
		db,
	}
}

func (r *ClearDBRepository) ClearDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := r.QueryRowxContext(
		ctx,
		`select auth.clean_expired_data()`,
	).Err(); err != nil {
		return err
	}

	return nil
}
