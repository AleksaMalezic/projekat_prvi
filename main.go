package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// ovo cu posle da zamenim json fajlom
type user struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Age        int    `json:"age"`
	Occupation string `json:"occupation"`
}

type User struct {
	Username string
	Password string
}

var (
	users         []user
	jsonFile      = "user.json"
	page          = "/users"
	pageid        = "/users/:id"
	host          = "localhost:8080"
	dummyUsername = "Aleksa"
	dummyPassword = "ryko123"
	secretKey     = []byte("secret-key")
)

func main() {
	//otvaranje json fajla
	binaryFile := jsonToBinary(jsonFile)
	defer binaryFile.Close()

	//prebacivanje bin fajla u stukturu
	binToStruct(binaryFile, &users)

	//HTTP requests
	router := gin.Default()

	router.Use(authMiddleware())

	router.POST("/login", loginHandler)

	//zasticeni endpointi
	protected := router.Group("/")
	protected.Use(authMiddleware())
	{
		router.GET(page, getUsers)
		router.GET(pageid, getUserByID)
		router.POST(page, postUsers)
		router.PUT(pageid, putUser)
		router.PATCH(pageid, patchUser)
		router.DELETE(pageid, deleteUser)
		router.GET("/protected", ProtectedHandler)
	}

	router.Run(host)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.URL.Path {
		case "/login":
			c.Next()
			return
		case "/users":
			c.Next()
			return
		}

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}
		tokenString = tokenString[len("Bearer "):]

		err := verifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// funkcije za http zahteve
func getUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, users)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range users {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

func postUsers(c *gin.Context) {
	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	users = append(users, newUser)
	c.IndentedJSON(http.StatusCreated, newUser)
}

func putUser(c *gin.Context) {
	id := c.Param("id")
	var updatedUser user

	if err := c.BindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	for i, a := range users {
		if a.ID == id {
			users[i] = updatedUser
			c.IndentedJSON(http.StatusOK, updatedUser)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

func patchUser(c *gin.Context) {
	id := c.Param("id")
	var updatedFields map[string]interface{}

	if err := c.BindJSON(&updatedFields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	//fmt.Printf("Updated fields: %+v\n", updatedFields)

	for i, a := range users {
		if a.ID == id {
			if name, ok := updatedFields["name"].(string); ok {
				users[i].Name = name
			}
			if age, ok := updatedFields["age"].(float64); ok {
				// Debug print
				fmt.Printf("Age type: %T, value: %v\n", updatedFields["age"], updatedFields["age"])
				users[i].Age = int(age)
			}
			if occupation, ok := updatedFields["occupation"].(string); ok {
				users[i].Occupation = occupation
			}
			c.IndentedJSON(http.StatusOK, users[i])
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	for i, a := range users {
		if a.ID == id {
			users = append(users[:i], users[i+1:]...)
			c.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
}

// funkcije za token
func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func loginHandler(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if u.Username == dummyUsername && u.Password == dummyPassword {
		tokenString, err := createToken(u.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	}
}

func ProtectedHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Message": "Welcome to the protected area"})
}
