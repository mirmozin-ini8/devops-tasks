package model

type User struct {
	ID           int    `db:"id" json:"id"`
	Username     string `db:"username" json:"username"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"-"`
	CreatedAt    string `db:"created_at" json:"created_at,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=20"`
	Email    *string `json:"email" binding:"omitempty,email"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
