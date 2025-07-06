package middleware

import (
	"github.com/EddyZe/foodApp/common/pkg/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/gin-gonic/gin"
	"net/http"
)

func JwtFilter(jwtsecret string, ls *localizer.LocalizeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		tokenString, ok := jwtutil.ExtractBearerTokenHeader(c)
		if !ok {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", getMsgUnAuthorized(ls, lang))
			c.Abort()
			return
		}

		claims, ok := jwtutil.ParseToken(tokenString, jwtsecret)
		if !ok {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", getMsgUnAuthorized(ls, lang))
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

func getMsgUnAuthorized(ls *localizer.LocalizeService, lang string) string {
	msg := ls.GetMessage(
		localizer.Unauthorized,
		lang,
		"You are not authorized",
		nil,
	)
	return msg
}
