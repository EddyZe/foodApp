package pkg

import (
	"os"

	"github.com/EddyZe/foodApp/common/logger"
	"github.com/sirupsen/logrus"
)

func InitLogger() *logrus.Entry {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	return logger.Init("auth-service", logLevel, "./logs/authservice.log")
}
