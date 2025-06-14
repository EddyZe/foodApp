package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/common/util/redisutil"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type UserRoleService struct {
	log   *logrus.Entry
	redis *datasourse.Redis
	urr   *repositories.UserRoleRepository
}

func NewUserRoleService(log *logrus.Entry, redis *datasourse.Redis, urr *repositories.UserRoleRepository) *UserRoleService {
	return &UserRoleService{
		redis: redis,
		urr:   urr,
		log:   log,
	}
}

func (s *UserRoleService) SetRole(userId, roleId sql.NullInt64) error {
	s.log.Debug("Установка роли")
	if err := s.redis.Del(redisutil.GenerateKey(rolesUserid, fmt.Sprint(userId.Int64))); err != nil {
		s.log.Errorf("ошибка удаления ключа: %v", err)
	}
	if err := s.urr.SetUserRole(entity.UserRole{
		UserId: userId,
		RoleId: roleId,
	}); err != nil {
		s.log.Errorf("Ошибка при установки роли: %v", err)
		return err
	}
	s.log.Debug("Роль установлена")
	return nil
}

func (s *UserRoleService) SetRoleTx(ctx context.Context, tx *sqlx.Tx, userId, roleId sql.NullInt64) error {
	s.log.Debug("Установка роли")
	if err := s.redis.Del(redisutil.GenerateKey(rolesUserid, fmt.Sprint(userId.Int64))); err != nil {
		s.log.Errorf("ошибка удаления ключа: %v", err)
	}
	if err := s.urr.SetUserRoleTx(ctx, tx, &entity.UserRole{
		UserId: userId,
		RoleId: roleId,
	}); err != nil {
		s.log.Errorf("Ошибка при установки роли: %v", err)
		return err
	}
	s.log.Debug("Роль установлена")
	return nil
}
