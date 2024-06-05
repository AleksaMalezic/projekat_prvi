package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func jsonToBinary(jsonFileName string) *os.File {
	jsonFile, err := os.Open(jsonFileName)
	if err != nil {
		fmt.Println("Cannot open JSON file:", err)
		return nil
	}
	return jsonFile
}

func binToStruct(binaryFile *os.File, users *[]user) {
	decoder := json.NewDecoder(binaryFile)
	err := decoder.Decode(users)
	if err != nil {
		fmt.Println("Cannot convert binary to struct:", err)
		return
	}
}

/*func postFunc(c *gin.Context, users *[]user) {
	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	*users = append(*users, newUser)
	c.IndentedJSON(http.StatusCreated, newUser)
}

func getHttp(page string, r *gin.Engine, users *[]user) {
	r.GET(page, func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, users)
	})
}

func postHttp(page string, r *gin.Engine, users *[]user) {
	r.POST(page, func(c *gin.Context) {
		postFunc(c, users)
	})
}

func runServer(r *gin.Engine, host string) {
	err := r.Run(host)
	if err != nil {
		fmt.Println("Failed to run the server:", err)
		return
	}
}*/
