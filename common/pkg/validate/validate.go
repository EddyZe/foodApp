package validate

import (
	"fmt"
	localizer2 "github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/go-playground/validator/v10"
	"regexp"
	"strings"
)

func ValidateBody(validationErrors validator.ValidationErrors, ls *localizer2.LocalizeService, lang string) string {
	errorMessages := ""
	for _, fieldError := range validationErrors {
		switch fieldError.Tag() {
		case "required":
			msg := ls.GetMessage(
				localizer2.FieldRequired,
				lang,
				fmt.Sprintf("Field %s required", camelToSnake(fieldError.Field())),
				map[string]interface{}{
					"field": camelToSnake(fieldError.Field()),
				},
			)

			errorMessages += msg + "; "
		case "email":
			msg := ls.GetMessage(
				localizer2.FieldEmail,
				lang,
				fmt.Sprintf("Invalid email in field %s", camelToSnake(fieldError.Field())),
				map[string]interface{}{
					"field": camelToSnake(fieldError.Field()),
				},
			)

			errorMessages += msg
		case "min":
			msg := ls.GetMessage(
				localizer2.FieldMin,
				lang,
				fmt.Sprintf("Field %s must be at least %s characters long", camelToSnake(fieldError.Field()), fieldError.Param()),
				map[string]interface{}{
					"field": camelToSnake(fieldError.Field()),
					"param": fieldError.Param(),
				},
			)

			errorMessages += msg + "; "
		default:
			msg := ls.GetMessage(
				localizer2.FieldDefault,
				lang,
				fmt.Sprintf("An error in the field %s. Rule: %s", camelToSnake(fieldError.Field()), fieldError.Tag()),
				map[string]interface{}{
					"field": camelToSnake(fieldError.Field()),
					"rule":  fieldError.Tag(),
				},
			)

			errorMessages += msg + "; "
		}
	}

	lenMsg := len(errorMessages)
	if lenMsg > 2 {
		lenMsg -= 2
	} else {
		lenMsg -= 1
	}
	return errorMessages[:lenMsg]
}

func camelToSnake(s string) string {
	// Добавим подчеркивание перед каждой заглавной буквой (кроме первой)
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(snake)
}
