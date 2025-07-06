package responseutil

import (
	commonDto "github.com/EddyZe/foodApp/common/domain/dto"
	"github.com/gin-gonic/gin"
)

func ErrorResponse(c *gin.Context, status int, code, message string, details ...interface{}) {
	err := commonDto.APIError{
		Code:    code,
		Message: message,
	}

	if details != nil || len(details) > 0 {
		err.Details = details
	}
	c.JSON(
		status,
		commonDto.Response{
			Status: status,
			Error:  err,
		},
	)
}

func SuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, commonDto.Response{
		Status: status,
		Data:   data,
	})
}
