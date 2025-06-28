package server

import (
	"github.com/EddyZe/foodApp/authservice/internal/transport/rest"
	commonService "github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(
	logger *logrus.Entry,
	us *services.UserService,
	ts *services.TokenService,
	rs *services.RoleService,
	bs *services.BanService,
	ms *services.MailService,
	mvs *services.EmailVerificationService,
	lms *commonService.LocalizeService,
	rp *services.ResetPasswordService,
	appInfo *config.AppInfo,
) *http.Server {
	router := gin.New()

	port := config.GetPort()

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	router.Use(middleware.Logger(logger))
	router.Use(gin.Recovery())

	auth := rest.NewAuthHandler(logger, us, ts, rs, bs, lms)
	emailVerificationHandler := rest.NewEmailVerificationHandler(us, ts, rs, logger, ms, mvs, lms, appInfo)
	resetPasswordHandler := rest.NewResetPasswordHandler(logger, ms, rp, lms, appInfo)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/ping", auth.Ping)

	apiV1 := router.Group("/api/v1")
	apiV1.POST("/sing-up", auth.Registry)
	apiV1.POST("/login", auth.Login)
	apiV1.POST("/refresh", auth.Refresh)
	apiV1.POST("/logout-all", middleware.JwtFilter(ts.Secret()), auth.LogoutAll)
	apiV1.POST("/logout", middleware.JwtFilter(ts.Secret()), auth.Logout)
	apiV1.POST("/ban", middleware.JwtFilter(ts.Secret()), middleware.IsAdmin(lms), auth.BanUser)
	apiV1.POST("/unban", middleware.JwtFilter(ts.Secret()), middleware.IsAdmin(lms), auth.UnBanUser)

	apiV1.POST("/email-code", middleware.JwtFilter(ts.Secret()), emailVerificationHandler.SendMailConfirmCode)
	apiV1.POST("/confirm-email", middleware.JwtFilter(ts.Secret()), emailVerificationHandler.ConfirmMail)
	apiV1.GET("/confirm-email-url", emailVerificationHandler.ConfirmEmailByUrl)

	apiV1.POST("/code", resetPasswordHandler.SendCode)

	logger.Infoln("Auth service starting. Port: ", port)
	return s
}
