package services

import (
	"errors"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/sirupsen/logrus"
	"time"
)

type ResetPasswordService struct {
	log  *logrus.Entry
	cfg  *config.ResetPasswordVerificationCfg
	repo *repositories.ResetPasswordRepository
}

func NewResetPasswordService(log *logrus.Entry, cfg *config.ResetPasswordVerificationCfg, repo *repositories.ResetPasswordRepository) *ResetPasswordService {
	return &ResetPasswordService{
		log:  log,
		repo: repo,
		cfg:  cfg,
	}
}

func (s *ResetPasswordService) Save(userId int64, code string) (*entity.ResetPasswordCode, error) {
	expiredAt := time.Now().Add(time.Minute * time.Duration(s.cfg.CodeExpiredMinute))
	newCode := entity.ResetPasswordCode{
		UserId:    userId,
		Code:      code,
		ExpiredAt: expiredAt,
	}

	if err := s.repo.Save(&newCode); err != nil {
		return nil, errors.New(errormsg.IsExists)
	}

	return &newCode, nil
}
