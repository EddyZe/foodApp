package pkg

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidateBody(validationErrors validator.ValidationErrors) string {
	errorMessages := ""
	for _, fieldError := range validationErrors {
		switch fieldError.Tag() {
		case "required":
			errorMessages += fmt.Sprintf("Поле '%s' обязательно, ", fieldError.Field())
		case "email":
			errorMessages += fmt.Sprintf("Некорректный email в поле '%s', ", fieldError.Field())
		case "min":
			errorMessages += fmt.Sprintf(
				"Поле '%s' должно быть не короче %s символов, ",
				fieldError.Field(),
				fieldError.Param(),
			)
		default:
			errorMessages = fmt.Sprintf(
				"Ошибка в поле '%s' (правило: %s)",
				fieldError.Field(),
				fieldError.Tag(),
			)
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
