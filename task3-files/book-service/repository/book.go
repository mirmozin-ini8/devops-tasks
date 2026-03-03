package repository

import (
	"book-service/database"
	"book-service/model"
	"database/sql"
	"errors"
)

func GetAllBooks() ([]model.Book, error) {
	rows, err := database.DB.Query(
		"SELECT id, title, author, price, stock FROM books ORDER BY ID",
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var books []model.Book
	for rows.Next() {
		var b model.Book
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	if books == nil {
		books = []model.Book{}
	}

	return books, nil
}

func GetBookByID(id int) (*model.Book, error) {
	var b model.Book
	err := database.DB.QueryRow(
		"SELECT id, title, author, price, stock FROM books WHERE id = $1", id,
	).Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func CreateBook(req model.CreateBookRequest) (*model.Book, error) {
	var b model.Book
	err := database.DB.QueryRow(
		`INSERT INTO books (title, author, price, stock)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, author, price, stock`,
		req.Title, req.Author, req.Price, req.Stock,
	).Scan(&b.ID, &b.Title, &b.Author, &b.Price, &b.Stock)

	if err != nil {
		return nil, err
	}
	return &b, nil
}

func UpdateBook(id int, req model.UpdateBookRequest) (*model.Book, error) {
	existing, err := GetBookByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Author != nil {
		existing.Author = *req.Author
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.Stock != nil {
		existing.Stock = *req.Stock
	}

	_, err = database.DB.Exec(
		`UPDATE books SET title=$1, author=$2, price=$3, stock=$4 WHERE id=$5`,
		existing.Title, existing.Author, existing.Price, existing.Stock, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func DeleteBook(id int) (bool, error) {
	result, err := database.DB.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}
