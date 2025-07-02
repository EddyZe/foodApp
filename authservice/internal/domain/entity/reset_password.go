package entity

import "time"

type ResetPasswordCode struct {
	Id        int64     `db:"id" json:"id"`
	UserId    int64     `db:"user_id" json:"user_id"`
	Code      string    `db:"code" json:"code"`
	IsValid   bool      `db:"is_valid" json:"is_valid"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiredAt time.Time `db:"expired_at" json:"expired_at"`
}

func (e *ResetPasswordCode) IsExpired() bool {
	return e.ExpiredAt.Unix() < time.Now().Unix()
}
