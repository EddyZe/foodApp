package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/dto/auth"
	"github.com/EddyZe/foodApp/common/util/errormessages"
	"github.com/EddyZe/foodApp/common/util/redisutil"
	"github.com/EddyZe/foodApp/common/util/roles"
	"github.com/sirupsen/logrus"
)

const (
	EmailAlreadyExists = "email уже занят"
)

var userEmailKeys = "users:email"
var userIdKeys = "users:id"

type UserService struct {
	log   *logrus.Entry
	redis *datasourse.Redis
	rs    *RoleService
	ur    *repositories.UserRepository
	urs   *UserRoleService
}

func NewUserService(
	log *logrus.Entry,
	redis *datasourse.Redis,
	rs *RoleService,
	ur *repositories.UserRepository,
	urs *UserRoleService,
) *UserService {
	return &UserService{
		log:   log,
		redis: redis,
		rs:    rs,
		urs:   urs,
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
		return nil, errors.New(EmailAlreadyExists)
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
	if err := s.urs.SetRoleTx(ctx, tx, newUser.Id, role.Id); err != nil {
		s.log.Errorf("Ошибка при установки роли пользователю: %v", err)
		return nil, errors.New(fmt.Sprintf("%v: %v", errormessages.Save, err))
	}

	s.log.Debug("Отправка комита в БД ")
	if err := s.ur.CommitTx(tx); err != nil {
		s.log.Errorf("Ошибка при комите транзакции: %v", err)
		return nil, errors.New(fmt.Sprintf("%s: %v", errormessages.CommitTx, err))
	}

	if err := s.redis.Put(redisutil.GenerateKey(userEmailKeys, newUser.Email), newUser); err != nil {
		s.log.Errorf("ошибка записи в редис: %v", err)
	}

	if err := s.redis.Put(redisutil.GenerateKey(userIdKeys, fmt.Sprint(newUser.Id.Int64)), newUser); err != nil {
		s.log.Errorf("ошибка записи в редис: %v", err)
	}

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

	if err := s.redis.Put(redisKey, u); err != nil {
		s.log.Errorf("ошибка записи в редис: %v", err)
	}
	return u, true
}
