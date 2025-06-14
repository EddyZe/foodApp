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

func ConnectionRedis(cfg *config.RedisConfig) (*Redis, error) {
	r := &Redis{
		expiration: time.Duration(cfg.Expiration) * time.Minute,
		cli: redis.NewClient(&redis.Options{
			Addr:       fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password:   cfg.Password,
			DB:         cfg.DB,
			MaxRetries: 5,
		}),
	}
	if err := r.cli.Ping().Err(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Redis) Put(key string, value interface{}) error {
	return r.PutEx(key, value, r.expiration)
}

func (r *Redis) PutEx(key string, value interface{}, expiration time.Duration) error {
	cli := r.cli
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	res := cli.Set(key, string(jsonValue), expiration)
	return res.Err()
}

func (r *Redis) Get(key string) (string, bool) {
	cli := r.cli
	res := cli.Get(key)
	if res.Err() != nil {
		return "", false
	}
	return res.Val(), true
}

func (r *Redis) Del(key string) error {
	cli := r.cli
	res := cli.Del(key)
	return res.Err()
}

func (r *Redis) Close() error {
	return r.cli.Close()
}
