package config

import (
	"log"
	"os"

	"github.com/EddyZe/foodApp/common/config"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	cfg AppConfig
)

type AppConfig struct {
	Postgres *PostgresConfig
	NewRelic *NewRelic
	Sentry   *Sentry
	Redis    *RedisConfig
}

type NewRelic struct {
	AppName string `env:"AUTH_APP_NAME, default=go-app"`
	License string `env:"AUTH_NEW_RELIC_LICENSE"`
}

type Sentry struct {
	Dsn string `env:"AUTH_SENTRY_DSN"`
}

type PostgresConfig struct {
	DbName string `env:"AUTH_DB_NAME" envDefault:"postgres"`
	Host   string `env:"AUTH_DB_HOST" envDefault:"localhost"`
	Port   string `env:"AUTH_DB_PORT" envDefault:"5432"`
	User   string `env:"AUTH_DB_USER" envDefault:"postgres"`
	Pass   string `env:"AUTH_DB_PASSWORD" envDefault:"postgres"`
}

type RedisConfig struct {
	Host       string `env:"REDIS_HOST" envDefault:"localhost"`
	Password   string `env:"REDIS_PASSWORD" envDefault:""`
	Port       string `env:"REDIS_PORT" envDefault:"6379"`
	DB         int    `env:"REDIS_DB" envDefault:"0"`
	Expiration int    `env:"REDIS_EXPIRATION" envDefault:"5"`
}

func Config(log *logrus.Entry) *AppConfig {
	return config.LoadEnvConfig(log, &cfg)
}

func LoadEnv() {
	if err := godotenv.Load("./../.env"); err != nil {
		log.Println("Error loading .env file")
	}
}

func GetPort() string {
	port := os.Getenv("AUTH_SERVER_PORT")
	if port == "" {
		port = "8081"
	}
	return port
}
