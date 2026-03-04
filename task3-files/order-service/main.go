package main

import (
	"log"
	"order-service/database"
	"order-service/handler"
	"order-service/middleware"
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

	router.GET("/health", handler.HealthCheck)
	router.GET("/metrics", func(c *gin.Context) {
		c.String(200, "tbd\n")
	})

	orders := router.Group("/orders")
	orders.Use(middleware.RequireAuth)
	{
		orders.POST("", handler.CreateOrder)
		orders.GET("/:id", handler.GetOrderByID)
		orders.GET("/user/:userid", handler.GetOrdersByUserID)
	}

	port := os.Getenv("ORDERS_SERVER_PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("order service starting on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
