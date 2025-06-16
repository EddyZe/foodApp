package services

import (
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/models"
	"net/smtp"
)

type MailService struct {
	Host string
	Port string
	Auth smtp.Auth
}

func NewMailService(cfg *config.SmptConfig) *MailService {
	var auth smtp.Auth

	if cfg.Password == "" || cfg.Username == "" {
		auth = nil
	} else {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &MailService{
		Host: cfg.Host,
		Port: cfg.Port,
		Auth: auth,
	}
}

func (s *MailService) SendMail(m *models.MailMessage) error {
	sub := fmt.Sprintf("Subject: %v \n\n", m.Subject())
	body := fmt.Sprintf("%v\n\n", m.Body())
	if err := smtp.SendMail(
		fmt.Sprintf("%s:%s", s.Host, s.Port),
		s.Auth,
		m.From(),
		m.To(),
		[]byte(fmt.Sprint(sub, body)),
	); err != nil {
		return err
	}
	return nil
}
