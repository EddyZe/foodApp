package responseutil

import (
	"net/http"

	commonDto "github.com/EddyZe/foodApp/common/dto"
	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, code int, message string, data ...interface{}) {
	err := commonDto.APIError{
		Code:    http.StatusText(code),
		Message: message,
	}

	if data != nil || len(data) > 0 {
		err.Data = data
	}
	c.JSON(
		code,
		commonDto.Response{
			Success: false,
			Error:   err,
		},
	)
}

func SuccessResponse(c *gin.Context, code int, data interface{}) {
	c.JSON(code, commonDto.Response{
		Success: true,
		Data:    data,
	})
}
