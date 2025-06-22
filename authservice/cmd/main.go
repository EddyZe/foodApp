package main

import (
	"github.com/EddyZe/foodApp/authservice/internal/app/authapp"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	log "github.com/EddyZe/foodApp/common/pkg/logger"
	"github.com/sirupsen/logrus"
)

func main() {

	//загрузка конфигурации из yaml
	cfg := config.MustLoad()
	logger := setLogger(cfg.LoggerCfg)

	//Загрузка конфигурации приложения
	config.LoadEnv(cfg.EnvFile)

	appConf := config.Config()

	authapp.MustRun(logger, appConf)

}

func setLogger(loggerCfg *config.LoggerCfg) *logrus.Entry {
	logger := log.Init(loggerCfg.AppEnv, loggerCfg.Service, loggerCfg.Level, loggerCfg.FilePath)
	return logger
}
