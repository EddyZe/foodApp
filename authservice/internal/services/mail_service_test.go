package services

import (
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/models"
	"testing"
)

func TestMailService(t *testing.T) {
	cfg := config.SmptConfig{
		Host: "localhost",
		Port: "1025",
	}

	mailServ := NewMailService(&cfg)
	err := mailServ.SendMail(models.NewMailMessage("test2@mail", "test", "test bodty", "test@mail"))

	if err != nil {
		t.Error(err)
	}
}
