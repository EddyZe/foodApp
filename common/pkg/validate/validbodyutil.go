package validate

import (
	"errors"
	localizer2 "github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func IsValidBody(c *gin.Context, body any, ls *localizer2.LocalizeService) (string, bool) {
	lang := c.GetHeader("Accept-language")
	if lang == "" {
		lang = "en"
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		var validationErrors validator.ValidationErrors
		errorMessages := ""
		if errors.As(err, &validationErrors) {
			errorMessages = ValidateBody(validationErrors, ls, lang)
		} else {
			errorMessages = ls.GetMessage(
				localizer2.InvalidBody,
				lang,
				"Invalid body",
				nil,
			)
		}
		return errorMessages, false
	}
	return "", true
}
