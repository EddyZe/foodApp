package entity

import "time"

type RefreshToken struct {
	Id        int64     `db:"id" json:"id"`
	UserId    int64     `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"token"`
	IssueAt   time.Time `db:"issue_at" json:"issue_at"`
	ExpiredAt time.Time `db:"expired_at" json:"expired_at"`
	IsRevoke  bool      `db:"is_revoke" json:"is_revoke"`
}
