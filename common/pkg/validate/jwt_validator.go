package validate

import (
	"fmt"
	"strings"
	"time"

	"github.com/EddyZe/foodApp/common/models"
	"github.com/golang-jwt/jwt/v5"
)

func ValidJwtToken(tokenString, secret string) (*models.JwtToken, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("невалидная подпись токена: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, false
	}
	if !token.Valid {
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false
	}
	exp := claims["exp"].(int64)
	if time.Now().Unix() > exp {
		return nil, false
	}

	if claims["sub"] == nil || claims["role"] == nil || claims["iat"] == nil {
		return nil, false
	}

	sub := claims["sub"].(int64)
	roles := strings.Split(claims["roles"].(string), ",")

	jwtTok := models.JwtToken{
		Token: tokenString,
		Sub:   sub,
		Role:  roles,
		Ext:   exp,
		Iat:   claims["iat"].(int64),
	}

	return &jwtTok, true
}
