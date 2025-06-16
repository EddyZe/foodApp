package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"

	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/common/util/redisutil"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var rolesUserEmail = "roles:user:email"
var rolesUserid = "roles:user:id"
var roleName = "roles:name"

type RoleService struct {
	log   *logrus.Entry
	redis *datasourse.Redis
	repo  *repositories.RoleRepository
}

func NewRoleService(log *logrus.Entry, redis *datasourse.Redis, repo *repositories.RoleRepository) *RoleService {
	return &RoleService{
		log:   log,
		redis: redis,
		repo:  repo,
	}
}

func (s *RoleService) GetRoleByUserId(userId int64) []entity.Role {
	redisKey := redisutil.GenerateKey(rolesUserid, fmt.Sprint(userId))
	redisRes, ok := s.redis.Get(redisKey)
	if ok {
		var role []entity.Role
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

func (s *RoleService) FindByNameTx(ctx context.Context, tx *sqlx.Tx, name string) (*entity.Role, error) {
	redisKey := redisutil.GenerateKey(roleName, name)
	data, ok := s.redis.Get(redisKey)
	if ok {
		var role entity.Role
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
