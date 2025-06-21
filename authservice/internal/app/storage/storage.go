package storage

import (
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/postgre"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/redis"
	"github.com/sirupsen/logrus"
)

func MustConnectPsql(logger *logrus.Entry, cfg *config.PostgresConfig) *postgre.PostgresDb {
	psql, err := ConnectPsql(logger, cfg)
	if err != nil {
		panic(err)
	}
	return psql
}

func ConnectPsql(logger *logrus.Entry, cfg *config.PostgresConfig) (*postgre.PostgresDb, error) {
	// подключение к постгес
	logger.Infoln("Connecting to database...")
	db, err := postgre.Connect(cfg)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
		return nil, err
	}
	logger.Infoln("Database connected")

	return db, nil
}

func MustConnectRedis(logger *logrus.Entry, cfg *config.RedisConfig) *redis.Redis {
	r, err := ConnectRedis(logger, cfg)
	if err != nil {
		panic(err)
	}

	return r
}

func ConnectRedis(logger *logrus.Entry, cfg *config.RedisConfig) (*redis.Redis, error) {
	logger.Infoln("Connecting to redis")
	r, err := redis.Connect(cfg)
	if err != nil {
		logger.Errorf("Error connecting to redis: %v", err)
		return nil, err
	}
	logger.Infoln("Redis connected")
	return r, nil
}
