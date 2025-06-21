package jwtutil

import (
	"fmt"
	"github.com/EddyZe/foodApp/common/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

func ParseToken(tokenString, secret string) (*models.JwtClaims, bool) {
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

	c, ok := getClaimsByToken(claims)
	if !ok {
		return nil, false
	}

	if c.Ext < time.Now().Unix() {
		return nil, false
	}

	if c.Iat > c.Ext {
		return nil, false
	}

	return c, true
}

func getClaimsByToken(claims jwt.MapClaims) (*models.JwtClaims, bool) {
	exp, ok := claims["ext"].(float64)
	if !ok {
		return nil, false
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return nil, false
	}

	rolesString, ok := claims["roles"].(string)
	if !ok {
		return nil, false
	}
	roles := strings.Split(rolesString, ",")

	email, ok := claims["email"].(string)
	if !ok {
		return nil, false
	}

	emailVerified, ok := claims["email_verified"].(bool)
	if !ok {
		return nil, false
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, false
	}

	jwtTok := models.JwtClaims{
		Sub:           int64(sub),
		Role:          roles,
		Ext:           int64(exp),
		Email:         email,
		EmailVerified: emailVerified,
		Iat:           int64(iat),
	}
	return &jwtTok, true
}
