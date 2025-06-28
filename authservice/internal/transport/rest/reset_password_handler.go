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
	ms      *services.MailService
	rp      *services.ResetPasswordService
	ls      *localizer.LocalizeService
	appInfo *config.AppInfo
}

func NewResetPasswordHandler(
	log *logrus.Entry,
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
	}
}

func (h *ResetPasswordHandler) SendCode(c *gin.Context) {
	var resPassDto dto.ResetPassword

	if msg, ok := validate.IsValidBody(c, &resPassDto, h.ls); !ok {
		responseutil.ErrorResponse(c, http.StatusBadRequest, errormsg.InvalidBody, commonDto.Message{
			Message: msg,
		})
	}
}
