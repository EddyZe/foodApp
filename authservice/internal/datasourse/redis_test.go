package datasourse

import (
	"encoding/json"
	"github.com/EddyZe/foodApp/authservice/internal/datasourse/redis"
	"log"
	"testing"

	"github.com/EddyZe/foodApp/authservice/internal/config"
	"github.com/go-playground/assert/v2"
	"github.com/joho/godotenv"
)

func loadAppCfg() *config.AppConfig {
	err := godotenv.Load("./../../../.env")
	if err != nil {
		log.Println(err)
	}
	return config.Config()
}

func TestSave(*testing.T) {
	cfg := loadAppCfg()
	log.Println(cfg.Redis.Host)
	red, err := redis.Connect(cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	if err := red.Put("key", cfg); err != nil {
		log.Fatal(err)
	}
}

func TestGet(*testing.T) {
	cfg := loadAppCfg()
	log.Println(cfg.Redis.Host)
	red, err := redis.Connect(cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	val, ok := red.Get("key")
	if !ok {
		log.Fatal(err)
	}
	log.Println(val)
	var cfg2 config.AppConfig
	if err := json.Unmarshal([]byte(val), &cfg2); err != nil {
		log.Fatal()
	}
	assert.IsEqual(cfg2.Redis.Host, cfg.Redis.Host)
}

func TestDel(*testing.T) {
	cfg := loadAppCfg()
	log.Println(cfg.Redis.Host)
	red, err := redis.Connect(cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	if err := red.Del("key"); err != nil {
		log.Fatal(err)
	}
	v, _ := red.Get("key")
	assert.IsEqual(v, "")
}
