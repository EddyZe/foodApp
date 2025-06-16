package middleware

import (
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/EddyZe/foodApp/common/pkg/validate"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func JwtFilter(jwtsecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)
		if token == "" {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			c.Abort()
			return
		}

		if _, ok := validate.ValidJwtToken(token, jwtsecret); !ok {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			c.Abort()
			return
		}

		c.Next()
	}
}
