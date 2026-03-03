package model

type Book struct {
	ID        int     `db:"id" json:"id"`
	Title     string  `db:"title" json:"title"`
	Author    string  `db:"author" json:"author"`
	Price     float64 `db:"price" json:"price"`
	Stock     int     `db:"stock" json:"stock"`
	CreatedAt string  `db:"created_at,omitempty" json:"created_at,omitempty"`
}

type CreateBookRequest struct {
	Title  string  `json:"title" binding:"required"`
	Author string  `json:"author" binding:"required"`
	Price  float64 `json:"price" binding:"required,gt=0"`
	Stock  int     `json:"stock" binding:"gte=0"`
}

type UpdateBookRequest struct {
	Title  *string  `json:"title"`
	Author *string  `json:"author"`
	Price  *float64 `json:"price" binding:"omitempty,gt=0"`
	Stock  *int     `json:"stock" binding:"omitempty,gte=0"`
}
