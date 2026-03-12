package unit_test

import (
	"order-service/database"
	"order-service/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}).
			AddRow("1", "1", "2", 3, 89.97, "confirmed", "2024-01-01T00:00:00Z")

		mock.ExpectQuery("INSERT INTO orders").
			WithArgs(1, 2, 3, 89.97).
			WillReturnRows(rows)

		order, err := repository.CreateOrder(1, 2, 3, 89.97)
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, 3, order.Quantity)
		assert.Equal(t, 89.97, order.TotalPrice)
		assert.Equal(t, "confirmed", order.Status)
	})
}

func TestGetOrderByID(t *testing.T) {
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

		order, err := repository.GetOrderByID(1)
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, 3, order.Quantity)
		assert.Equal(t, "confirmed", order.Status)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE id = \\$1").
			WithArgs(99).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}))

		order, err := repository.GetOrderByID(99)
		assert.NoError(t, err)
		assert.Nil(t, order)
	})
}

func TestGetOrdersByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}
	defer db.Close()
	database.DB = db

	t.Run("ReturnsTwoOrders", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"}).
			AddRow("1", "1", "2", 1, 29.99, "confirmed", "2024-01-01T00:00:00Z").
			AddRow("2", "1", "3", 2, 59.98, "confirmed", "2024-01-02T00:00:00Z")

		mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE user_id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		orders, err := repository.GetOrdersByUserID(1)
		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, 29.99, orders[0].TotalPrice)
		assert.Equal(t, 59.98, orders[1].TotalPrice)
	})

	t.Run("ReturnsEmptySlice", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "book_id", "quantity", "total_price", "status", "created_at"})

		mock.ExpectQuery("SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE user_id = \\$1").
			WithArgs(99).
			WillReturnRows(rows)

		orders, err := repository.GetOrdersByUserID(99)
		assert.NoError(t, err)
		assert.NotNil(t, orders)
		assert.Len(t, orders, 0)
	})
}
