package handlers

import (
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/entity"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/EddyZe/foodApp/authservice/internal/util/stringutils"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/dto"
	"github.com/EddyZe/foodApp/common/dto/auth"
	"github.com/EddyZe/foodApp/common/models"
	"github.com/EddyZe/foodApp/common/pkg/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/EddyZe/foodApp/common/pkg/validate"
	"github.com/EddyZe/foodApp/common/services/localizer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type AuthHandler struct {
	us           *services.UserService
	ts           *services.TokenService
	rs           *services.RoleService
	log          *logrus.Entry
	bs           *services.BanService
	sendMailServ *services.MailService
	mvs          *services.EmailVerificationCodeService
	lms          *localizer.LocalizeService
	appInfo      *config.AppInfo
}

func NewAuthHandler(
	log *logrus.Entry,
	us *services.UserService,
	ts *services.TokenService,
	rs *services.RoleService,
	bs *services.BanService,
	sendMailServ *services.MailService,
	mvs *services.EmailVerificationCodeService,
	lms *localizer.LocalizeService,
	appInfo *config.AppInfo,
) *AuthHandler {
	return &AuthHandler{
		us:           us,
		log:          log,
		ts:           ts,
		rs:           rs,
		bs:           bs,
		sendMailServ: sendMailServ,
		mvs:          mvs,
		lms:          lms,
		appInfo:      appInfo,
	}
}

// Ping ...
func (h *AuthHandler) Ping(c *gin.Context) {
	c.String(200, "pong")
}

// Registry регистрация пользователя
func (h *AuthHandler) Registry(c *gin.Context) {
	var registerDto auth.RegisterDto

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

	responseutil.SuccessResponse(c, http.StatusCreated, &auth.TokensDto{
		AccessToken:  token,
		RefreshToken: refreshToken,
	})
}

// Login авторизирует пользователя
func (h *AuthHandler) Login(c *gin.Context) {
	var loginDto auth.LoginDto
	lang := c.GetHeader("Accept-Language")

	if msg, ok := validate.IsValidBody(c, &loginDto, h.lms); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, msg)
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

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
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

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
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

// SendMailConfirmCode отправляет письмо для подтверждения почты
func (h *AuthHandler) SendMailConfirmCode(c *gin.Context) {
	lang := c.GetHeader("Accept-Language")

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

	u, err := h.us.GetById(userId)
	if err != nil {
		if err.Error() == errormsg.NotFound {
			responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
			return
		}

		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	code, err := h.mvs.GenerateAndSaveCode(u.Id.Int64, 8)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	body := h.lms.GetMessage(
		localizer.SendVerifiedCodeBody,
		lang,
		fmt.Sprintf("\"Send this is code: %s", code),
		map[string]interface{}{
			"appName":        h.appInfo.AppName,
			"url":            "test@test.com",
			"appSupportLink": h.appInfo.SupportLink,
			"code":           code,
		},
	)

	subject := h.lms.GetMessage(
		localizer.SendVerifiedCodeSubject,
		lang,
		"Confirm Your Email",
		map[string]interface{}{
			"appName": h.appInfo.AppName,
		},
	)

	if err := h.sendMailServ.SendMailFromApp(
		subject,
		body,
		u.Email,
	); err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, nil)
}

// TODO НАСТРОИТЬ ПОДТВЕРЖДЕНИЕ ПО ССЫЛКЕ (ГЕНЕРАЦИЯ ССЫЛКИ)
func (h *AuthHandler) ConfirmMail(c *gin.Context) {
	lang := c.GetHeader("Accept-Language")
	token, ok := jwtutil.ExtractBearerTokenHeader(c)
	if !ok {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	claims, ok := c.Get("claims")
	if !ok {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	claimsMap, ok := claims.(*models.JwtClaims)
	if !ok {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	var ce auth.ConfirmEmail
	if msg, ok := validate.IsValidBody(c, &ce, h.lms); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidBody, dto.Message{
			Message: msg,
		})
		return
	}

	code, ok := h.mvs.FindCode(ce.Code)
	if !ok || code.Code != ce.Code || !code.IsVerified {
		msg := h.lms.GetMessage(
			localizer.InvalidEmailCode,
			lang,
			"Invalid code. Please check the entered code.",
			nil,
		)
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidEmailCode, dto.Message{
			Message: msg,
		})
		return
	}

	if code.IsExpired() {
		msg := h.lms.GetMessage(
			localizer.ExpiredEmailCode,
			lang,
			"The code has expired!",
			nil,
		)
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidEmailCode, dto.Message{
			Message: msg,
		})
		return
	}

	updateUser, err := h.us.SetEmailConfirmed(claimsMap.Sub, true)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}
	code.IsVerified = false
	if err := h.mvs.SetVerified(code.Code, code.IsVerified); err != nil {
		h.log.Error("ошибка при замене статуса email code: ", err)
	}
	if err := h.ts.Logout(token); err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	userRoles := h.rs.GetRoleByUserId(updateUser.Id.Int64)
	accessTok, err := h.ts.GenerateJwtByUser(updateUser, userRoles)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}
	refreshToken := h.ts.GenerateUUID()
	if _, _, err := h.ts.SaveRefreshAndAccessToken(updateUser.Id.Int64, accessTok, refreshToken); err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, &auth.TokensDto{
		AccessToken:  accessTok,
		RefreshToken: refreshToken,
	})
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
	responseutil.ErrorResponse(
		c,
		http.StatusForbidden,
		errormsg.AccountIsBlocked,
		dto.Message{
			Message: msg,
		})
}
