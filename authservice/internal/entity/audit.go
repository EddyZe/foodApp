package entity

import (
	"database/sql"
	"time"
)

type Audit struct {
	Id        sql.NullInt64 `db:"id" json:"id"`
	UserId    sql.NullInt64 `db:"user_id" json:"userId"`
	Action    string        `db:"action" json:"action"`
	IpAddress string        `db:"ip_address" json:"ip_address"`
	UserAgent string        `db:"user_agent" json:"user_agent"`
	CreatedAt time.Time     `db:"created_at" json:"created_at"`
}
