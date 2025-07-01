package rest

import (
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/domain/dto"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/internal/util/errormsg"
	commonDto "github.com/EddyZe/foodApp/common/domain/dto"
	"github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/EddyZe/foodApp/common/pkg/responseutil"
	"github.com/EddyZe/foodApp/common/pkg/validate"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ResetPasswordHandler struct {
	log     *logrus.Entry
	us      *services.UserService
	ms      *services.MailService
	rp      *services.ResetPasswordService
	ls      *localizer.LocalizeService
	appInfo *config.AppInfo
}

func NewResetPasswordHandler(
	log *logrus.Entry,
	us *services.UserService,
	ms *services.MailService,
	rp *services.ResetPasswordService,
	ls *localizer.LocalizeService,
	appInfo *config.AppInfo,
) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		log:     log,
		ms:      ms,
		rp:      rp,
		ls:      ls,
		appInfo: appInfo,
		us:      us,
	}
}

func (h *ResetPasswordHandler) SendCode(c *gin.Context) {
	var resPassDto dto.ResetPassword

	lang := c.GetHeader("Accept-Language")

	if msg, ok := validate.IsValidBody(c, &resPassDto, h.ls); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidBody, commonDto.Message{
			Message: msg,
		})
		return
	}

	user, ok := h.us.GetByEmail(resPassDto.Email)
	if !ok {
		msg := h.ls.GetMessage(
			localizer.UserNotFoundByEmail,
			lang,
			"User not found",
			map[string]interface{}{
				"email": resPassDto.Email,
			},
		)
		responseutil.ErrorResponse(c, http.StatusNotFound, errormsg.NotFound, commonDto.Message{
			Message: msg,
		})
		return
	}

	code, err := h.rp.GenerateAndSaveCode(user.Id.Int64)
	if err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	subject := h.ls.GetMessage(
		localizer.ResetPasswordSubject,
		lang,
		"Reset password",
		map[string]interface{}{
			"appName": h.appInfo.AppName,
		})

	letter := h.ls.GetMessage(
		localizer.ResetPasswordEmail,
		lang,
		"Enter code: "+code.Code,
		map[string]interface{}{
			"appName":        h.appInfo.AppName,
			"appSupportLink": h.appInfo.SupportLink,
			"code":           code.Code,
		})

	if err := h.ms.SendMailFromApp(subject, letter, resPassDto.Email); err != nil {
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, nil)
}

func (h *ResetPasswordHandler) EditPassword(c *gin.Context) {
	var enterCode dto.EnterCodeResetPassword

	if msg, ok := validate.IsValidBody(c, &enterCode, h.ls); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidBody, commonDto.Message{
			Message: msg,
		})
		return
	}

	lang := c.GetHeader("Accept-Language")

	code, err := h.rp.GetCode(enterCode.Code)
	if err != nil {
		msg := h.ls.GetMessage(
			localizer.InvalidResetPasswordCode,
			lang,
			"Invalid code",
			nil,
		)
		responseutil.ErrorResponse(c, http.StatusNotFound, errormsg.NotFound, commonDto.Message{
			Message: msg,
		})
		return
	}

	if code.IsExpired() || !code.IsValid {
		msg := h.ls.GetMessage(
			localizer.CodeExpired,
			lang,
			"Code expired",
			nil,
		)
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.CodeExpired, commonDto.Message{
			Message: msg,
		})
		return
	}

	if err := h.us.EditPassword(code.UserId, enterCode.NewPassword); err != nil {
		if err.Error() == errormsg.LastPasswordIsExists {
			msg := h.ls.GetMessage(
				localizer.LastPasswords,
				lang,
				"The new password should not be equal to the last two",
				nil,
			)
			responseutil.ErrorResponse(c, http.StatusBadRequest, err.Error(), commonDto.Message{
				Message: msg,
			})
			return
		}
		responseutil.ErrorResponse(c, http.StatusInternalServerError, errormsg.ServerInternalError)
		return
	}

	code.IsValid = false
	if err := h.rp.SetIsValid(code.Code, false); err != nil {
		h.log.Error("ошибка при обнавлении статуса кода: ", err)
		return
	}

	responseutil.SuccessResponse(c, http.StatusOK, nil)
}
