package entity

import (
	"database/sql"
	"time"
)

type AccessToken struct {
	Id        sql.NullInt64 `db:"id" json:"id"`
	Token     string        `db:"token" json:"token"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	ExpiredAt time.Time     `db:"expired_at" json:"expired_at"`
}
