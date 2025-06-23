package localizer

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestGetMessage(t *testing.T) {
	ls := NewLocalizeService(logrus.NewEntry(logrus.New()), "./locales")
	res := ls.GetMessage(
		SendVerifiedCodeBody,
		"ru",
		"hello {{.code}}",
		map[string]interface{}{
			"code": "test",
		},
	)

	fmt.Println(res)
}
