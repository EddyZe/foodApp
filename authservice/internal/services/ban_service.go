package services

import (
	"encoding/json"
	"fmt"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/common/util/redisutil"
	"github.com/sirupsen/logrus"
)

var banUserId = "ban:user:id"

type BanService struct {
	log   *logrus.Entry
	redis *datasourse.Redis
	repo  *repositories.BanRepository
}

func NewBanService(log *logrus.Entry, redis *datasourse.Redis, repo *repositories.BanRepository) *BanService {
	return &BanService{
		log:   log,
		repo:  repo,
		redis: redis,
	}
}

// GetActiveUserBan ищет блокировки пользователя
func (s *BanService) GetActiveUserBan(userId int64) (*entity.Ban, bool) {
	redisKey := redisutil.GenerateKey(banUserId, fmt.Sprint(userId))

	s.log.Debug("Поиск в редисе заблокированного пользователя")
	if data, ok := s.redis.Get(redisKey); ok {
		var ban entity.Ban
		if err := json.Unmarshal([]byte(data), &ban); err != nil {
			s.log.Error("ошибка при генерации модели из json: ", err)
		} else {
			return &ban, true
		}
	}

	s.log.Debug("Поиск заблокированного пользователя в БД")
	ban, err := s.repo.FindActiveUserBans(userId)
	if err != nil {
		s.log.Error("Ошибка при поиске пользователя в базе: ", err)
		return nil, false
	}

	s.log.Debug("Сохранение в редис заблокированного пользователя")
	if err := s.redis.Put(redisKey, ban); err != nil {
		s.log.Error("ошибка при записи в redis: ", err)
	}

	return ban, true
}
