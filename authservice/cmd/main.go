package main

import (
	"log"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/server"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/sirupsen/logrus"
)

func main() {
	config.LoadEnv()
	logger := pkg.InitLogger("Auth-Service[MAIN] - ")

	logger.Infoln("Starting Auth Service...")
	//Загрузка конфигурации приложения
	appConf := config.Config(logger)

	//Подключение к БД
	postgre, red, err := connectionDbs(logger, appConf)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
		return
	}
	defer func() {
		if err := postgre.Close(); err != nil {
			logger.Errorf("Error closing connection to database: %v", err)
		}
	}()
	defer func() {
		if err := red.Close(); err != nil {
			logger.Errorf("Error closing connection to database: %v", err)
		}
	}()

	//Запуск сервера
	serv := server.New()
	if err := serv.ListenAndServe(); err != nil {
		log.Fatalf("Error starting Auth Service: %v", err)
	}
}

func connectionDbs(logger *logrus.Entry, appConf *config.AppConfig) (postgres *datasourse.PostgresDb, redis *datasourse.Redis, err error) {
	// подключение к постгес
	logger.Infoln("Connecting to database...")
	db, err := datasourse.ConnectionPostgres(appConf.Postgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
		return nil, nil, err
	}
	logger.Infoln("Database connected")

	//подключение к редис
	logger.Infoln("Connecting to redis")
	r, err := datasourse.ConnectionRedis(appConf.Redis)
	if err != nil {
		logger.Errorf("Error connecting to redis: %v", err)
		return nil, nil, err
	}
	logger.Infoln("Redis connected")
	return db, r, nil
}
