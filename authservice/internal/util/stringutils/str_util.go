package stringutils

import (
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
)

func RoleMapString(roles []entity.Role) []string {
	var rls []string
	for _, role := range roles {
		rls = append(rls, role.Name)
	}

	return rls
}
