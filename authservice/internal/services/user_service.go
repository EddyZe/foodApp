package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/dto/auth"
	"github.com/EddyZe/foodApp/common/pkg/redisutil"
	"github.com/EddyZe/foodApp/common/pkg/roles"
	"github.com/EddyZe/foodApp/common/util/errormessages"
	"github.com/sirupsen/logrus"
)

var userEmailKeys = "users:email"
var userIdKeys = "users:id"

type UserService struct {
	log   *logrus.Entry
	redis *datasourse.Redis
	rs    *RoleService
	ur    *repositories.UserRepository
}

func NewUserService(
	log *logrus.Entry,
	redis *datasourse.Redis,
	rs *RoleService,
	ur *repositories.UserRepository,
) *UserService {
	return &UserService{
		log:   log,
		redis: redis,
		rs:    rs,
		ur:    ur,
	}
}

// CreateUser функция создает пользователя
func (s *UserService) CreateUser(dto *auth.RegisterDto) (*entity.User, error) {
	s.log.Debug("Создание пользователя")
	s.log.Debug("Хеширование пароля")
	passwordHash, err := pkg.PasswordHash(dto.Password)
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
		return nil, errors.New(fmt.Sprintf("%s: %v", errormessages.CreateTx, err))
	}

	s.log.Debug("Поиск пользователь с таким же email: ", dto.Email)
	if _, err := s.ur.FindByEmailTx(ctx, tx, dto.Email); err == nil {
		return nil, errors.New(errormsg.IsExists)
	}

	s.log.Debug("Сохранение пользователя")
	if err := s.ur.SaveTx(ctx, tx, &newUser); err != nil {
		s.log.Errorf("Ошибка при сохранении пользователя: %v", err)
		return nil, errors.New(fmt.Sprintf("%s: %v", errormessages.Save, err))
	}

	s.log.Debug("Поиск и роли")
	role, err := s.rs.FindByNameTx(ctx, tx, roles.User)
	if err != nil {
		s.log.Errorf("ошибка при поиске роли: %v", err)
		return nil, errors.New(fmt.Sprintf("%s: %v", errormessages.Find, err))
	}

	s.log.Debug("Установка роли пользователю: ", dto.Email)
	if err := s.rs.SetRoleTx(ctx, tx, newUser.Id, role.Id); err != nil {
		s.log.Errorf("Ошибка при установки роли пользователю: %v", err)
		return nil, errors.New(fmt.Sprintf("%v: %v", errormessages.Save, err))
	}

	s.log.Debug("Отправка комита в БД ")
	if err := s.ur.CommitTx(tx); err != nil {
		s.log.Errorf("Ошибка при комите транзакции: %v", err)
		return nil, errors.New(fmt.Sprintf("%s: %v", errormessages.CommitTx, err))
	}

	s.updateCache(&newUser)

	return &newUser, nil
}

func (s *UserService) HasEmailExists(email string) bool {
	redisKey := redisutil.GenerateKey(userEmailKeys, email)
	if _, ok := s.redis.Get(redisKey); ok {
		return true
	}
	return s.ur.HasEmailExists(email)
}

func (s *UserService) GetByEmail(email string) (*entity.User, bool) {
	redisKey := redisutil.GenerateKey(userEmailKeys, email)
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
	redisKey := redisutil.GenerateKey(userIdKeys, fmt.Sprint(id))
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

func (s *UserService) updateCache(u *entity.User) {
	rediskey2 := redisutil.GenerateKey(userEmailKeys, u.Email)
	redisKey := redisutil.GenerateKey(userIdKeys, fmt.Sprint(u.Id.Int64))
	if err := s.redis.Put(redisKey, u); err != nil {
		s.log.Errorf("ошибка при сохранении в redis при изменении статуса email: %v", err)
	}

	if err := s.redis.Put(rediskey2, u); err != nil {
		s.log.Errorf("ошибка при сохранении в redis при изменении статуса email: %v", err)
	}
}

func (s *UserService) removeCache(u *entity.User) {
	rediskey2 := redisutil.GenerateKey(userEmailKeys, u.Email)
	redisKey := redisutil.GenerateKey(userIdKeys, fmt.Sprint(u.Id.Int64))
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
