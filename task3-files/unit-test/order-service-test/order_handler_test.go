package unit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"order-service/database"
	"order-service/handler"
	"order-service/middleware"
	"order-service/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func generateTestToken(userID int) string {
	os.Setenv("SECRET_KEY", "test-secret-key")
	claims := &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("test-secret-key"))
	return signed
}

func startFakeBookServer(book *model.BookResponse, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
	}))
}

func startFakeUserServer(user *model.UserResponse, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}))
}

func TestCreateOrderHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		userServer := startFakeUserServer(&model.UserResponse{ID: 1, Username: "mir", Email: "mir@test.com"}, http.StatusOK)
		defer userServer.Close()
		os.Setenv("USER_SERVICE_URL", userServer.URL)

		bookServer := startFakeBookServer(&model.BookResponse{ID: 1, Title: "Go Book", Price: 29.99, Stock: 10}, http.StatusOK)
		defer bookServer.Close()
		os.Setenv("BOOK_SERVICE_URL", bookServer.URL)

		rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}).
			AddRow("1", "1", "1", 2, 59.98, "confirmed", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(1, 1, 2, 59.98).
			WillReturnRows(rows)

		r := setupRouter()
		r.POST("/orders", middleware.RequireAuth, handler.CreateOrder)

		body := `{"book_id":1,"quantity":2}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var order model.Order
		err := json.Unmarshal(w.Body.Bytes(), &order)
		assert.NoError(t, err)
		assert.Equal(t, 2, order.Quantity)
		assert.Equal(t, "confirmed", order.Status)
	})

	t.Run("InsufficientStock", func(t *testing.T) {
		userServer := startFakeUserServer(&model.UserResponse{ID: 1}, http.StatusOK)
		defer userServer.Close()
		os.Setenv("USER_SERVICE_URL", userServer.URL)

		bookServer := startFakeBookServer(&model.BookResponse{ID: 1, Price: 29.99, Stock: 1}, http.StatusOK)
		defer bookServer.Close()
		os.Setenv("BOOK_SERVICE_URL", bookServer.URL)

		r := setupRouter()
		r.POST("/orders", middleware.RequireAuth, handler.CreateOrder)

		body := `{"book_id":1,"quantity":5}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("BookNotFound", func(t *testing.T) {
		userServer := startFakeUserServer(&model.UserResponse{ID: 1}, http.StatusOK)
		defer userServer.Close()
		os.Setenv("USER_SERVICE_URL", userServer.URL)

		bookServer := startFakeBookServer(nil, http.StatusNotFound)
		defer bookServer.Close()
		os.Setenv("BOOK_SERVICE_URL", bookServer.URL)

		r := setupRouter()
		r.POST("/orders", middleware.RequireAuth, handler.CreateOrder)

		body := `{"book_id":99,"quantity":1}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		userServer := startFakeUserServer(nil, http.StatusNotFound)
		defer userServer.Close()
		os.Setenv("USER_SERVICE_URL", userServer.URL)

		r := setupRouter()
		r.POST("/orders", middleware.RequireAuth, handler.CreateOrder)

		body := `{"book_id":1,"quantity":1}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("MissingFields", func(t *testing.T) {
		r := setupRouter()
		r.POST("/orders", middleware.RequireAuth, handler.CreateOrder)

		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("NoToken", func(t *testing.T) {
		r := setupRouter()
		r.POST("/orders", middleware.RequireAuth, handler.CreateOrder)

		body := `{"book_id":1,"quantity":1}`
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetOrderByIDHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}).
			AddRow("1", "1", "2", 3, 89.97, "confirmed", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		r := setupRouter()
		r.GET("/orders/:id", middleware.RequireAuth, handler.GetOrderByID)

		req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var order model.Order
		err := json.Unmarshal(w.Body.Bytes(), &order)
		assert.NoError(t, err)
		assert.Equal(t, 3, order.Quantity)
		assert.Equal(t, "confirmed", order.Status)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}))

		r := setupRouter()
		r.GET("/orders/:id", middleware.RequireAuth, handler.GetOrderByID)

		req := httptest.NewRequest(http.MethodGet, "/orders/99", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		r := setupRouter()
		r.GET("/orders/:id", middleware.RequireAuth, handler.GetOrderByID)

		req := httptest.NewRequest(http.MethodGet, "/orders/abc", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetOrdersByUserIDHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}).
			AddRow("1", "1", "2", 1, 29.99, "confirmed", "2024-01-01T00:00:00Z").
			AddRow("2", "1", "3", 2, 59.98, "confirmed", "2024-01-02T00:00:00Z")

		mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		r := setupRouter()
		r.GET("/orders/user/:userId", middleware.RequireAuth, handler.GetOrdersByUserID)

		req := httptest.NewRequest(http.MethodGet, "/orders/user/1", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var orders []model.Order
		err := json.Unmarshal(w.Body.Bytes(), &orders)
		assert.NoError(t, err)
		assert.Len(t, orders, 2)
	})

	t.Run("WrongUserID", func(t *testing.T) {
		r := setupRouter()
		r.GET("/orders/user/:userId", middleware.RequireAuth, handler.GetOrdersByUserID)

		req := httptest.NewRequest(http.MethodGet, "/orders/user/1", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(2))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		r := setupRouter()
		r.GET("/orders/user/:userId", middleware.RequireAuth, handler.GetOrdersByUserID)

		req := httptest.NewRequest(http.MethodGet, "/orders/user/abc", nil)
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
