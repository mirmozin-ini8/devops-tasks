package main

import (
	"book-service/database"
	"book-service/handler"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env variables found")
	}

	database.Connect()

	router := gin.Default()

	books := router.Group("/books")
	{
		books.GET("/health", handler.HealthCheck)
		books.GET("/metrics", func(c *gin.Context) {
			c.String(200, "tbd\n")
		})
		books.GET("", handler.GetAllBooks)
		books.GET("/:id", handler.GetBookByID)
		books.POST("", handler.CreateBook)
		books.PUT("/:id", handler.UpdateBook)
		books.DELETE("/:id", handler.DeleteBook)
	}

	port := os.Getenv("BOOKS_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("book service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
