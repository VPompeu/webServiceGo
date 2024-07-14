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
	//dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	a := app.App{}
	a.Initialize(
		dbUser,
		dbPassword,
		dbName,
		dbPort)

	a.Run(":8080")
}
