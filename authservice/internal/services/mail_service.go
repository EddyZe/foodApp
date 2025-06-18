package services

import (
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/sirupsen/logrus"
	"net/smtp"
)

type MailService struct {
	log  *logrus.Entry
	From string
	Host string
	Port string
	Auth smtp.Auth
}

func NewMailService(log *logrus.Entry, cfg *config.SmptConfig) *MailService {
	var auth smtp.Auth

	if cfg.Password == "" || cfg.Username == "" {
		auth = nil
	} else {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &MailService{
		log:  log,
		Host: cfg.Host,
		Port: cfg.Port,
		Auth: auth,
		From: cfg.From,
	}
}

func (s *MailService) SendMail(from, subject, body string, to ...string) error {
	sub := fmt.Sprintf("Subject: %v \n\n", subject)
	body = fmt.Sprintf("%v\n\n", body)
	if err := smtp.SendMail(
		fmt.Sprintf("%s:%s", s.Host, s.Port),
		s.Auth,
		from,
		to,
		[]byte(fmt.Sprint(sub, body)),
	); err != nil {
		s.log.Error("SMTP Send Mail Error ошибка при отпавке email:", err)
		return err
	}
	return nil
}

func (s *MailService) SendMailFromApp(subject, body string, to ...string) error {
	return s.SendMail(s.From, subject, body, to...)
}
