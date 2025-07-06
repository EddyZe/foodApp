package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/redis"
	"github.com/EddyZe/foodApp/authservice/internal/domain/dto"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/EddyZe/foodApp/authservice/internal/util/passencoder"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/common/pkg/redisutil"
	"github.com/EddyZe/foodApp/common/pkg/roles"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	log   *logrus.Entry
	redis *redis.Redis
	rs    *RoleService
	ur    *repositories.UserRepository
	hps   *HistoryPasswordService
}

func NewUserService(
	log *logrus.Entry,
	redis *redis.Redis,
	rs *RoleService,
	ur *repositories.UserRepository,
	hps *HistoryPasswordService,
) *UserService {
	return &UserService{
		log:   log,
		redis: redis,
		rs:    rs,
		ur:    ur,
		hps:   hps,
	}
}

// CreateUser функция создает пользователя
func (s *UserService) CreateUser(dto *dto.RegisterDto) (*entity.User, error) {
	s.log.Debug("Создание пользователя")
	s.log.Debug("Хеширование пароля")
	passwordHash, err := passencoder.PasswordHash(dto.Password)
	if err != nil {
		s.log.Errorf("Ошибка при хешировании пароля: %v", err)
		return nil, err
	}
	newUser := entity.User{
		Email:    dto.Email,
		Password: passwordHash,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.log.Debug("Создание транзакции для базы данных")
	tx, err := s.ur.CreateTx()
	if err != nil {
		s.log.Errorf("Ошибка при открытии транзакции: %v", err)
		return nil, err
	}

	s.log.Debug("Поиск пользователь с таким же email: ", dto.Email)
	if _, err := s.ur.FindByEmailTx(ctx, tx, dto.Email); err == nil {
		return nil, errors.New(errormsg.IsExists)
	}

	s.log.Debug("Сохранение пользователя")
	if err := s.ur.SaveTx(ctx, tx, &newUser); err != nil {
		s.log.Errorf("Ошибка при сохранении пользователя: %v", err)
		return nil, err
	}

	s.log.Debug("Поиск и роли")
	role, err := s.rs.FindByNameTx(ctx, tx, roles.User)
	if err != nil {
		s.log.Errorf("ошибка при поиске роли: %v", err)
		return nil, err
	}

	s.log.Debug("Установка роли пользователю: ", dto.Email)
	if err := s.rs.SetRoleTx(ctx, tx, newUser.Id, role.Id); err != nil {
		s.log.Errorf("Ошибка при установки роли пользователю: %v", err)
		return nil, err
	}

	s.log.Debug("Отправка комита в БД ")
	if err := s.ur.CommitTx(tx); err != nil {
		s.log.Errorf("Ошибка при комите транзакции: %v", err)
		return nil, err
	}

	s.updateCache(&newUser)

	return &newUser, nil
}

func (s *UserService) HasEmailExists(email string) bool {
	redisKey := redisutil.GenerateKey(redis.UserEmailKeys, email)
	if _, ok := s.redis.Get(redisKey); ok {
		return true
	}
	return s.ur.HasEmailExists(email)
}

func (s *UserService) GetByEmail(email string) (*entity.User, bool) {
	redisKey := redisutil.GenerateKey(redis.UserEmailKeys, email)
	jsonDataUser, ok := s.redis.Get(redisKey)
	if ok {
		var res entity.User
		if err := json.Unmarshal([]byte(jsonDataUser), &res); err != nil {
			s.log.Error("ошибка преобразования json из редис:", err)
		}
		return &res, true
	}

	u, err := s.ur.FindByEmail(email)
	if err != nil {
		s.log.Errorf("email не найден: %v", err)
		return nil, false
	}

	s.updateCache(u)
	return u, true
}

func (s *UserService) GetByRefreshToken(refreshToken string) (*entity.User, error) {
	u, err := s.ur.FindByRefreshToken(refreshToken)
	if err != nil {
		s.log.Errorf("пользователь с таким токеном не найден: %v", err)
		return nil, errors.New(errormsg.NotFound)
	}

	return u, nil
}

func (s *UserService) GetById(id int64) (*entity.User, error) {
	redisKey := redisutil.GenerateKey(redis.UserIdKeys, fmt.Sprint(id))
	if jsonDataUser, ok := s.redis.Get(redisKey); ok {
		var res entity.User
		if err := json.Unmarshal([]byte(jsonDataUser), &res); err == nil {
			return &res, nil
		}
	}

	u, err := s.ur.FindById(id)
	if err != nil {
		s.log.Debugf("пользователь с таким id не найден: %v", err)
		return nil, errors.New(errormsg.NotFound)
	}

	if err := s.redis.Put(redisKey, u); err != nil {
		s.log.Errorf("ошибка при созранении в редис пользователя: %v", err)
	}

	return u, nil
}

func (s *UserService) SetEmailConfirmed(userId int64, b bool) (*entity.User, error) {
	updateUser, err := s.ur.SetEmailIsConfirm(userId, b)
	if err != nil {
		s.log.Errorf("ошибка при обновлении статуса подтверждения email: %v", err)
		return nil, err
	}

	s.removeCache(updateUser)
	s.updateCache(updateUser)

	return updateUser, nil
}

func (s *UserService) EditPassword(userId int64, newPassword string) error {
	currentUser, err := s.GetById(userId)
	if err != nil {
		s.log.Errorf("ошибка при изменении пароля пользователя: %v", err)
		return errors.New(errormsg.NotFound)
	}

	s.removeCache(currentUser)

	currentPassword := currentUser.Password
	lastPassword := s.hps.GetLastPassword(userId)

	if passencoder.CheckEqualsPassword(newPassword, currentPassword) || passencoder.CheckEqualsPassword(newPassword, lastPassword.OldPassword) {
		s.log.Debug("Пароль равен текущему или последнему изменненному")
		return errors.New(errormsg.LastPasswordIsExists)
	}

	newPasswordHash, err := passencoder.PasswordHash(newPassword)
	if err != nil {
		s.log.Errorf("ошибка при генерации хеша пароля: %v", err)
		return err
	}

	currentUser.Password = newPasswordHash
	currentUser.UpdatedAt = time.Now()

	if err := s.hps.Save(userId, currentPassword); err != nil {
		s.log.Error("ошибка при сохранеии текущего пароля в истории: ", err)
	}

	if err := s.ur.EditPassword(userId, newPasswordHash); err != nil {
		s.log.Errorf("ошибка при обновлении пароля: %v", err)
		return err
	}
	s.updateCache(currentUser)
	return nil
}

func (s *UserService) updateCache(u *entity.User) {
	rediskey2 := redisutil.GenerateKey(redis.UserEmailKeys, u.Email)
	redisKey := redisutil.GenerateKey(redis.UserIdKeys, fmt.Sprint(u.Id.Int64))
	if err := s.redis.Put(redisKey, u); err != nil {
		s.log.Errorf("ошибка при сохранении в redis при изменении статуса email: %v", err)
	}

	if err := s.redis.Put(rediskey2, u); err != nil {
		s.log.Errorf("ошибка при сохранении в redis при изменении статуса email: %v", err)
	}
}

func (s *UserService) removeCache(u *entity.User) {
	rediskey2 := redisutil.GenerateKey(redis.UserEmailKeys, u.Email)
	redisKey := redisutil.GenerateKey(redis.UserIdKeys, fmt.Sprint(u.Id.Int64))
	if _, ok := s.redis.Get(redisKey); ok {
		if err := s.redis.Del(redisKey); err != nil {
			s.log.Debugf("%v ключ не удален из редис: %v", redisKey, err)
		}
	}

	if _, ok := s.redis.Get(rediskey2); ok {
		if err := s.redis.Del(rediskey2); err != nil {
			s.log.Debugf("%v ключ не удален из редис: %v", rediskey2, err)
		}
	}
}
