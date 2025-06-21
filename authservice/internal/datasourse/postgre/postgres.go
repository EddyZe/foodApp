package postgre

import (
	"fmt"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDb struct {
	*sqlx.DB
}

func Connect(config *config.PostgresConfig) (*PostgresDb, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&client_encoding=%s",
		config.User,
		config.Pass,
		config.Host,
		config.Port,
		config.DbName,
		"UTF8",
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresDb{db}, nil
}

func (db *PostgresDb) Close() error {
	return db.DB.Close()
}

func (db *PostgresDb) Ping() error {
	return db.DB.Ping()
}
