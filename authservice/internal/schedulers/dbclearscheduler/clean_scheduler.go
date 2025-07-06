package dbclearscheduler

import (
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type CleanDbScheduler struct {
	log *logrus.Entry
	cs  *services.CleanDBService
}

func NewCleanDBScheduler(log *logrus.Entry, cs *services.CleanDBService) *CleanDbScheduler {
	return &CleanDbScheduler{
		log: log,
		cs:  cs,
	}
}

func (s *CleanDbScheduler) Start() error {
	c := cron.New()

	if _, err := c.AddFunc("*/15 * * * *", func() {
		if err := s.cs.Clean(); err != nil {
			s.log.Error(err)
		}
	}); err != nil {
		s.log.Error("Ошибка при запуске шедулера: ", err)
		return err
	}
	c.Start()
	return nil
}
