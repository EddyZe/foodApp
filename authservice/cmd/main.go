package main

import (
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/server"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/EddyZe/foodApp/common/services/localizer"
	"github.com/sirupsen/logrus"
	"log"
)

func main() {
	config.LoadEnv()
	logger := pkg.InitLogger("Auth-Service[MAIN] - ")

	logger.Infoln("Starting Auth Service...")
	//Загрузка конфигурации приложения
	appConf := config.Config(logger)

	//Подключение к БД
	postgre, red, err := connectionDbs(logger, appConf)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
		return
	}
	defer func() {
		if err := postgre.Close(); err != nil {
			logger.Errorf("Error closing connection to database: %v", err)
		}
	}()
	defer func() {
		if err := red.Close(); err != nil {
			logger.Errorf("Error closing connection to database: %v", err)
		}
	}()

	//инициализация зависимостей
	logger.Infoln("Создание репозиториев")
	ur := repositories.NewUserRepository(postgre)
	rr := repositories.NewRoleRepository(postgre)
	urr := repositories.NewUserRoleRepository(postgre)
	br := repositories.NewBanRepository(postgre)
	ar := repositories.NewAccessTokenRepository(postgre)
	trr := repositories.NewRefreshTokenRepository(postgre)
	evr := repositories.NewEmailVerificationCodeRepository(postgre)
	evtr := repositories.NewEmailVerificationTokenRepository(postgre)
	logger.Infoln("Репозитории созданы")

	logger.Infoln("Созание сервисов")
	rs := services.NewRoleService(logger, red, rr, urr)
	bs := services.NewBanService(logger, br)
	us := services.NewUserService(
		logger,
		red,
		rs,
		ur,
	)
	ts := services.NewTokenService(appConf.Tokens, red, trr, logger, ar)
	ms := services.NewMailService(logger, appConf.SmptConfig)
	mvs := services.NewEmailVerificationCodeService(logger, appConf.EmailVerification, evr, evtr)
	lms := localizer.NewLocalizeService(logger, appConf.LocalizerConfig.DirFiles)
	logger.Infoln("Создане сервисов завершено")

	//Запуск сервера
	logger.Infoln("Запуск сервера")
	serv := server.New(us, ts, rs, bs, ms, mvs, lms, appConf.AppInfo)
	if err := serv.ListenAndServe(); err != nil {
		log.Fatalf("Error starting Auth Service: %v", err)
	}
}

func connectionDbs(logger *logrus.Entry, appConf *config.AppConfig) (postgres *datasourse.PostgresDb, redis *datasourse.Redis, err error) {
	// подключение к постгес
	logger.Infoln("Connecting to database...")
	db, err := datasourse.ConnectionPostgres(appConf.Postgres)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
		return nil, nil, err
	}
	logger.Infoln("Database connected")

	//подключение к редис
	logger.Infoln("Connecting to redis")
	r, err := datasourse.ConnectionRedis(appConf.Redis)
	if err != nil {
		logger.Errorf("Error connecting to redis: %v", err)
		return nil, nil, err
	}
	logger.Infoln("Redis connected")
	return db, r, nil
}
