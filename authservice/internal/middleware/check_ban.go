package middleware

import (
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/gin-gonic/gin"
)

func CheckBun(bs *services.BanService) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
