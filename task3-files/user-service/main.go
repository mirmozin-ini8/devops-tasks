package main

import (
	"log"
	"os"
	"user-service/database"
	"user-service/handler"
	"user-service/metrics"
	"user-service/middleware"

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

	router.POST("/login", handler.Login)

	users := router.Group("/users")
	{
		users.GET("/health", handler.HealthCheck)
		users.GET("/metrics", gin.WrapH(promhttp.Handler()))
		users.POST("", handler.Register)
		users.GET("/:id", handler.GetUser)
		users.PUT("/:id", middleware.RequireAuth, handler.UpdateUser)
	}

	_ = metrics.UsersRegisteredTotal

	port := os.Getenv("USERS_SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("user service starting on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
