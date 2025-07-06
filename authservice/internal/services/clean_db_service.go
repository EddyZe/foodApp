package services

import (
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/sirupsen/logrus"
)

type CleanDBService struct {
	log  *logrus.Entry
	repo *repositories.ClearDBRepository
}

func NewCleanDBService(log *logrus.Entry, repo *repositories.ClearDBRepository) *CleanDBService {
	return &CleanDBService{
		repo: repo,
		log:  log,
	}
}

func (s *CleanDBService) Clean() error {
	s.log.Debug("Чистка просроченых данных в базе")
	if err := s.repo.ClearDB(); err != nil {
		s.log.Error("Ошибка при попытки очистить просроченые данные из базы: ", err)
		return err
	}
	s.log.Debug("Очистка завершена")
	return nil
}
