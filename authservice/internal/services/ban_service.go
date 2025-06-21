package services

import (
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/sirupsen/logrus"
)

type BanService struct {
	log  *logrus.Entry
	repo *repositories.BanRepository
}

func NewBanService(log *logrus.Entry, repo *repositories.BanRepository) *BanService {
	return &BanService{
		log:  log,
		repo: repo,
	}
}

// GetActiveUserBan ищет блокировки пользователя
func (s *BanService) GetActiveUserBan(userId int64) (*entity.Ban, bool) {

	s.log.Debug("Поиск заблокированного пользователя в БД")
	ban, err := s.repo.FindActiveUserBans(userId)
	if err != nil {
		s.log.Debug("пользователь не найден с баном: ", err)
		return nil, false
	}

	return ban, true
}
