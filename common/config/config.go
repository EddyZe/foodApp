package config

import (
	"context"
	"sync"
	"time"

	"github.com/sethvargo/go-envconfig"
)

var configOnce sync.Once

func LoadEnvConfig[T any](cfg T) T {
	configOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()
		if err := envconfig.Process(ctx, cfg); err != nil {
			panic(err)
		}
	})
	return cfg
}
