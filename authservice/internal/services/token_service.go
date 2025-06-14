package services

import (
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	cfg   *config.TokenConfig
	redis *datasourse.Redis
}

func NewTokenService(cfg *config.TokenConfig, redis *datasourse.Redis) *TokenService {
	return &TokenService{
		cfg:   cfg,
		redis: redis,
	}
}

// GenerateJwt генерирует jwt токен
func (s *TokenService) GenerateJwt(claims map[string]interface{}) (string, error) {
	secret := s.cfg.Secret
	expiration := time.Now().Add(time.Duration(s.cfg.TokenExpirationMinute) * time.Minute).Unix()
	jwtClaims := jwt.MapClaims{
		"ext": expiration,
		"iat": time.Now().Unix(),
	}

	for key, value := range claims {
		jwtClaims[key] = value
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *TokenService) GenerateRefreshToken() string {

	return uuid.New().String()
}
