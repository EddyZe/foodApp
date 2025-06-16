package jwtutil

import "github.com/EddyZe/foodApp/authservice/internal/entity"

func GenerateClaims(userId int64, userRoles []entity.Role) map[string]interface{} {
	var rls string
	for i, role := range userRoles {
		if i != len(userRoles)-1 {
			rls += role.Name + ","
		} else {
			rls += role.Name
		}

	}

	return map[string]interface{}{
		"sub":  userId,
		"role": rls,
	}
}
