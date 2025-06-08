package config

import (
	"context"
	"sync"
	"time"

	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
)

var configOnce sync.Once

func LoadEnvConfig[T any](log *logrus.Entry, cfg T) T {
	configOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()
		if err := envconfig.Process(ctx, cfg); err != nil {
			log.Fatal(err)
		}
		log.Infoln("Environments initialized.")
	})
	return cfg
}
