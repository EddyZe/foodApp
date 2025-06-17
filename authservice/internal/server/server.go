package server

import (
	"net/http"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/handlers"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(
	us *services.UserService,
	ts *services.TokenService,
	rs *services.RoleService,
	bs *services.BanService,
	ms *services.MailService,
) *http.Server {
	logger := pkg.InitLogger("Auth-Service [SERVER]")
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

	auth := handlers.NewAuthHandler(logger, us, ts, rs, bs)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/ping", auth.Ping)

	apiV1 := router.Group("/api/v1")
	apiV1.POST("/sing-up", auth.Registry)
	apiV1.POST("/login", auth.Login)
	apiV1.POST("/refresh", auth.Refresh)
	apiV1.POST("/logout-all", auth.LogoutAll)
	apiV1.POST("/logout", auth.Logout)

	logger.Infoln("Auth service starting. Port: ", port)
	return s
}
