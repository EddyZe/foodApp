package rest

import (
	"fmt"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	dto2 "github.com/EddyZe/foodApp/authservice/internal/domain/dto"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	"github.com/EddyZe/foodApp/common/domain/dto"
	"github.com/EddyZe/foodApp/common/domain/models"
	"github.com/EddyZe/foodApp/common/pkg/jwtutil"
	"github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/EddyZe/foodApp/common/pkg/validate"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type EmailVerificationHandler struct {
	us           *services.UserService
	ts           *services.TokenService
	rs           *services.RoleService
	log          *logrus.Entry
	sendMailServ *services.MailService
	mvs          *services.EmailVerificationService
	lms          *localizer.LocalizeService
	appInfo      *config.AppInfo
}

func NewEmailVerificationHandler(
	us *services.UserService,
	ts *services.TokenService,
	rs *services.RoleService,
	log *logrus.Entry,
	sendMailServ *services.MailService,
	mvs *services.EmailVerificationService,
	lms *localizer.LocalizeService,
	appInfo *config.AppInfo,
) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		us:           us,
		ts:           ts,
		rs:           rs,
		log:          log,
		sendMailServ: sendMailServ,
		mvs:          mvs,
		lms:          lms,
		appInfo:      appInfo,
	}
}

// SendMailConfirmCode отправляет письмо для подтверждения почты
func (h *EmailVerificationHandler) SendMailConfirmCode(c *gin.Context) {
	lang := c.GetHeader("Accept-Language")

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

	if claimsMap.EmailVerified {
		h.responseEmailConfirmed(c, lang)
		return
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

	urlToken := h.ts.GenerateUUID()
	if err := h.mvs.SaveVerificationToken(code.Id.Int64, urlToken); err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		if err := h.mvs.Delete(code.Code); err != nil {
			h.log.Error("ошибка удаления сегерированного кода")
		}
		return
	}
	url := fmt.Sprintf("%s/confirm-email-url?token=%s&code=%s", h.appInfo.AppUrl, urlToken, code.Code)

	body := h.lms.GetMessage(
		localizer.SendVerifiedCodeBody,
		lang,
		fmt.Sprintf("Send this is code: %s", code.Code),
		map[string]interface{}{
			"appName":        h.appInfo.AppName,
			"url":            url,
			"appSupportLink": h.appInfo.SupportLink,
			"code":           code.Code,
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

// ConfirmEmailByUrl подтверждение почты по ссылке из письма
func (h *EmailVerificationHandler) ConfirmEmailByUrl(c *gin.Context) {
	lang := c.GetHeader("Accept-language")
	if lang == "" {
		lang = "en"
	}
	tokenString := c.Query("token")
	if tokenString == "" {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}
	codeString := c.Query("code")
	if codeString == "" {
		msg := h.lms.GetMessage(
			localizer.InvalidEmailCode,
			lang,
			"Invalid code. Please check the entered code.",
			nil,
		)
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidEmailCode, dto.Message{
			Message: msg,
		})
	}

	token, ok := h.mvs.GetToken(tokenString)
	if !ok || token.IsExpired() || !token.IsActive {
		responseutil.ErrorResponse(c, http.StatusUnauthorized, errormsg.Unauthorized)
		return
	}

	code, ok := h.mvs.GetByEmailVerifToken(tokenString)
	if !ok || code.Code != codeString || !code.IsVerified {
		h.responseInvalidCode(c, lang)
		return
	}

	if code.IsExpired() {
		h.responseExpiredCode(c, lang)
		return
	}

	updateUser, err := h.us.SetEmailConfirmed(code.UserId, true)
	if err != nil {
		h.log.Error(err)
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	if err := h.mvs.Delete(codeString); err != nil {
		h.log.Error(err)
		code.IsVerified = false
		if err := h.mvs.SetVerifiedCode(code.Code, code.IsVerified); err != nil {
			h.log.Error(err)
		}
		token.IsActive = false
		if err := h.mvs.SetIsActiveToken(tokenString, token.IsActive); err != nil {
			h.log.Error(err)
		}
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

	responseutil.SuccessResponse(c, http.StatusOK, &dto2.TokensDto{
		AccessToken:  accessTok,
		RefreshToken: refreshToken,
	})
}

// ConfirmMail подтверждение почты по коду
func (h *EmailVerificationHandler) ConfirmMail(c *gin.Context) {
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

	if claimsMap.EmailVerified {
		h.responseEmailConfirmed(c, lang)
		return
	}

	var ce dto2.ConfirmEmail
	if msg, ok := validate.IsValidBody(c, &ce, h.lms); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidBody, dto.Message{
			Message: msg,
		})
		return
	}

	code, ok := h.mvs.FindCode(ce.Code)
	if !ok || code.Code != ce.Code || !code.IsVerified {
		h.responseInvalidCode(c, lang)
		return
	}

	if code.IsExpired() {
		h.responseExpiredCode(c, lang)
		return
	}

	updateUser, err := h.us.SetEmailConfirmed(claimsMap.Sub, true)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	if err := h.mvs.Delete(code.Code); err != nil {
		h.log.Error(err)
		code.IsVerified = false
		if err := h.mvs.SetVerifiedCode(code.Code, code.IsVerified); err != nil {
			h.log.Error("ошибка при замене статуса email code: ", err)
		}
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

	responseutil.SuccessResponse(c, http.StatusOK, &dto2.TokensDto{
		AccessToken:  accessTok,
		RefreshToken: refreshToken,
	})
}

func (h *EmailVerificationHandler) responseInvalidCode(c *gin.Context, lang string) {
	msg := h.lms.GetMessage(
		localizer.InvalidEmailCode,
		lang,
		"Invalid code. Please check the entered code.",
		nil,
	)
	responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidEmailCode, dto.Message{
		Message: msg,
	})
}

func (h *EmailVerificationHandler) responseExpiredCode(c *gin.Context, lang string) {
	msg := h.lms.GetMessage(
		localizer.ExpiredEmailCode,
		lang,
		"The code has expired!",
		nil,
	)
	responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidEmailCode, dto.Message{
		Message: msg,
	})
}

func (h *EmailVerificationHandler) responseEmailConfirmed(c *gin.Context, lang string) {
	msg := h.lms.GetMessage(
		localizer.EmailConfirm,
		lang,
		"Email is confirmed",
		nil,
	)
	responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.EmailIsConfirmed, &dto.Message{
		Message: msg,
	})
}
