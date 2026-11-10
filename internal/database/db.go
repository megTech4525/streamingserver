package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func Connect() (*sqlx.DB, error) {
	_ = godotenv.Load()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE"),
		// os.Getenv("DB_HOST"),
	)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Println("ping database failed", err)
		return nil, err
	}
	fmt.Println("Successfully connected to the database!")

	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(25)
	log.Println("the database is connected", os.Getenv("DB_NAME"))

	return db, nil

}
