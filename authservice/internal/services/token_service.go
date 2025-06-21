package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/redis"
	entity2 "github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/EddyZe/foodApp/authservice/internal/util/stringutils"
	"github.com/EddyZe/foodApp/common/domain/models"
	"github.com/EddyZe/foodApp/common/pkg/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/redisutil"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var refreshTokenUser = "refresh:token:user"

type TokenService struct {
	log   *logrus.Entry
	cfg   *config.TokenConfig
	redis *redis.Redis
	rs    *repositories.RefreshTokenRepository
	ar    *repositories.AccessTokenRepository
}

func NewTokenService(
	cfg *config.TokenConfig,
	redis *redis.Redis,
	rs *repositories.RefreshTokenRepository,
	log *logrus.Entry,
	ar *repositories.AccessTokenRepository,
) *TokenService {
	return &TokenService{
		log:   log,
		cfg:   cfg,
		redis: redis,
		rs:    rs,
		ar:    ar,
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

func (s *TokenService) GenerateJwtByUser(u *entity2.User, roles []entity2.Role) (string, error) {
	token, err := s.GenerateJwt(
		jwtutil.GenerateClaims(&models.JwtClaims{
			Email:         u.Email,
			EmailVerified: u.EmailIsConfirm,
			Role:          stringutils.RoleMapString(roles),
			Sub:           u.Id.Int64,
		}),
	)
	if err != nil {
		s.log.Error(err)
		return "", err
	}

	return token, nil
}

func (s *TokenService) GenerateUUID() string {
	s.log.Debug("Генерация токена")
	refreshToken := uuid.New().String()
	return refreshToken
}

func (s *TokenService) ValidateRefreshToken(refreshToken string) bool {
	redisKey := redisutil.GenerateKey(refreshTokenUser, refreshToken)
	jsonToken, ok := s.redis.Get(redisKey)
	if !ok {
		jt, err := s.rs.FindByToken(refreshToken)
		if err != nil {
			return false
		}
		data, err := json.Marshal(&jt)
		if err != nil {
			s.log.Error("ошибка при преобразовании json refresh token")
			return false
		}
		jsonToken = string(data)

		ex := time.Duration(int64(s.cfg.RefreshTokenExpirationMinute)) * time.Minute

		if err := s.redis.PutEx(redisKey, jsonToken, ex); err != nil {
			s.log.Error("Ошибка при сохранеии в редис")
		}

	}

	var res entity2.RefreshToken
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

func (s *TokenService) SaveRefreshAndAccessToken(userId int64, accessToken, refreshToken string) (*entity2.AccessToken, *entity2.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tx, err := s.ar.CreateTx()
	if err != nil {
		s.log.Error("ошибка создания транзакции: ", err)
		return nil, nil, err
	}
	expiredAccess := time.Now().Add(time.Duration(s.cfg.TokenExpirationMinute) * time.Minute)

	at := entity2.AccessToken{
		Token:     accessToken,
		ExpiredAt: expiredAccess,
	}

	if err := s.ar.SaveTx(ctx, tx, &at); err != nil {
		s.log.Error("ошибка сохранения access token: ", err)
		return nil, nil, err
	}

	expiredRefresh := time.Now().Add(time.Duration(s.cfg.RefreshTokenExpirationMinute) * time.Minute)
	rt := entity2.RefreshToken{
		UserId:        userId,
		AccessTokenId: at.Id,
		Token:         refreshToken,
		ExpiredAt:     expiredRefresh,
	}

	s.log.Debug("сохранение созданного токена")
	if err := s.rs.SaveTx(ctx, tx, &rt); err != nil {
		s.log.Error("ошибка при соханении рефрешь токена", err)
		return nil, nil, err
	}

	if err := s.rs.CommitTx(tx); err != nil {
		s.log.Error("ошибка при комите транщакции: ", err)
		return nil, nil, err
	}

	s.log.Debug("сохранение токена в редис")
	redisKey := redisutil.GenerateKey(refreshTokenUser, refreshToken)

	ex := time.Duration(int64(s.cfg.RefreshTokenExpirationMinute)) * time.Minute

	if err := s.redis.PutEx(redisKey, rt, ex); err != nil {
		s.log.Error("ошибка при сохранении токена в редис ", err)
	}

	return &at, &rt, err
}

func (s *TokenService) ReplaceTokens(refreshToken string, claims map[string]interface{}) (*entity2.AccessToken, *entity2.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.log.Debug("создание транзазкции для замены рефреш токена: ", refreshToken)
	tx, err := s.rs.CreateTx()
	if err != nil {
		s.log.Error("Ошибка при создании транзации при попытке записать новый рефреш токен: ", err)
		return nil, nil, err
	}

	var token *entity2.RefreshToken

	rediskey := redisutil.GenerateKey(refreshTokenUser, refreshToken)

	s.log.Debug("поиск рефрешь токена в редисе")
	if jsonToken, ok := s.redis.Get(rediskey); ok {
		if err := json.Unmarshal([]byte(jsonToken), &token); err != nil {
			s.log.Error("Ошибка при конвертации json refresh token redis", err)
			t, err := s.rs.FindByTokenTx(ctx, tx, refreshToken)
			if err != nil {
				s.log.Error("Refresh token не найден в базе")
				return nil, nil, errors.New(errormsg.NotFound)
			}
			token = t
		}
		if err := s.redis.Del(rediskey); err != nil {
			s.log.Error("ошибка удаления токена из редис ", err)
		}
	}

	s.log.Debug("Поиск в редисе завершен")

	token, err = s.rs.FindByTokenTx(ctx, tx, refreshToken)
	if err != nil {
		s.log.Error("refresh token не найден в базе")
		return nil, nil, errors.New(errormsg.NotFound)
	}

	s.log.Debug("удаляем связанные access токены")

	if err := s.ar.DeleteByRefreshTokenByTokenTx(ctx, tx, refreshToken); err != nil {
		s.log.Error("ошибка при удалении access токеновт", err)
	}

	s.log.Debug("удаляем старый токен")
	if err := s.rs.DeleteByTokenTx(ctx, tx, refreshToken); err != nil {
		s.log.Error("Ошибка при удалении рефрешь токена", err)
		return nil, nil, err
	}

	s.log.Debug("применяем транзакцию")
	if err := s.rs.CommitTx(tx); err != nil {
		s.log.Error("Ошибка при комите транзакции при замене токена: ", err)
		return nil, nil, err
	}
	s.log.Debug("токены удалены")

	newRefreshToken := s.GenerateUUID()
	newAccessToken, err := s.GenerateJwt(claims)
	if err != nil {
		s.log.Error("ошибка генерации access токена")
		return nil, nil, err
	}

	return s.SaveRefreshAndAccessToken(token.UserId, newAccessToken, newRefreshToken)
}

func (s *TokenService) IsRevokeRefreshToken(refreshToken string, isRevoke bool) error {
	redisKey := redisutil.GenerateKey(refreshTokenUser, refreshToken)
	var res *entity2.RefreshToken
	if jsonData, ok := s.redis.Get(redisKey); ok {
		if err := json.Unmarshal([]byte(jsonData), &res); err != nil {
			s.log.Error("ошибка десириализации json refresh token", err)
		}
		if err := s.redis.Del(redisKey); err != nil {
			s.log.Error("ошибка удаления из редис refresh token", err)
		}
	}
	if res == nil {
		r, err := s.rs.FindByToken(refreshTokenUser)
		if err != nil {
			s.log.Error("<UNK> <UNK> <UNK> refresh token", err)
			return errors.New(errormsg.NotFound)
		}
		res = r
	}

	res.IsRevoke = isRevoke
	if err := s.rs.Update(res); err != nil {
		s.log.Error("при изменении статуса refresh token", err)
		return err
	}

	ex := time.Duration(int64(s.cfg.RefreshTokenExpirationMinute)) * time.Minute

	if err := s.redis.PutEx(redisKey, res, ex); err != nil {
		s.log.Error("ошибка при добавлении в редис refresh token", err)
	}
	return nil
}

func (s *TokenService) Logout(accessToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	tx, err := s.rs.CreateTx()
	if err != nil {
		s.log.Error("Ошибка при открытии транзакции: ", err)
		return err
	}

	refreshtoken, err := s.rs.RemoveByAccessTokenTx(
		ctx,
		tx,
		accessToken,
	)
	if err != nil {
		s.log.Debug("токен не найден refresh token: ", err)
		return err
	}

	redisKey := redisutil.GenerateKey(refreshTokenUser, refreshtoken)
	if err := s.redis.Del(redisKey); err != nil {
		s.log.Error("ошибка удаления refresh token из redis", err)
	}

	if err := s.ar.DeleteByTokenTx(
		ctx,
		tx,
		accessToken,
	); err != nil {
		s.log.Debug("токен не найден access token: ", err)
		return err
	}

	if err := s.ar.CommitTx(tx); err != nil {
		s.log.Error("ошибка при комите транзакции во время удаления токенов: ", err)
		return err
	}
	return nil
}

func (s *TokenService) LogoutAll(userid int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	tx, err := s.rs.CreateTx()
	if err != nil {
		s.log.Error("Ошибка при создании транзакции при удалении всех токенов: ", err)
		return err
	}
	tokens, err := s.rs.RemoveAllRefreshTokensUserTx(ctx, tx, userid)
	if err != nil {
		s.log.Error("ошибка удаления refresh токенов: ", err)
		return err
	}

	s.log.Info(tokens)

	var ids []int64

	for _, token := range tokens {
		ids = append(ids, token.AccessTokenId.Int64)
		redisKey := redisutil.GenerateKey(refreshTokenUser, token.Token)
		if err := s.redis.Del(redisKey); err != nil {
			continue
		}
	}

	if err := s.ar.DeleteByIdsTx(ctx, tx, ids...); err != nil {
		s.log.Error("ошибка при удалении access токеновт", err)
		return err
	}

	if err := s.rs.CommitTx(tx); err != nil {
		s.log.Error("ошибка при комите транзацкии при удалении токенов: ", err)
		return err
	}

	return nil
}

func (s *TokenService) Secret() string {
	return s.cfg.Secret
}
