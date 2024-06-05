package main

import (
	"net/http"

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
)

func main() {

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
	//r.PUT(page, putUsers)
	//r.PATCH(page, patchUser)
	//r.DELETE(page, deleteUsers)

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
