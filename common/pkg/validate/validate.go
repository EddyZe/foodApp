package validate

import (
	"fmt"
	"github.com/EddyZe/foodApp/common/services/localizer"
	"github.com/go-playground/validator/v10"
)

func ValidateBody(validationErrors validator.ValidationErrors, ls *localizer.LocalizeService, lang string) string {
	errorMessages := ""
	for _, fieldError := range validationErrors {
		switch fieldError.Tag() {
		case "required":
			msg := ls.GetMessage(
				localizer.FieldRequired,
				lang,
				fmt.Sprintf("Field %s required", fieldError.Field()),
				map[string]interface{}{
					"field": fieldError.Field(),
				},
			)

			errorMessages += msg + "; "
		case "email":
			msg := ls.GetMessage(
				localizer.FieldEmail,
				lang,
				fmt.Sprintf("Invalid email in field %s", fieldError.Field()),
				map[string]interface{}{
					"field": fieldError.Field(),
				},
			)

			errorMessages += msg
		case "min":
			msg := ls.GetMessage(
				localizer.FieldMin,
				lang,
				fmt.Sprintf("Field %s must be at least %s characters long", fieldError.Field(), fieldError.Param()),
				map[string]interface{}{
					"field": fieldError.Field(),
					"param": fieldError.Param(),
				},
			)

			errorMessages += msg + "; "
		default:
			msg := ls.GetMessage(
				localizer.FieldDefault,
				lang,
				fmt.Sprintf("An error in the field %s. Rule: %s", fieldError.Field(), fieldError.Tag()),
				map[string]interface{}{
					"field": fieldError.Field(),
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
