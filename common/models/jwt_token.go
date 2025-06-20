package models

type JwtClaims struct {
	Ext           int64
	Iat           int64
	Email         string
	EmailVerified bool
	Role          []string
	Sub           int64
}
