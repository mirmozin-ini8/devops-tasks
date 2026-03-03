package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("USERS_DB_HOST"),
		os.Getenv("USERS_DB_PORT"),
		os.Getenv("USERS_DB_USER"),
		os.Getenv("USERS_DB_PASSWORD"),
		os.Getenv("USERS_DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("failed to ping the db", err)
	}

	log.Println("users_db connected successfully")
	createTable()
}

func Ping() {
	err := DB.Ping()
	if err != nil {
		log.Fatal("failed to ping the db: ", err)
	}
}

func createTable() {
	query := `
		CREATE TABLE IF NOT EXISTS users (
		id				SERIAL PRIMARY KEY,
		username		TEXT NOT NULL UNIQUE,
		email			TEXT NOT NULL UNIQUE,
		password_hash	TEXT NOT NULL,
		created_at		TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("failed to create users table: ", err)
	}

	log.Println("users table ready")
}
