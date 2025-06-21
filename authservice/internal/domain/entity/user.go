package entity

import (
	"database/sql"
	"time"
)

type User struct {
	Id             sql.NullInt64 `db:"id" json:"id"`
	Email          string        `db:"email" json:"email"`
	Password       string        `db:"password" json:"password"`
	EmailIsConfirm bool          `db:"email_is_confirm" json:"email_is_confirm"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at"`
}
