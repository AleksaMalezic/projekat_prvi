package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
)

func checkDbError(err error) {
	if err != nil {
		panic(err)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}
		tokenString = tokenString[len("Bearer "):]

		err := verifyToken(c, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func loginHandler(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	user, err := getUserByEmail(u.Email)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if u.UserPassword != user.UserPassword {
		log.Printf("Password comparison failed")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := createToken(user.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func getUserByEmail(email string) (User, error) {
	var user User
	query := "SELECT id, email, password, name FROM public.user WHERE email=$1"
	err := db.QueryRow(query, email).Scan(&user.UserID, &user.Email, &user.UserPassword, &user.UserName)
	if err != nil {
		return user, err
	}
	return user, nil
}

func createToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func verifyToken(c *gin.Context, tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unexpected signing method"})
			c.Abort()
			return nil, nil
		}
		return secretKey, nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return err
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return err
	}

	return nil
}

func ProtectedHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the protected route!"})
}
