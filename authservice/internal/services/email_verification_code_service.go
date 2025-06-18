package services

import (
	"context"
	"database/sql"
	"errors"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/util/codegen"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/sirupsen/logrus"
	"time"
)

type EmailVerificationCodeService struct {
	log *logrus.Entry
	cfg *config.EmailVerificationCfg
	evr *repositories.EmailVerificationCodeRepository
}

func NewEmailVerificationCodeService(
	log *logrus.Entry,
	cfg *config.EmailVerificationCfg,
	evr *repositories.EmailVerificationCodeRepository,
) *EmailVerificationCodeService {
	return &EmailVerificationCodeService{
		log: log,
		cfg: cfg,
		evr: evr,
	}
}

func (s *EmailVerificationCodeService) Save(code *entity.EmailVerificationCode) error {
	c := code.Code
	if c, _ := s.evr.FindByCode(c); c != nil {
		return errors.New(errormsg.IsExists)
	} else {
		if err := s.evr.Save(code); err != nil {
			s.log.Error(err)
			return errors.New(errormsg.ServerInternalError)
		}
	}
	return nil
}

func (s *EmailVerificationCodeService) Delete(code string) error {
	return s.evr.DeleteByCode(code)
}

func (s *EmailVerificationCodeService) GetByCode(code string) (*entity.EmailVerificationCode, error) {
	res, err := s.evr.FindByCode(code)
	if err != nil {
		s.log.Error(err)
		return nil, errors.New(errormsg.NotFound)
	}

	return res, nil
}

func (s *EmailVerificationCodeService) GenerateRandomCode(length int) string {
	return codegen.GenerateRandomCode(length)
}

func (s *EmailVerificationCodeService) GenerateAndSaveCode(userId int64, lengthCode int) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := s.evr.CreateTx()
	if err != nil {
		s.log.Error("Ошибка открытия транзакции при генерации кода для подтверждения email: ", err)
		return "", errors.New(errormsg.ServerInternalError)
	}

	var code string

	for {
		code = codegen.GenerateRandomCode(lengthCode)
		if c, _ := s.evr.FindByCodeTx(ctx, tx, code); c != nil {
			continue
		}

		ex := time.Now().Add(time.Duration(s.cfg.CodeExpiredMinute) * time.Minute)

		codeEntity := entity.EmailVerificationCode{
			Id:        sql.NullInt64{},
			UserId:    userId,
			Code:      code,
			ExpiredAt: ex,
		}

		if err := s.evr.SaveTx(ctx, tx, &codeEntity); err != nil {
			s.log.Error("Ошибка при сохранении кода для подтверждения почты: ", err)
			continue
		}

		break
	}

	if err := s.evr.CommitTx(tx); err != nil {
		s.log.Error("ошибка при комите транзакции при генерации и сохранени email code: ", err)
		return "", errors.New(errormsg.ServerInternalError)
	}

	return code, nil
}
