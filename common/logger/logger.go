package logger

import (
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log logrus.Logger

func Init(service, logLevel, filepath string) *logrus.Entry {
	log := logrus.New()

	if logLevel == "" {
		logLevel = "info"
	}
	switch logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	appenv := os.Getenv("APP_ENV")
	if appenv == "" {
		appenv = "development"
	}

	if appenv == "production" {
		log.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "log.level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})
		log.SetOutput(
			&lumberjack.Logger{
				Filename:   filepath,
				MaxSize:    100,
				MaxBackups: 3,
				MaxAge:     28,
				Compress:   true,
			},
		)
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
		log.SetOutput(os.Stderr)
	}
	log.SetReportCaller(true)

	deploementId := os.Getenv("DEPLOYMENT_ID")
	if deploementId == "" {
		deploementId = "default"
	}

	return log.WithFields(logrus.Fields{
		"service":    service,
		"go_version": runtime.Version(),
		"deployment": deploementId,
		"env":        appenv,
	})
}
