package entity

import "time"

type PasswordHistory struct {
	Id          int64     `db:"id" json:"id"`
	UserId      int64     `db:"user_id" json:"user_id"`
	OldPassword string    `db:"old_password" json:"old_password"`
	ChangedAt   time.Time `db:"changed_at" json:"changed_at"`
}
