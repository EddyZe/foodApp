package services

import (
	"errors"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type EmailVerificationCodeService struct {
	log *logrus.Entry
	evr *repositories.EmailVerificationCodeRepository
}

func NewEmailVerificationCodeService(
	log *logrus.Entry,
	evr *repositories.EmailVerificationCodeRepository,
) *EmailVerificationCodeService {
	return &EmailVerificationCodeService{
		log: log,
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
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "1234567890QWERTYUIOPASDFGHJKLZXCVBNMqwertyuiopasdfghjklzxcvbnm"

	b := make([]byte, length)
	for i := range b {
		b[i] = chars[seededRand.Intn(len(chars))]
	}

	return string(b)
}
