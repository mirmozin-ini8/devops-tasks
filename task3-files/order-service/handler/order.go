package handler

import (
	"net/http"
	"order-service/client"
	"order-service/database"
	"order-service/metrics"
	"order-service/model"
	"order-service/repository"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateOrder(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int)

	var req model.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := client.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user service unavailable"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	book, err := client.GetBook(req.BookID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "book service unavailable"})
		return
	}
	if book == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	if book.Stock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "insufficient stock",
			"available stock": book.Stock,
			"requested":       req.Quantity,
		})
		return
	}

	totalPrice := book.Price * float64(req.Quantity)

	order, err := repository.CreateOrder(userID, req.BookID, req.Quantity, totalPrice)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	metrics.OrdersCreatedTotal.Inc()
	metrics.OrdersRevenueTotal.Add(totalPrice)

	c.JSON(http.StatusCreated, order)
}

func GetOrderByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := repository.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func GetOrdersByUserID(c *gin.Context) {
	requestedUserID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	loggedInUserID, _ := c.Get("user_id")

	if loggedInUserID.(int) != requestedUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot view another user's orders"})
		return
	}

	orders, err := repository.GetOrdersByUserID(requestedUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func HealthCheck(c *gin.Context) {
	err := database.DB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "down",
			"error":  "orders_db connection failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "order-service",
	})
}
