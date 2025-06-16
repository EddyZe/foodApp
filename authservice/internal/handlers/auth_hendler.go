package handlers

import (
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/util/jwtutil"
	"net/http"
	"strings"

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

	refreshToken, err := h.ts.GenerateRefreshToken(user.Id.Int64)
	if err != nil {
		h.log.Error(err)
	}

	responseutil.SuccessResponse(c, http.StatusCreated, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken.Token,
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

	if _, ok := h.checkBan(c, u.Id.Int64); ok {
		return
	}

	userRoles := h.rs.GetRoleByUserId(u.Id.Int64)

	token, err := h.ts.GenerateJwt(
		jwtutil.GenerateClaims(u.Id.Int64, userRoles),
	)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, "Ошибка на стороне сервера")
	}

	refreshToken, err := h.ts.GenerateRefreshToken(u.Id.Int64)
	if err != nil {
		h.log.Error(err)
	}

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken.Token,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if !h.ts.ValidateRefreshToken(token) {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	res, err := h.ts.ReplaceRefreshToken(token)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusUnauthorized, "ошибка на стороне сервера")
		return
	}

	if _, ok := h.checkBan(c, res.UserId); ok {
		return
	}

	accessToken, err := h.ts.GenerateJwt(
		jwtutil.GenerateClaims(res.UserId, h.rs.GetRoleByUserId(res.UserId)),
	)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
		AccessToken:  accessToken,
		RefreshToken: res.Token,
	})
}

func (h *AuthHandler) checkBan(c *gin.Context, userId int64) (*entity.Ban, bool) {
	if ban, ok := h.bs.GetActiveUserBan(userId); ok {
		responseutil.ErrorResponse(
			c,
			http.StatusForbidden,
			"Аккаунт заблокирован",
			ban,
		)
		return ban, true
	}
	return nil, false
}
