package server

import (
	"net/http"
	"time"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/handlers"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/middleware"
	"github.com/gin-gonic/gin"
)

func New() *http.Server {
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

	auth := handlers.NewAuthHandler()

	router.GET("/ping", auth.Ping)

	logger.Infoln("Auth service starting. Port: ", port)
	return s
}
