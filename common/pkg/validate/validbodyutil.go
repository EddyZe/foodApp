package validate

import (
	"errors"
	"net/http"

	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func IsValidBody(c *gin.Context, body any) bool {
	if err := c.ShouldBindJSON(&body); err != nil {
		var validationErrors validator.ValidationErrors
		errorMessages := ""
		if errors.As(err, &validationErrors) {
			errorMessages = pkg.ValidateBody(validationErrors)
		} else {
			errorMessages = "невалидное тело запроса"
		}
		responseutil.ErrorResponse(c, http.StatusBadRequest, errorMessages)
		return false
	}
	return true
}
