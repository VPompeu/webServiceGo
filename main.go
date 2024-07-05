package main

import (
	"github.com/VPompeu/agenda-astrologica/app"
	"github.com/joho/godotenv"

	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	a := app.App{}
	a.Initialize(
		dbUser,
		dbPassword,
		dbName)

	a.Run(":8080")
}
