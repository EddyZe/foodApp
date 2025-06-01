package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load("./../.env"); err != nil {
		log.Println("Error loading .env file")
	}
}

func GetPort() string {
	port := os.Getenv("AUTH_SERVER_PORT")
	if port == "" {
		port = "8081"
	}
	return port
}
