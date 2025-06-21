package entity

import "database/sql"

type Role struct {
	Id          sql.NullInt64 `db:"id" json:"id"`
	Name        string        `db:"name" json:"name"`
	Description string        `db:"description" json:"description"`
}
