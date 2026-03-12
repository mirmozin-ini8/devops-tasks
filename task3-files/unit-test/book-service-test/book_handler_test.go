package unit_test

import (
	"book-service/database"
	"book-service/handler"
	"book-service/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}
func TestGetAllBooksHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("ReturnsBooksAsJSON", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "song of ice and fire", "william", 29.99, 10)

		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books ORDER BY ID").
			WillReturnRows(rows)

		r := setupRouter()
		r.GET("/books", handler.GetAllBooks)

		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var books []model.Book
		err = json.Unmarshal(w.Body.Bytes(), &books)
		assert.NoError(t, err)
		assert.Len(t, books, 1)
		assert.Equal(t, "song of ice and fire", books[0].Title)
	})

	t.Run("ReturnsEmptyArray", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"})
		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books ORDER BY ID").
			WillReturnRows(rows)

		r := setupRouter()
		r.GET("/books", handler.GetAllBooks)

		req := httptest.NewRequest(http.MethodGet, "/books", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var books []model.Book
		err = json.Unmarshal(w.Body.Bytes(), &books)
		assert.NoError(t, err)
		assert.Len(t, books, 0)
	})
}

func TestGetBookByIDHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "song of ice and fire", "william", 29.99, 10)

		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		r := setupRouter()
		r.GET("/books/:id", handler.GetBookByID)

		req := httptest.NewRequest(http.MethodGet, "/books/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var book model.Book
		err = json.Unmarshal(w.Body.Bytes(), &book)
		assert.NoError(t, err)
		assert.Equal(t, 1, book.ID)
		assert.Equal(t, "song of ice and fire", book.Title)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}))

		r := setupRouter()
		r.GET("/books/:id", handler.GetBookByID)

		req := httptest.NewRequest(http.MethodGet, "/books/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		r := setupRouter()
		r.GET("/books/:id", handler.GetBookByID)

		req := httptest.NewRequest(http.MethodGet, "/books/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCreateBookHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "parasyte", "itadori", 34.99, 20)

		mock.ExpectQuery("INSERT INTO books").
			WithArgs("parasyte", "itadori", 34.99, 20).
			WillReturnRows(rows)

		r := setupRouter()
		r.POST("/books", handler.CreateBook)

		body := `{"title":"parasyte","author":"itadori","price":34.99,"stock":20}`
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var book model.Book
		err = json.Unmarshal(w.Body.Bytes(), &book)
		assert.NoError(t, err)
		assert.Equal(t, "parasyte", book.Title)
		assert.Equal(t, 34.99, book.Price)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		r := setupRouter()
		r.POST("/books", handler.CreateBook)

		body := `{"price":34.99}`
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidPrice", func(t *testing.T) {
		r := setupRouter()
		r.POST("/books", handler.CreateBook)

		body := `{"title":"Test","author":"Author","price":-5,"stock":10}`
		req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateBookHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		fetchRows := sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}).
			AddRow(1, "Old Title", "william", 29.99, 10)

		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(fetchRows)

		mock.ExpectExec("UPDATE books").
			WithArgs("new title", "william", 29.99, 10, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		r := setupRouter()
		r.PUT("/books/:id", handler.UpdateBook)

		body := `{"title":"new title"}`
		req := httptest.NewRequest(http.MethodPut, "/books/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var book model.Book
		err = json.Unmarshal(w.Body.Bytes(), &book)
		assert.NoError(t, err)
		assert.Equal(t, "new title", book.Title)
		assert.Equal(t, "william", book.Author)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, title, author, price, stock FROM books WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author", "price", "stock"}))

		r := setupRouter()
		r.PUT("/books/:id", handler.UpdateBook)

		body := `{"title":"anything"}`
		req := httptest.NewRequest(http.MethodPut, "/books/99", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		r := setupRouter()
		r.PUT("/books/:id", handler.UpdateBook)

		body := `{"title":"anything"}`
		req := httptest.NewRequest(http.MethodPut, "/books/xyz", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteBookHandler(t *testing.T) {
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

		r := setupRouter()
		r.DELETE("/books/:id", handler.DeleteBook)

		req := httptest.NewRequest(http.MethodDelete, "/books/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM books WHERE id = \\$1").
			WithArgs(99).
			WillReturnResult(sqlmock.NewResult(0, 0))

		r := setupRouter()
		r.DELETE("/books/:id", handler.DeleteBook)

		req := httptest.NewRequest(http.MethodDelete, "/books/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		r := setupRouter()
		r.DELETE("/books/:id", handler.DeleteBook)

		req := httptest.NewRequest(http.MethodDelete, "/books/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
