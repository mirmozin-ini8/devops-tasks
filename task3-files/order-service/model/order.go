package model

type Order struct {
	ID         string  `db:"id" json:"id"`
	UserID     string  `db:"user_id" json:"user_id"`
	BookID     string  `db:"book_id" json:"book_id"`
	Quantity   int     `db:"quantity" json:"quantity"`
	TotalPrice float64 `db:"total_price" json:"total_price"`
	Status     string  `db:"status" json:"status"`
	CreatedAt  string  `db:"created_at" json:"created_at,omitempty"`
}

type CreateOrderRequest struct {
	BookID   int `json:"book_id" binding:"required,gt=0"`
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type BookResponse struct {
	ID    int     `json:"id"`
	Title string  `json:"title"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
