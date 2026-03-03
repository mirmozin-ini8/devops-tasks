package main

import (
	"log"
	"os"
	"user-service/database"
	"user-service/handler"
	"user-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}

	database.Connect()

	router := gin.Default()

	router.GET("/health", handler.HealthCheck)

	router.GET("/metrics", func(c *gin.Context) {
		c.String(200, "tbd\n")
	})

	router.POST("/login", handler.Login)

	users := router.Group("/users")
	{
		users.POST("", handler.Register)
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", middleware.RequireAuth, handler.UpdateUser)
	}

	port := os.Getenv("USERS_SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("user service starting on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
