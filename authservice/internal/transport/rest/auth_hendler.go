package rest

import (
	dto2 "github.com/EddyZe/foodApp/authservice/internal/domain/dto"
	"github.com/EddyZe/foodApp/authservice/internal/domain/entity"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/EddyZe/foodApp/authservice/internal/util/passencoder"
	"github.com/EddyZe/foodApp/authservice/internal/util/stringutils"
	"github.com/EddyZe/foodApp/common/domain/dto"
	"github.com/EddyZe/foodApp/common/domain/models"
	"github.com/EddyZe/foodApp/common/pkg/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/EddyZe/foodApp/common/pkg/validate"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type AuthHandler struct {
	us  *services.UserService
	ts  *services.TokenService
	rs  *services.RoleService
	log *logrus.Entry
	bs  *services.BanService
	lms *localizer.LocalizeService
}

func NewAuthHandler(
	log *logrus.Entry,
	us *services.UserService,
	ts *services.TokenService,
	rs *services.RoleService,
	bs *services.BanService,
	lms *localizer.LocalizeService,
) *AuthHandler {
	return &AuthHandler{
		us:  us,
		log: log,
		ts:  ts,
		rs:  rs,
		bs:  bs,
		lms: lms,
	}
}

// Ping ...
func (h *AuthHandler) Ping(c *gin.Context) {
	c.String(200, "pong")
}

// Registry регистрация пользователя
func (h *AuthHandler) Registry(c *gin.Context) {
	var registerDto dto2.RegisterDto

	if msg, ok := validate.IsValidBody(c, &registerDto, h.lms); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, msg)
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

	userRoles := h.rs.GetRoleByUserId(user.Id.Int64)
	token, err := h.ts.GenerateJwt(jwtutil.GenerateClaims(&models.JwtClaims{
		Email:         user.Email,
		EmailVerified: user.EmailIsConfirm,
		Role:          stringutils.RoleMapString(userRoles),
		Sub:           user.Id.Int64,
	}))
	if err != nil {
		h.log.Error(err)
	}

	refreshToken := h.ts.GenerateUUID()

	if _, _, err := h.ts.SaveRefreshAndAccessToken(user.Id.Int64, token, refreshToken); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusCreated, &dto2.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

// Login авторизирует пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var loginDto dto2.LoginDto
	lang := c.GetHeader("Accept-Language")

	if msg, ok := validate.IsValidBody(c, &loginDto, h.lms); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, msg)
		return
	}

	u, ok := h.us.GetByEmail(loginDto.Email)

	if !ok || !passencoder.CheckEqualsPassword(loginDto.Password, u.Password) {
		responseutil.ErrorResponse(
			c,
			http.StatusBadRequest,
			errormsg.InvalidEmailOrPassword,
		)
		return
	}

	if ban, ok := h.isBan(u.Id.Int64); ok {
		h.banResponse(c, ban, lang)
		return
	}

	userRoles := h.rs.GetRoleByUserId(u.Id.Int64)

	token, err := h.ts.GenerateJwtByUser(u, userRoles)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	refreshToken := h.ts.GenerateUUID()

	if _, _, err := h.ts.SaveRefreshAndAccessToken(u.Id.Int64, token, refreshToken); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, &dto2.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

// Refresh заменяет авторизационные токены
func (h *AuthHandler) Refresh(c *gin.Context) {
	lang := c.GetHeader("Accept-Language")
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	if token == "" {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	if !h.ts.ValidateRefreshToken(token) {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	u, err := h.us.GetByRefreshToken(token)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	if ban, ok := h.isBan(u.Id.Int64); ok {
		h.banResponse(c, ban, lang)
		return
	}

	userRoles := h.rs.GetRoleByUserId(u.Id.Int64)

	access, refreshToken, err := h.ts.ReplaceTokens(token, jwtutil.GenerateClaims(&models.JwtClaims{
		Email:         u.Email,
		EmailVerified: u.EmailIsConfirm,
		Role:          stringutils.RoleMapString(userRoles),
		Sub:           u.Id.Int64,
	}))
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, &dto2.TokensDto{
		AccessToken:  access.Token,
		RefreshToken: refreshToken.Token,
	})
}

// Logout удаляет токены
func (h *AuthHandler) Logout(c *gin.Context) {
	token, ok := jwtutil.ExtractBearerTokenHeader(c)
	if !ok {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	if err := h.ts.Logout(token); err != nil {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, nil)
}

// LogoutAll удаляет все токены пользователя
func (h *AuthHandler) LogoutAll(c *gin.Context) {

	claims, ok := c.Get("claims")
	if !ok {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	claimsMap, ok := claims.(*models.JwtClaims)
	if !ok {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
	}

	userId := claimsMap.Sub

	if err := h.ts.LogoutAll(userId); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, nil)
}

// isBan проверка блокировки пользователя
func (h *AuthHandler) isBan(userId int64) (*entity.Ban, bool) {
	if ban, ok := h.bs.GetActiveUserBan(userId); ok {
		return ban, true
	}
	return nil, false
}

// banResponse отправляет сообщение с ответом, что пользователь заблокирован
func (h *AuthHandler) banResponse(c *gin.Context, ban *entity.Ban, lang string) {
	msg := h.getMsgToBan(ban, lang)
	responseutil.ErrorResponse(
		c,
		http.StatusForbidden,
		errormsg.AccountIsBlocked,
		dto.Message{
			Message: msg,
		})
}

func (h *AuthHandler) BanUser(c *gin.Context) {
	var userBan *dto2.BanUser

	if msg, ok := validate.IsValidBody(c, &userBan, h.lms); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidBody, msg)
		return
	}
	if userBan.Days == 0 {
		userBan.Days = 5
	}

	if ban, ok := h.isBan(userBan.UserId); ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.UserIsAlreadyBlocked, ban)
		return
	}

	if userBan.IsForever {
		ban, err := h.bs.BanUserForever(userBan.UserId, userBan.Cause)
		if err != nil {
			responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
			return
		}
		responseutil.SuccessResponse(c, http.StatusOK, ban)
		return
	}

	expiredAt := time.Now().Add(time.Duration(userBan.Days) * 24 * time.Hour)

	ban, err := h.bs.BanUser(userBan.UserId, userBan.Cause, expiredAt)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, ban)
}

func (h *AuthHandler) getMsgToBan(ban *entity.Ban, lang string) string {
	var expired string
	if ban.IsForever {
		expired = h.lms.GetMessage(
			localizer.AccountBanForever,
			lang,
			"forever",
			nil)
	} else {
		expired = ban.ExpiredAt.Format("02-01-2006 15:04:05")
	}

	msg := h.lms.GetMessage(
		localizer.AccountIsBlocked,
		lang,
		"The account is blocked",
		map[string]interface{}{
			"banExpired": expired,
		})

	return msg
}
