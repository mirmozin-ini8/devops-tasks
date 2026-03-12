package unit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"user-service/database"
	"user-service/handler"
	"user-service/middleware"
	"user-service/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
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

func TestRegisterHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).
			AddRow(1, "mir", "mir@test.com", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("INSERT INTO users").
			WithArgs("mir", "mir@test.com", sqlmock.AnyArg()).
			WillReturnRows(rows)

		r := setupRouter()
		r.POST("/users", handler.Register)

		body := `{"username":"mir","email":"mir@test.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var user model.User
		err = json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "mir", user.Username)
		assert.Empty(t, user.PasswordHash)
	})

	t.Run("MissingFields", func(t *testing.T) {
		r := setupRouter()
		r.POST("/users", handler.Register)

		body := `{"username":"mir"}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PasswordTooShort", func(t *testing.T) {
		r := setupRouter()
		r.POST("/users", handler.Register)

		body := `{"username":"mir","email":"mir@test.com","password":"abc"}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	t.Run("Success", func(t *testing.T) {
		os.Setenv("SECRET_KEY", "test-secret-key")

		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "mir", "mir@test.com", string(hash))

		mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE username = \\$1").
			WithArgs("mir").
			WillReturnRows(rows)

		r := setupRouter()
		r.POST("/login", handler.Login)

		body := `{"username":"mir","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp model.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "mir", resp.User.Username)
		assert.Empty(t, resp.User.PasswordHash)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "mir", "mir@test.com", string(hash))

		mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE username = \\$1").
			WithArgs("mir").
			WillReturnRows(rows)

		r := setupRouter()
		r.POST("/login", handler.Login)

		body := `{"username":"mir","password":"wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE username = \\$1").
			WithArgs("nobody").
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}))

		r := setupRouter()
		r.POST("/login", handler.Login)

		body := `{"username":"nobody","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("MissingFields", func(t *testing.T) {
		r := setupRouter()
		r.POST("/login", handler.Login)

		body := `{"username":"mir"}`
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetUserHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).
			AddRow(1, "mir", "mir@test.com", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		r := setupRouter()
		r.GET("/users/:id", handler.GetUser)

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user model.User
		err = json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "mir", user.Username)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at"}))

		r := setupRouter()
		r.GET("/users/:id", handler.GetUser)

		req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		r := setupRouter()
		r.GET("/users/:id", handler.GetUser)

		req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateUserHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		fetchRows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).
			AddRow(1, "mir", "mir@test.com", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(fetchRows)

		mock.ExpectExec("UPDATE users").
			WithArgs("mir_updated", "mir@test.com", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		r := setupRouter()
		r.PUT("/users/:id", middleware.RequireAuth, handler.UpdateUser)

		body := `{"username":"mir_updated"}`
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user model.User
		err = json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, "mir_updated", user.Username)
		assert.Equal(t, "mir@test.com", user.Email)
	})

	t.Run("NoToken", func(t *testing.T) {
		r := setupRouter()
		r.PUT("/users/:id", middleware.RequireAuth, handler.UpdateUser)

		body := `{"username":"mir_updated"}`
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("WrongUserID", func(t *testing.T) {
		r := setupRouter()
		r.PUT("/users/:id", middleware.RequireAuth, handler.UpdateUser)

		body := `{"username":"mir_updated"}`
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(2))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at"}))

		r := setupRouter()
		r.PUT("/users/:id", middleware.RequireAuth, handler.UpdateUser)

		body := `{"username":"mir_updated"}`
		req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		r := setupRouter()
		r.PUT("/users/:id", middleware.RequireAuth, handler.UpdateUser)

		body := `{"username":"mir_updated"}`
		req := httptest.NewRequest(http.MethodPut, "/users/abc", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+generateTestToken(1))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
