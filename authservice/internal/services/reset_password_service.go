package services

import (
	"errors"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/util/codegen"
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
		s.log.Error("ошибка при сохрании кода сброса пароля: ", err)
		return nil, errors.New(errormsg.IsExists)
	}

	return &newCode, nil
}

func (s *ResetPasswordService) SetIsValid(code string, b bool) error {
	if err := s.repo.SetIsValid(code, b); err != nil {
		s.log.Error("ошибка при обновлении статуса resetPassword: ", err)
		return errors.New(errormsg.NotFound)
	}
	return nil
}

func (s *ResetPasswordService) DeleteByCode(code string) error {
	if err := s.repo.DeleteByCode(code); err != nil {
		s.log.Error("ошибка при удалении кода сбросов пароля: ", err)
		return errors.New(errormsg.NotFound)
	}
	return nil
}

func (s *ResetPasswordService) GetCode(code string) (*entity.ResetPasswordCode, error) {
	resetCode, err := s.repo.FindByCode(code)
	if err != nil {
		s.log.Error("ошибка при поиске кода: ", err)
		return nil, errors.New(errormsg.NotFound)
	}
	return resetCode, nil
}

func (s *ResetPasswordService) GenerateAndSaveCode(userId int64) (*entity.ResetPasswordCode, error) {
	for {
		codeString := codegen.GenerateRandomCode(6)
		code, err := s.Save(userId, codeString)
		if err != nil {
			if err.Error() == errormsg.IsExists {
				continue
			}
			return nil, err
		}
		return code, nil
	}
}
