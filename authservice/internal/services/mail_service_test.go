package services

import (
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestMailService(t *testing.T) {
	cfg := config.SmptConfig{
		Host: "localhost",
		Port: "1025",
	}

	mailServ := NewMailService(logrus.NewEntry(logrus.New()), &cfg)

	err := mailServ.SendMail("test2@mail", "test", "test bodty", "test@mail")

	if err != nil {
		t.Error(err)
	}
}
