package rest

import "github.com/EddyZe/foodApp/common/pkg/localizer"

func unauthMsg(ls *localizer.LocalizeService, lang string) string {
	msg := ls.GetMessage(
		localizer.Unauthorized,
		lang,
		"Unauthorized",
		nil,
	)
	return msg
}
