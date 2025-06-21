package entity

import (
	"database/sql"
	"time"
)

type EmailVerificationToken struct {
	Id            int64         `db:"id" json:"id"`
	Token         string        `db:"token" json:"token"`
	AccessTokenId sql.NullInt64 `db:"access_token_id" json:"access_token_id"`
	CodeId        int64         `db:"code_id" json:"code_id"`
	IsActive      bool          `db:"is_active" json:"is_active"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
	ExpiredAt     time.Time     `db:"expired_at" json:"expired_at"`
}

func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().Unix() > t.ExpiredAt.Unix()
}
