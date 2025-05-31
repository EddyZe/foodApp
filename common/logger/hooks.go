package logger

import (
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

type SentryHook struct{}

func (h *SentryHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
}

func (h *SentryHook) Fire(entry *logrus.Entry) error {
	event := sentry.NewEvent()
	event.Message = entry.Message
	event.Level = sentry.Level(entry.Level.String())
	event.Extra = entry.Data

	sentry.CaptureEvent(event)
	return nil
}

// AddSentryHook добавляет интеграцию с Sentry
func AddSentryHook(dsn string) {
	if dsn == "" {
		return
	}

	err := sentry.Init(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		log.Errorf("Sentry init failed: %v", err)
		return
	}

	log.AddHook(&SentryHook{})
}
