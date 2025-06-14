package models

type JwtToken struct {
	Ext   int64
	Iat   int64
	Role  []string
	Sub   int64
	Token string
}
