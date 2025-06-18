package entity

import (
	"database/sql"
	"time"
)

type EmailVerificationCode struct {
	Id         sql.NullInt64 `db:"id" json:"id"`
	UserId     int64         `db:"user_id" json:"user_id"`
	Code       string        `db:"code" json:"code"`
	IsVerified bool          `db:"is_verified" json:"is_verified"`
	CreatedAt  time.Time     `db:"created_at" json:"created_at"`
	ExpiredAt  time.Time     `db:"expired_at" json:"expired_at"`
}
