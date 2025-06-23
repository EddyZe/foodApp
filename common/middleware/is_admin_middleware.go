package middleware

import (
	"github.com/EddyZe/foodApp/common/domain/models"
	"github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/gin-gonic/gin"
	"net/http"
)

func IsAdmin(locliz *localizer.LocalizeService) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "en"
		}
		claims, ok := c.Get("claims")
		if !ok {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			c.Abort()
			return
		}

		claimsMap := claims.(*models.JwtClaims)

		roles := claimsMap.Role

		for _, role := range roles {
			if role == "admin" {
				c.Next()
				return
			}
		}

		errMsg := locliz.GetMessage(
			localizer.Forbidden,
			lang,
			"Not enough right",
			nil,
		)

		responseutil.ErrorResponse(c, http.StatusForbidden, errMsg)
		c.Abort()
	}
}
