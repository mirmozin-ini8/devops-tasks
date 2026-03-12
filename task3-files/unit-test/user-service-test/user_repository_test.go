package unit_test

import (
	"testing"
	"user-service/database"
	"user-service/model"
	"user-service/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetUserByID(t *testing.T) {
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

		user, err := repository.GetUserByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "mir", user.Username)
		assert.Equal(t, "mir@test.com", user.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at"}))

		user, err := repository.GetUserByID(99)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestGetUserByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}).
			AddRow(1, "mir", "mir@test.com", "$2a$10$hashedpassword")

		mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE username = \\$1").
			WithArgs("mir").
			WillReturnRows(rows)

		user, err := repository.GetUserByUsername("mir")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "mir", user.Username)
		assert.Equal(t, "$2a$10$hashedpassword", user.PasswordHash)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, email, password_hash FROM users WHERE username = \\$1").
			WithArgs("nobody").
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password_hash"}))

		user, err := repository.GetUserByUsername("nobody")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestCreateUser(t *testing.T) {
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
			WithArgs("mir", "mir@test.com", "alreadyhashed").
			WillReturnRows(rows)

		user, err := repository.CreateUser("mir", "mir@test.com", "alreadyhashed")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "mir", user.Username)
		assert.Equal(t, "mir@test.com", user.Email)
	})
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("UpdateUsername", func(t *testing.T) {
		fetchRows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).
			AddRow(1, "mir", "mir@test.com", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(fetchRows)

		mock.ExpectExec("UPDATE users").
			WithArgs("mir_updated", "mir@test.com", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		newUsername := "mir_updated"
		req := model.UpdateUserRequest{Username: &newUsername}

		user, err := repository.UpdateUser(1, req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "mir_updated", user.Username)
		assert.Equal(t, "mir@test.com", user.Email)
	})

	t.Run("UpdateEmail", func(t *testing.T) {
		fetchRows := sqlmock.NewRows([]string{"id", "username", "email", "created_at"}).
			AddRow(1, "mir", "mir@test.com", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(fetchRows)

		mock.ExpectExec("UPDATE users").
			WithArgs("mir", "new@test.com", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		newEmail := "new@test.com"
		req := model.UpdateUserRequest{Email: &newEmail}

		user, err := repository.UpdateUser(1, req)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "mir", user.Username)
		assert.Equal(t, "new@test.com", user.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, username, email, created_at FROM users WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at"}))

		newUsername := "anything"
		req := model.UpdateUserRequest{Username: &newUsername}

		user, err := repository.UpdateUser(99, req)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}
