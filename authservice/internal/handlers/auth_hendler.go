package handlers

import (
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/EddyZe/foodApp/authservice/internal/util/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/headers"
	"net/http"
	"strconv"
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
		if err.Error() == errormsg.IsExists {
			responseutil.ErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			h.log.Error(err)
			responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
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

	refreshToken := h.ts.GenerateRefreshToken()

	if _, _, err := h.ts.SaveRefreshAndAccessToken(user.Id.Int64, token, refreshToken); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusCreated, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginDto auth.LoginDto

	if !validate.IsValidBody(c, &loginDto) {
		return
	}

	u, ok := h.us.GetByEmail(loginDto.Email)

	if !ok || !pkg.CheckEqualsPassword(loginDto.Password, u.Password) {
		responseutil.ErrorResponse(
			c,
			http.StatusBadRequest,
			errormsg.InvalidEmailOrPassword,
		)
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
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
	}

	refreshToken := h.ts.GenerateRefreshToken()

	if _, _, err := h.ts.SaveRefreshAndAccessToken(u.Id.Int64, token, refreshToken); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken,
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

	u, err := h.us.GetByRefreshToken(token)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if _, ok := h.checkBan(c, u.Id.Int64); ok {
		return
	}

	userRoles := h.rs.GetRoleByUserId(u.Id.Int64)

	access, refreshToken, err := h.ts.ReplaceTokens(token, jwtutil.GenerateClaims(u.Id.Int64, userRoles))
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
		AccessToken:  access.Token,
		RefreshToken: refreshToken.Token,
	})
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userId, err := strconv.ParseInt(c.GetHeader(headers.XAuthenticationUserHeader), 10, 64)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if err := h.ts.RemoveAllRefreshTokenUser(userId); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, nil)
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
