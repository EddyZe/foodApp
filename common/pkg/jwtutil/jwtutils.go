package jwtutil

import (
	"github.com/EddyZe/foodApp/common/models"
	"github.com/gin-gonic/gin"
	"strings"
)

func GenerateClaims(token *models.JwtClaims) map[string]interface{} {
	var rls string
	for i, role := range token.Role {
		if i != len(token.Role)-1 {
			rls += role + ","
		} else {
			rls += role
		}

	}

	return map[string]interface{}{
		"sub":            token.Sub,
		"email":          token.Email,
		"email_verified": token.EmailVerified,
		"roles":          rls,
	}
}

func ExtractBearerTokenHeader(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}
