package services

import (
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/util/localizeids"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestGetMessage(t *testing.T) {
	ls, err := NewLocalizeService(logrus.NewEntry(logrus.New()), "./../../locales")
	if err != nil {
		t.Fatal(err)
	}
	res, err := ls.GetMessage(
		localizeids.SendVerifiedCodeBody,
		"ru",
		"hello",
		map[string]interface{}{
			"code": "test",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}
