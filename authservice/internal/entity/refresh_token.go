package entity

type RefreshToken struct {
	Id        int64  `db:"id" json:"id"`
	UserId    int64  `db:"user_id" json:"user_id"`
	Token     string `db:"token" json:"token"`
	IssueAt   int64  `db:"issue_at" json:"issue_at"`
	ExpiredAt int64  `db:"expired_at" json:"expired_at"`
	IsRevoke  bool   `db:"is_revoke" json:"is_revoke"`
}
