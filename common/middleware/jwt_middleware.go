package middleware

import (
	"github.com/EddyZe/foodApp/common/pkg/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/gin-gonic/gin"
	"net/http"
)

func JwtFilter(jwtsecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := jwtutil.ExtractBearerTokenHeader(c)
		if !ok {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			c.Abort()
			return
		}

		claims, ok := jwtutil.ParseToken(tokenString, jwtsecret)
		if !ok {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
