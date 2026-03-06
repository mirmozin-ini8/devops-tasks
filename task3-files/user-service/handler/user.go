package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"user-service/database"
	"user-service/metrics"
	"user-service/model"
	"user-service/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"user-service/middleware"
)

func Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user, err := repository.CreateUser(req.Username, req.Email, string(hashedPassword))
	if err != nil {

		if strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	metrics.UsersRegisteredTotal.Inc()

	c.JSON(http.StatusCreated, user)
}

func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	user.PasswordHash = ""

	c.JSON(http.StatusOK, model.LoginResponse{
		Token: token,
		User:  *user,
	})
}

func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := repository.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	loggedInUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if loggedInUserID.(int) != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot update another user's profile"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := repository.UpdateUser(id, req)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func HealthCheck(c *gin.Context) {
	err := database.DB.Ping()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "down",
			"error":  "users_db database connection failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "user-service",
	})
}

func generateToken(userID int) (string, error) {
	claims := &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}
