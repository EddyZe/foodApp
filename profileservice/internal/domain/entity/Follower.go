package entity

import "time"

type Follower struct {
	ProfileId  int64     `db:"profile_id" json:"profile_id"`
	FollowerId int64     `db:"follower_id" json:"follower_id"`
	FollowedAt time.Time `db:"followed_at" json:"followed_at"`
}
