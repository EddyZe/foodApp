package entity

import "database/sql"

type UserRole struct {
	UserId sql.NullInt64 `db:"user_id" json:"user_id"`
	RoleId sql.NullInt64 `db:"role_id" json:"role_id"`
}
