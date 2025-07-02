package services

import (
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/sirupsen/logrus"
)

type HistoryPasswordService struct {
	log  *logrus.Entry
	repo *repositories.PasswordHistoryRepository
}

func NewHistoryPasswordService(log *logrus.Entry, repo *repositories.PasswordHistoryRepository) *HistoryPasswordService {
	return &HistoryPasswordService{
		log:  log,
		repo: repo,
	}
}

func (s *HistoryPasswordService) Save(userId int64, oldPassword string) error {
	oldPass := entity.PasswordHistory{
		UserId:      userId,
		OldPassword: oldPassword,
	}
	s.log.Debugf("Сохранение старого пароля")

	return s.repo.Save(&oldPass)
}

func (s *HistoryPasswordService) GetLastPasswords(userId int64, limit int) []entity.PasswordHistory {
	s.log.Debug("Получение последних паролей пользователя с ID: ", userId)
	return s.repo.GetLastPasswords(userId, limit)
}

func (s *HistoryPasswordService) GetLastPassword(userId int64) *entity.PasswordHistory {
	s.log.Debug("Поиск последнего старого пароля пользователя с ID: ", userId)
	passwords := s.GetLastPasswords(userId, 0)
	if len(passwords) == 0 {
		s.log.Debug("Нет старых паролей")
		return &entity.PasswordHistory{}
	}

	return &passwords[0]
}
