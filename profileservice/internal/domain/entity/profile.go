package entity

import "time"

type Profile struct {
	Id        int64     `db:"id" json:"id"`
	UserId    int64     `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Bio       string    `db:"bio" json:"bio"`
}
