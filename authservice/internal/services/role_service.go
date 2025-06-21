package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/redis"
	entity2 "github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"

	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/common/pkg/redisutil"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var rolesUserid = "roles:user:id"
var roleName = "roles:name"

type RoleService struct {
	log   *logrus.Entry
	redis *redis.Redis
	repo  *repositories.RoleRepository
	urr   *repositories.UserRoleRepository
}

func NewRoleService(log *logrus.Entry, redis *redis.Redis, repo *repositories.RoleRepository, urr *repositories.UserRoleRepository) *RoleService {
	return &RoleService{
		log:   log,
		redis: redis,
		repo:  repo,
		urr:   urr,
	}
}

func (s *RoleService) GetRoleByUserId(userId int64) []entity2.Role {
	redisKey := redisutil.GenerateKey(rolesUserid, fmt.Sprint(userId))
	redisRes, ok := s.redis.Get(redisKey)
	if ok {
		var role []entity2.Role
		if err := json.Unmarshal([]byte(redisRes), &role); err != nil {
			s.log.Errorf("ошибка приобразования json из redis: %v", err)
		}
		return role
	}

	res := s.repo.GetRoleByUserId(userId)
	if err := s.redis.Put(redisKey, res); err != nil {
		s.log.Errorf("ошибка записи в редис: %v", err)
	}

	return res
}

func (s *RoleService) FindByNameTx(ctx context.Context, tx *sqlx.Tx, name string) (*entity2.Role, error) {
	redisKey := redisutil.GenerateKey(roleName, name)
	data, ok := s.redis.Get(redisKey)
	if ok {
		var role entity2.Role
		if err := json.Unmarshal([]byte(data), &role); err != nil {
			s.log.Errorf("ошибка преобразования json из redis: %v", err)
		}

		return &role, nil
	}

	role, err := s.repo.FindByNameTx(ctx, tx, name)
	if err != nil {
		return nil, fmt.Errorf(errormsg.NotFound)
	}

	if err := s.redis.Put(redisKey, role); err != nil {
		s.log.Errorf("ошибка записи в редис: %v", err)
	}

	return role, nil
}

func (s *RoleService) SetRole(userId, roleId sql.NullInt64) error {
	s.log.Debug("Установка роли")
	if err := s.redis.Del(redisutil.GenerateKey(rolesUserid, fmt.Sprint(userId.Int64))); err != nil {
		s.log.Errorf("ошибка удаления ключа: %v", err)
	}
	if err := s.urr.SetUserRole(entity2.UserRole{
		UserId: userId,
		RoleId: roleId,
	}); err != nil {
		s.log.Errorf("Ошибка при установки роли: %v", err)
		return err
	}
	s.log.Debug("Роль установлена")
	return nil
}

func (s *RoleService) SetRoleTx(ctx context.Context, tx *sqlx.Tx, userId, roleId sql.NullInt64) error {
	s.log.Debug("Установка роли")
	if err := s.redis.Del(redisutil.GenerateKey(rolesUserid, fmt.Sprint(userId.Int64))); err != nil {
		s.log.Errorf("ошибка удаления ключа: %v", err)
	}
	if err := s.urr.SetUserRoleTx(ctx, tx, &entity2.UserRole{
		UserId: userId,
		RoleId: roleId,
	}); err != nil {
		s.log.Errorf("Ошибка при установки роли: %v", err)
		return err
	}
	s.log.Debug("Роль установлена")
	return nil
}
