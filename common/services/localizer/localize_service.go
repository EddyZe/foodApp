package localizer

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"strings"
)

type LocalizeService struct {
	log    *logrus.Entry
	bundle *i18n.Bundle
}

func NewLocalizeService(log *logrus.Entry, localizeDir string) *LocalizeService {

	bundle := i18n.NewBundle(language.Russian)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	files, err := os.ReadDir(localizeDir)
	if err != nil {
		log.Debug("Директория с переводами не найдена. Будут применены стандартные сообщения!")
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".toml") {
			_, err := bundle.LoadMessageFile(filepath.Join(localizeDir, f.Name()))
			if err != nil {
				continue
			}
		}
	}
	return &LocalizeService{
		log:    log,
		bundle: bundle,
	}
}

// GetMessage находит переыод сообщения
func (s *LocalizeService) GetMessage(
	idTranslate, lang, defaultMessage string,
	templateData map[string]interface{},
) string {
	localizer := i18n.NewLocalizer(s.bundle, lang)
	res, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    idTranslate,
			Other: defaultMessage,
		},
		TemplateData: templateData,
	})

	if err != nil {
		s.log.Error("ошибка при локализации сообщения: ", err)
		return defaultMessage
	}

	return res
}
