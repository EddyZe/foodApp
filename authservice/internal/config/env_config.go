package config

import (
	"log"
	"os"

	"github.com/EddyZe/foodApp/common/config"
	"github.com/joho/godotenv"
)

var (
	cfg AppConfig
)

type AppConfig struct {
	Postgres          *PostgresConfig
	NewRelic          *NewRelic
	Sentry            *Sentry
	Redis             *RedisConfig
	Tokens            *TokenConfig
	SmptConfig        *SmptConfig
	EmailVerification *EmailVerificationCfg
	ResetPassword     *ResetPasswordVerificationCfg
	AppInfo           *AppInfo
	LocalizerConfig   *LocalizerConfig
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
	Host       string `env:"AUTH_REDIS_HOST" envDefault:"localhost"`
	Password   string `env:"AUTH_REDIS_PASSWORD" envDefault:""`
	Port       string `env:"AUTH_REDIS_PORT" envDefault:"6379"`
	DB         int    `env:"AUTH_REDIS_DB" envDefault:"0"`
	Expiration int    `env:"AUTH_REDIS_EXPIRATION" envDefault:"5"`
}

type TokenConfig struct {
	Secret                       string `env:"JWT_SECRET" envDefault:""`
	TokenExpirationMinute        int    `env:"TOKEN_EXPIRATION_MINUTES" envDefault:"15"`
	RefreshTokenExpirationMinute int    `env:"REFRESH_TOKEN_EXPIRATION_MINUTES" envDefault:"36000"`
}

type SmptConfig struct {
	Host     string `env:"SMTP_HOST" envDefault:"localhost"`
	Port     string `env:"SMTP_PORT" envDefault:"1025"`
	Username string `env:"SMTP_USERNAME" envDefault:""`
	Password string `env:"SMTP_PASSWORD" envDefault:""`
	From     string `env:"SMTP_FROM" envDefault:"test@test.com"`
}

type AppInfo struct {
	AppName     string `env:"APP_NAME" envDefault:"foodApp"`
	AppUrl      string `env:"APP_URL" envDefault:"http://localhost:8085"`
	SupportLink string `env:"APP_SUPPORT_LINK" envDefault:"support"`
}

type LocalizerConfig struct {
	DirFiles string `env:"AUTH_LOCALIZER_DIR_FILES" envDefault:"./locales"`
}

type EmailVerificationCfg struct {
	CodeExpiredMinute int `env:"EMAIL_VERIFICATION_CODE_EXPIRED_MINUTE" envDefault:"10"`
}

type ResetPasswordVerificationCfg struct {
	CodeExpiredMinute int `env:"RESET_PASSWORD_VERIFICATION_CODE_EXPIRED_MINUTE" envDefault:"10"`
}

func Config() *AppConfig {
	return config.LoadEnvConfig(&cfg)
}

func LoadEnv(envPath string) {
	if err := godotenv.Load(envPath); err != nil {
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
