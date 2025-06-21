package services

import (
	"context"
	"database/sql"
	"errors"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	entity2 "github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/util/codegen"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/sirupsen/logrus"
	"time"
)

type EmailVerificationService struct {
	log  *logrus.Entry
	cfg  *config.EmailVerificationCfg
	evr  *repositories.EmailVerificationCodeRepository
	evtr *repositories.EmailVerificationTokenRepository
}

func NewEmailVerificationCodeService(
	log *logrus.Entry,
	cfg *config.EmailVerificationCfg,
	evr *repositories.EmailVerificationCodeRepository,
	evtr *repositories.EmailVerificationTokenRepository,
) *EmailVerificationService {
	return &EmailVerificationService{
		log:  log,
		cfg:  cfg,
		evr:  evr,
		evtr: evtr,
	}
}

func (s *EmailVerificationService) Save(code *entity2.EmailVerificationCode) error {
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

func (s *EmailVerificationService) Delete(code string) error {
	return s.evr.DeleteByCode(code)
}

func (s *EmailVerificationService) GetByCode(code string) (*entity2.EmailVerificationCode, error) {
	res, err := s.evr.FindByCode(code)
	if err != nil {
		s.log.Error(err)
		return nil, errors.New(errormsg.NotFound)
	}

	return res, nil
}

func (s *EmailVerificationService) GenerateRandomCode(length int) string {
	return codegen.GenerateRandomCode(length)
}

func (s *EmailVerificationService) GenerateAndSaveCode(userId int64, lengthCode int) (*entity2.EmailVerificationCode, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx, err := s.evr.CreateTx()
	if err != nil {
		s.log.Error("Ошибка открытия транзакции при генерации кода для подтверждения email: ", err)
		return nil, err
	}

	var code string
	var codeEntity entity2.EmailVerificationCode

	for {
		code = codegen.GenerateRandomCode(lengthCode)
		if c, _ := s.evr.FindByCodeTx(ctx, tx, code); c != nil {
			continue
		}

		ex := time.Now().Add(time.Duration(s.cfg.CodeExpiredMinute) * time.Minute)

		codeEntity = entity2.EmailVerificationCode{
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
		return nil, errors.New(errormsg.ServerInternalError)
	}

	return &codeEntity, nil
}

func (s *EmailVerificationService) GetByEmailVerifToken(token string) (*entity2.EmailVerificationCode, bool) {
	code, err := s.evr.FindCodeByVerifiedToken(token)
	if err != nil {
		s.log.Debugf("код по токену %v не найден: %v", token, err)
		return nil, false
	}

	return code, true
}

func (s *EmailVerificationService) SaveVerificationToken(codeId int64, token string) error {
	expired := time.Now().Add(time.Duration(s.cfg.CodeExpiredMinute) * time.Minute)
	tok := &entity2.EmailVerificationToken{
		Id:        0,
		Token:     token,
		CodeId:    codeId,
		ExpiredAt: expired,
	}

	return s.evtr.Save(tok)
}

func (s *EmailVerificationService) SetIsActiveToken(token string, isActive bool) error {
	return s.evtr.SetIsActive(token, isActive)
}

func (s *EmailVerificationService) GetToken(token string) (*entity2.EmailVerificationToken, bool) {
	res, err := s.evtr.FindToken(token)
	if err != nil {
		s.log.Debugf("токен не %s найден: %v ", token, err)
		return nil, false
	}

	return res, true
}

func (s *EmailVerificationService) FindCode(code string) (*entity2.EmailVerificationCode, bool) {
	res, err := s.evr.FindByCode(code)
	if err != nil {
		s.log.Debug("код не был найден: ", err)
		return nil, false
	}
	return res, true
}

func (s *EmailVerificationService) SetVerifiedCode(codeString string, b bool) error {
	return s.evr.SetVerified(codeString, b)
}
