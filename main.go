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
	dbUser := os.Getenv("CLOUD_SQL_USER")
	dbPassword := os.Getenv("CLOUD_SQL_PASSWORD")
	dbName := os.Getenv("DCLOUD_SQL_DATABASE")
	dbHost := os.Getenv("CLOUD_SQL_CONNECTION_NAME")

	a := app.App{}
	a.Initialize(
		dbHost,
		dbUser,
		dbPassword,
		dbName)

	a.Run(":8080")
}
