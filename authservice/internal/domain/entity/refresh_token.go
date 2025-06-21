package entity

import (
	"database/sql"
	"time"
)

type RefreshToken struct {
	Id            sql.NullInt64 `db:"id" json:"id"`
	UserId        int64         `db:"user_id" json:"user_id"`
	AccessTokenId sql.NullInt64 `db:"access_token_id" json:"access_token_id"`
	Token         string        `db:"token" json:"token"`
	IssueAt       time.Time     `db:"issue_at" json:"issue_at"`
	ExpiredAt     time.Time     `db:"expired_at" json:"expired_at"`
	IsRevoke      bool          `db:"is_revoke" json:"is_revoke"`
}
