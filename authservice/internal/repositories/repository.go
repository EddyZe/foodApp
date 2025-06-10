package repositories

import "github.com/jmoiron/sqlx"

type Repository interface {
	CreateTx() (*sqlx.Tx, error)
	CommitTx(tx *sqlx.Tx) error
}

func commitTx(tx *sqlx.Tx) error {
	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return nil
}

func createTx(db *sqlx.DB) (*sqlx.Tx, error) {
	return db.Beginx()
}
