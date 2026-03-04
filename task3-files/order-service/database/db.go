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
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("ORDERS_DB_HOST"), os.Getenv("ORDERS_DB_PORT"), os.Getenv("ORDERS_DB_USER"), os.Getenv("ORDERS_DB_PASSWORD"), os.Getenv("ORDERS_DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("failed to open db connection: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("failed to ping the db: ", err)
	}

	log.Println("orders_db connected successfully")
	createTable()
}

func createTable() {
	query := `
		CREATE TABLE IF NOT EXISTS orders ( 
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			book_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL CHECK (quantity > 0),
			total_price NUMERIC(10,2) NOT NULL,
			status TEXT NOT NULL DEFAULT 'confirmed',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("failed to create orders table: ", err)
	}
	log.Println("orders table ready")
}
