package main

import (
	"log"

	"github.com/EddyZe/foodApp/authservice/pkg"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("./../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	logger := pkg.InitLogger()

	logger.Info("Starting Auth Service...")

}
