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
		os.Getenv("BOOKS_DB_HOST"),
		os.Getenv("BOOKS_DB_PORT"),
		os.Getenv("BOOKS_DB_USER"),
		os.Getenv("BOOKS_DB_PASSWORD"),
		os.Getenv("BOOKS_DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("failed to ping the db: ", err)
	}

	log.Println("books_db connected successfully")
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
		CREATE TABLE IF NOT EXISTS books (
		id			SERIAL PRIMARY KEY,
		title		TEXT NOT NULL,
		author		TEXT NOT NULL,
		price		NUMERIC(10, 2) NOT NULL,
		stock		INTEGER NOT NULL DEFAULT 0,
		created_at	TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("failed to create books table: ", err)
	}

	log.Println("books table ready")
}
