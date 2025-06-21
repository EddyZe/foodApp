package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type YamlConfig struct {
	Env       string     `yaml:"env"`
	EnvFile   string     `yaml:"envfile" env-default:".env"`
	LoggerCfg *LoggerCfg `yaml:"logger"`
}
type LoggerCfg struct {
	Level    string `yaml:"level" env-default:"info"`
	AppEnv   string `yaml:"app-env" env-default:"development"`
	Service  string `yaml:"service" env-default:"authservice"`
	FilePath string `yaml:"file-path" env-default:"./logs"`
}

func MustLoad() *YamlConfig {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file does not exist")
	}

	var cfg YamlConfig

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}
