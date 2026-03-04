package repository

import (
	"database/sql"
	"errors"
	"order-service/database"
	"order-service/model"
)

func CreateOrder(userID, bookID, quantity int, totalPrice float64) (*model.Order, error) {
	var o model.Order
	err := database.DB.QueryRow(
		`
		INSERT INTO orders (user_id, book_id, quantity, total_price, status) VALUES ($1, $2, $3, $4, 'confirmed') RETURNING id, user_id, book_id, quantity, total_price, status, created_at
		`, userID, bookID, quantity, totalPrice).Scan(&o.ID, &o.UserID, &o.BookID, &o.Quantity, &o.TotalPrice, &o.Status, &o.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &o, nil
}

func GetOrderByID(id int) (*model.Order, error) {
	var o model.Order
	err := database.DB.QueryRow(
		`SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE id = $1`, id).Scan(&o.ID, &o.UserID, &o.BookID, &o.Quantity, &o.TotalPrice, &o.Status, &o.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &o, nil
}

func GetOrdersByUserID(userID int) ([]model.Order, error) {
	rows, err := database.DB.Query(
		`SELECT id, user_id, book_id, quantity, total_price, status, created_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var orders []model.Order

	for rows.Next() {
		var o model.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.BookID, &o.Quantity, &o.TotalPrice, &o.Status, &o.CreatedAt)

		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}
