package services

import (
	"context"
	"database/sql"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/sirupsen/logrus"
	"time"
)

type BanService struct {
	log  *logrus.Entry
	repo *repositories.BanRepository
	ts   *TokenService
}

func NewBanService(log *logrus.Entry, repo *repositories.BanRepository, ts *TokenService) *BanService {
	return &BanService{
		log:  log,
		repo: repo,
		ts:   ts,
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

func (s *BanService) BanUser(userId int64, cause string, expiredAt time.Time) (*entity.Ban, error) {
	ban := entity.Ban{
		UserId: sql.NullInt64{
			Int64: userId,
			Valid: true,
		},
		Cause:     cause,
		ExpiredAt: expiredAt,
	}

	if err := s.setBan(userId, &ban); err != nil {
		s.log.Error("ошибка при блокировке пользователя: ", err)
		return nil, err
	}

	return &ban, nil
}

func (s *BanService) BanUserForever(userId int64, cause string) (*entity.Ban, error) {
	ban := entity.Ban{
		UserId: sql.NullInt64{
			Int64: userId,
			Valid: true,
		},
		Cause:     cause,
		IsForever: true,
		ExpiredAt: time.Now(),
	}

	if err := s.setBan(userId, &ban); err != nil {
		s.log.Error("ошибка при блокировке пользователя: ", err)
		return nil, err
	}

	return &ban, nil
}

func (s *BanService) setBan(userId int64, ban *entity.Ban) error {
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	tx, err := s.repo.CreateTx()
	if err != nil {
		s.log.Error("ошибка при создании транзакции при блокировке пользователя: ", err)
		return err
	}

	if err := s.ts.logoutAll(ctx, tx, userId); err != nil {
		return err
	}

	if err := s.repo.SetBanTx(ctx, tx, ban); err != nil {
		return err
	}

	if err := s.repo.CommitTx(tx); err != nil {
		s.log.Error("ошибка при комите транзакции, при созранении бана: ", err)
		return err
	}

	return nil
}

func (s *BanService) UnBanUser(userId int64) bool {
	if err := s.repo.DeleteByUserId(userId); err != nil {
		s.log.Error("ошибка при разблокировке пользователя: ", err)
		return false
	}

	return true
}
