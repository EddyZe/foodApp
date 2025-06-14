package entity

import (
	"database/sql"
	"time"
)

type Ban struct {
	Id        sql.NullInt64 `db:"id" json:"-"`
	UserId    sql.NullInt64 `db:"user_id" json:"-"`
	IsForever bool          `db:"is_forever" json:"is_forever"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
	ExpiredAt time.Time     `db:"expired_at" json:"expired_at"`
	Cause     string        `db:"cause" json:"cause"`
}
