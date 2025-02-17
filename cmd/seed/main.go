package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/mochaeng/sapphire-backend/internal/database"
	"github.com/mochaeng/sapphire-backend/internal/env"
	"github.com/mochaeng/sapphire-backend/internal/store/postgres"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}

	addr := env.GetString(
		"DATABASE_URL",
		"postgres://hutao:adminpassword@localhost:8888/sapphire?sslmode=disable",
	)
	conn, err := database.NewConnection(addr, 1, 1, 900)
	if err != nil {
		log.Printf("error while creating connection to seed the database: %s", err)
		return
	}
	defer conn.Close()

	store := postgres.NewPostgresStore(conn)
	database.Seed(store, conn)
}
