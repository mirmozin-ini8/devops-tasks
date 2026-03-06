package main

import (
	"book-service/database"
	"book-service/handler"
	"book-service/metrics"
	"book-service/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env variables found")
	}

	database.Connect()

	router := gin.Default()
	router.Use(middleware.MetricMiddleware())

	books := router.Group("/books")
	{
		books.GET("/health", handler.HealthCheck)
		books.GET("/metrics", gin.WrapH(promhttp.Handler()))
		books.GET("", handler.GetAllBooks)
		books.GET("/:id", handler.GetBookByID)
		books.POST("", handler.CreateBook)
		books.PUT("/:id", handler.UpdateBook)
		books.DELETE("/:id", handler.DeleteBook)
	}

	_ = metrics.BooksCreatedTotal

	port := os.Getenv("BOOKS_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("book service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
