package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var Conn *pgx.Conn

func ConnectDB() {
	var err error

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in .env")
	}

	// PostgreSQL に接続
	Conn, err = pgx.Connect(context.Background(), databaseURL+"?statement_cache_mode=describe")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	log.Println("Connected to the database.")
}

func CloseDB() {
	if Conn != nil {
		if err := Conn.Close(context.Background()); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}
