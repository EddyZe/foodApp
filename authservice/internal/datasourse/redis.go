package datasourse

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/go-redis/redis"
)

type Redis struct {
	expiration time.Duration
	cli        *redis.Client
}

func NewRedis(cfg *config.RedisConfig) *Redis {
	return &Redis{
		expiration: time.Duration(cfg.Expiration) * time.Minute,
		cli: redis.NewClient(&redis.Options{
			Addr:       fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password:   cfg.Password,
			DB:         cfg.DB,
			MaxRetries: 5,
		}),
	}
}

func (r *Redis) Put(key string, value interface{}) error {
	cli := r.cli
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	res := cli.Set(key, string(jsonValue), r.expiration)
	return res.Err()
}

func (r *Redis) Get(key string) (string, error) {
	cli := r.cli
	res := cli.Get(key)
	return res.Val(), res.Err()
}

func (r *Redis) Del(key string) error {
	cli := r.cli
	res := cli.Del(key)
	return res.Err()
}
