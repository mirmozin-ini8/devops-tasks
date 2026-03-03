package repository

import (
	"database/sql"
	"errors"
	"user-service/database"
	"user-service/model"
)

func GetUserByID(id int) (*model.User, error) {
	var u model.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, created_at FROM users WHERE id = $1", id).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	err := database.DB.QueryRow(
		"SELECT id, username, email, password_hash FROM users WHERE username = $1", username).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func CreateUser(username, email, passwordHash string) (*model.User, error) {
	var u model.User
	err := database.DB.QueryRow(
		`
		INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, username, email, created_at
		`, username, email, passwordHash).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &u, err
}

func UpdateUser(id int, req model.UpdateUserRequest) (*model.User, error) {
	existing, err := GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Username != nil {
		existing.Username = *req.Username
	}

	if req.Email != nil {
		existing.Email = *req.Email
	}

	_, err = database.DB.Exec(
		"UPDATE users SET username=$1, email=$2 WHERE id=$3", existing.Username, existing.Email, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}
