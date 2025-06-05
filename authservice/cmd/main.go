package main

import (
	"log"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/server"
	"github.com/EddyZe/foodApp/authservice/pkg"
)

func main() {
	config.LoadEnv()
	logger := pkg.InitLogger("Auth-Service[MAIN] - ")

	logger.Infoln("Starting Auth Service...")
	//Загрузка конфигурации приложения
	appConf := config.Config(logger)

	//Подключение к БД
	logger.Infoln("Connecting to database...")
	db, err := datasourse.ConnectionPostgres(appConf.Postgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
		return
	}
	defer func() {
		err := db.Close()
		if err != nil {
			logger.Errorf("Error closing database connection: %v", err)
		}
	}()
	logger.Infoln("Database connected")

	//Запуск сервера
	serv := server.New()
	if err := serv.ListenAndServe(); err != nil {
		log.Fatalf("Error starting Auth Service: %v", err)
	}
}
