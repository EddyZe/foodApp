package authapp

import (
	"github.com/EddyZe/foodApp/authservice/internal/app/storage"
	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/EddyZe/foodApp/authservice/internal/repositories"
	"github.com/EddyZe/foodApp/authservice/internal/server"
	"github.com/EddyZe/foodApp/authservice/internal/services"
	"github.com/EddyZe/foodApp/common/pkg/localizer"
	"github.com/sirupsen/logrus"
)

func MustRun(logger *logrus.Entry, appConf *config.AppConfig) {

	psql := storage.MustConnectPsql(logger, appConf.Postgres)
	red := storage.MustConnectRedis(logger, appConf.Redis)

	//инициализация зависимостей
	logger.Infoln("Создание репозиториев")
	ur := repositories.NewUserRepository(psql)
	rr := repositories.NewRoleRepository(psql)
	urr := repositories.NewUserRoleRepository(psql)
	br := repositories.NewBanRepository(psql)
	ar := repositories.NewAccessTokenRepository(psql)
	trr := repositories.NewRefreshTokenRepository(psql)
	evr := repositories.NewEmailVerificationCodeRepository(psql)
	evtr := repositories.NewEmailVerificationTokenRepository(psql)
	logger.Infoln("Репозитории созданы")

	logger.Infoln("Созание сервисов")
	rs := services.NewRoleService(logger, red, rr, urr)
	us := services.NewUserService(
		logger,
		red,
		rs,
		ur,
	)
	ts := services.NewTokenService(appConf.Tokens, red, trr, logger, ar)
	bs := services.NewBanService(logger, br, ts)
	ms := services.NewMailService(logger, appConf.SmptConfig)
	mvs := services.NewEmailVerificationCodeService(logger, appConf.EmailVerification, evr, evtr)
	lms := localizer.NewLocalizeService(logger, appConf.LocalizerConfig.DirFiles)
	logger.Infoln("Создане сервисов завершено")

	//Запуск сервера
	logger.Infoln("Запуск сервера")
	serv := server.New(logger, us, ts, rs, bs, ms, mvs, lms, appConf.AppInfo)
	if err := serv.ListenAndServe(); err != nil {
		panic(err)
	}

}
