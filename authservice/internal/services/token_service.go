package services

import (
	"context"
	"encoding/json"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/common/util/redisutil"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var refreshTokenKeyUserId = "refresh:token:user"

type TokenService struct {
	log   *logrus.Entry
	cfg   *config.TokenConfig
	redis *datasourse.Redis
	rs    *repositories.RefreshTokenRepository
}

func NewTokenService(cfg *config.TokenConfig, redis *datasourse.Redis, rs *repositories.RefreshTokenRepository, log *logrus.Entry) *TokenService {
	return &TokenService{
		log:   log,
		cfg:   cfg,
		redis: redis,
		rs:    rs,
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

func (s *TokenService) GenerateRefreshToken(userId int64) (*entity.RefreshToken, error) {
	s.log.Debug("Генерация рефрешь токена")
	refreshToken := uuid.New().String()
	s.log.Debug("установка истечении времени токену")
	expired := time.Now().Add(time.Duration(s.cfg.RefreshTokenExpirationMinute) * time.Minute).Unix()

	tok := entity.RefreshToken{
		UserId:    userId,
		Token:     refreshToken,
		ExpiredAt: time.Unix(expired, 0),
	}

	s.log.Debug("сохранение созданного токена")
	if err := s.rs.Save(&tok); err != nil {
		s.log.Error("ошибка при соханении рефрешь токена", err)
		return nil, err
	}

	s.log.Debug("сохранение токена в редис")
	redisKey := redisutil.GenerateKey(refreshTokenKeyUserId, refreshToken)
	if err := s.redis.Put(redisKey, tok); err != nil {
		s.log.Error("ошибка при сохранении токена в редис", err)
	}

	return &tok, nil
}

func (s *TokenService) ValidateRefreshToken(refreshToken string) bool {
	redisKey := redisutil.GenerateKey(refreshTokenKeyUserId, refreshToken)
	jsonToken, ok := s.redis.Get(redisKey)
	if !ok {
		jt, err := s.rs.FindByToken(refreshTokenKeyUserId)
		if err != nil {
			return false
		}
		data, err := json.Marshal(&jt)
		if err != nil {
			s.log.Error("ошибка при преобразовании json refresh token")
			return false
		}
		jsonToken = string(data)

		if err := s.redis.Put(redisKey, jsonToken); err != nil {
			s.log.Error("Ошибка при сохранеии в редис")
		}

	}

	var res entity.RefreshToken
	if err := json.Unmarshal([]byte(jsonToken), &res); err != nil {
		s.log.Error("ошибка при преобразовании json refresh token", err)
		return false
	}

	currentTime := time.Now().Unix()
	if currentTime > res.ExpiredAt.Unix() || res.IsRevoke {
		return false
	}

	return true

}

func (s *TokenService) ReplaceRefreshToken(refreshToken string) (*entity.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.log.Debug("создание транзазкции для замены рефреш токена: ", refreshToken)
	tx, err := s.rs.CreateTx()
	if err != nil {
		s.log.Error("Ошибка при создании транзации при попытке записать новый рефреш токен: ", err)
		return nil, err
	}

	var token *entity.RefreshToken

	rediskey := redisutil.GenerateKey(refreshTokenKeyUserId, refreshToken)

	s.log.Debug("поиск рефрешь токена в редисе")
	if jsonToken, ok := s.redis.Get(rediskey); ok {
		if err := json.Unmarshal([]byte(jsonToken), &token); err != nil {
			s.log.Error("Ошибка при конвертации json refresh token redis", err)
			t, err := s.rs.FindByTokenTx(ctx, tx, refreshToken)
			if err != nil {
				return nil, err
			}
			token = t
		}
		if err := s.redis.Del(rediskey); err != nil {
			s.log.Error("ошибка удаления токена из редис ", err)
		}
	}

	s.log.Debug("Поиск в редисе завершен")

	if token == nil {
		s.log.Debug("ищем токен в БД")
		token, err = s.rs.FindByTokenTx(ctx, tx, refreshToken)
		if err != nil {
			s.log.Error("Ошибка при получении нового refresh token", err)
			return nil, err
		}
	}

	s.log.Debug("удаляем старый токе")
	if err := s.rs.DeleteByTokenTx(ctx, tx, refreshToken); err != nil {
		s.log.Error("Ошибка при удалении рефрешь токена", err)
		return nil, err
	}

	s.log.Debug("применяем транзакцию")
	if err := s.rs.CommitTx(tx); err != nil {
		s.log.Error("Ошибка при комите транзакции при замене токена: ", err)
		return nil, err
	}

	s.log.Debug("токен удален")
	s.log.Debug("начата генерация нового токена")
	return s.GenerateRefreshToken(token.UserId)
}
