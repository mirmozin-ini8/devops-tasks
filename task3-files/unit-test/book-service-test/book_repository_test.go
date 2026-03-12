package unit_test

import (
	"book-service/database"
	"book-service/model"
	"book-service/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetAllBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}

	defer db.Close()
	database.DB = db

	t.Run("ReturnTwoBooks", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "wwe guide", "john cena", 29.99, 10).
			AddRow(2, "manga volumes", "akira", 39.99, 5)

		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books ORDER BY ID").WillReturnRows(rows)
		books, err := repository.GetAllBooks()

		assert.NoError(t, err)
		assert.Len(t, books, 2)
		assert.Equal(t, "wwe guide", books[0].Title)
		assert.Equal(t, 1, books[0].ID)
		assert.Equal(t, "manga volumes", books[1].Title)
	})

	t.Run("ReturnsEmptySlice", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"})
		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books ORDER BY ID").WillReturnRows(rows)

		books, err := repository.GetAllBooks()
		assert.NoError(t, err)
		assert.NotNil(t, books)
		assert.Len(t, books, 0)
	})
}

func TestGetBookByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}

	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "wwe guide", "john cena", 29.99, 10)

		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		book, err := repository.GetBookByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, 1, book.ID)
		assert.Equal(t, "wwe guide", book.Title)
		assert.Equal(t, 29.99, book.Price)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}))

		book, err := repository.GetBookByID(99)

		assert.NoError(t, err)
		assert.Nil(t, book)
	})
}

func TestCreateBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}

	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		req := model.CreateBookRequest{
			Title:  "jujutsu kaisen",
			Author: "akira",
			Price:  34.99,
			Stock:  20,
		}

		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(3, "jujutsu kaisen", "akira", 34.99, 20)

		mock.ExpectQuery("INSERT INTO books").
			WithArgs(req.Title, req.Author, req.Price, req.Stock).
			WillReturnRows(rows)

		book, err := repository.CreateBook(req)
		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, 3, book.ID)
		assert.Equal(t, "jujutsu kaisen", book.Title)
		assert.Equal(t, 34.99, book.Price)
	})
}

func TestUpdateBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("UpdateTitle", func(t *testing.T) {
		fetchRows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "old title", "mir", 29.99, 10)

		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(fetchRows)

		mock.ExpectExec("UPDATE books").
			WithArgs("new title", "mir", 29.99, 10, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		newTitle := "new title"
		req := model.UpdateBookRequest{Title: &newTitle}

		book, err := repository.UpdateBook(1, req)
		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, "new title", book.Title)
		assert.Equal(t, "mir", book.Author)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}))

		newTitle := "anything"
		req := model.UpdateBookRequest{Title: &newTitle}

		book, err := repository.UpdateBook(99, req)
		assert.NoError(t, err)
		assert.Nil(t, book)
	})
}

func TestDeleteBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))

		deleted, err := repository.DeleteBook(1)
		assert.NoError(t, err)
		assert.True(t, deleted)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM books WHERE id = \\$1").
			WithArgs(99).
			WillReturnResult(sqlmock.NewResult(0, 0))

		deleted, err := repository.DeleteBook(99)
		assert.NoError(t, err)
		assert.False(t, deleted)
	})
}
