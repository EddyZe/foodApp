package handlers

import (
	"net/http"

	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/dto/auth"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/EddyZe/foodApp/common/pkg/validate"
	"github.com/EddyZe/foodApp/common/util/roles"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	us  *services.UserService
	ts  *services.TokenService
	rs  *services.RoleService
	log *logrus.Entry
	bs  *services.BanService
}

func NewAuthHandler(log *logrus.Entry, us *services.UserService, ts *services.TokenService,
	rs *services.RoleService, bs *services.BanService) *AuthHandler {
	return &AuthHandler{
		us:  us,
		log: log,
		ts:  ts,
		rs:  rs,
		bs:  bs,
	}
}

func (h *AuthHandler) Ping(c *gin.Context) {
	c.String(200, "pong")
}

func (h *AuthHandler) Registry(c *gin.Context) {
	var registerDto auth.RegisterDto

	if !validate.IsValidBody(c, &registerDto) {
		return
	}

	user, err := h.us.CreateUser(&registerDto)
	if err != nil {
		if err.Error() == services.EmailAlreadyExists {
			responseutil.ErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			h.log.Error(err)
			responseutil.ErrorResponse(c, http.StatusInternalServerError, "ошибка на стороне сервера")
		}

		return
	}

	token, err := h.ts.GenerateJwt(
		map[string]interface{}{
			"sub":  user.Id.Int64,
			"role": roles.User,
		},
	)
	if err != nil {
		h.log.Error(err)
	}

	responseutil.SuccessResponse(c, http.StatusCreated, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: "test",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginDto auth.LoginDto

	if !validate.IsValidBody(c, &loginDto) {
		return
	}

	u, ok := h.us.GetByEmail(loginDto.Email)

	if !ok || !pkg.CheckEqualsPassword(loginDto.Password, u.Password) {
		responseutil.ErrorResponse(c, http.StatusBadRequest, "неверный логин или пароль")
		return
	}

	if ban, ok := h.bs.GetActiveUserBan(u.Id.Int64); ok {
		responseutil.ErrorResponse(
			c,
			http.StatusForbidden,
			"Аккаунт заблокирован",
			ban,
		)
		return
	}

	userRoles := h.rs.GetRoleByUserId(u.Id.Int64)

	rls := ""
	for i, role := range userRoles {
		if i != len(userRoles)-1 {
			rls += role.Name + ","
		} else {
			rls += role.Name
		}

	}

	token, err := h.ts.GenerateJwt(
		map[string]interface{}{
			"sub":  u.Id.Int64,
			"role": rls,
		},
	)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, "Ошибка на стороне сервера")
	}

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: "test",
	})

}
