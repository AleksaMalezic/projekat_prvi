package main

import (
	"fmt"
	"net/http"

	//"github.com/golang-jwt/jwt"
	"github.com/gin-gonic/gin"
)

// ovo cu posle da zamenim json fajlom
type user struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Age        int    `json:"age"`
	Occupation string `json:"occupation"`
}

var (
	users    []user
	jsonFile = "user.json"
	page     = "/users"
	pageid   = "/users/:id"
	host     = "localhost:8080"

	secretKey = []byte("secret-key")
)

func main() {

	//token

	//otvaranje json fajla
	binaryFile := jsonToBinary(jsonFile)
	defer binaryFile.Close()

	//prebacivanje bin fajla u stukturu
	binToStruct(binaryFile, &users)

	//HTTP requests
	r := gin.Default()

	r.GET(page, getUsers)
	r.GET(pageid, getUserByID)
	r.POST(page, postUsers)
	r.PUT(pageid, putUser)
	r.PATCH(pageid, patchUser)
	r.DELETE(pageid, deleteUser)

	r.Run(host)

}

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

	fmt.Printf("Updated fields: %+v\n", updatedFields)

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
