package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	UserID       int    `json:"id"`
	UserName     string `json:"name"`
	Email        string `json:"email"`
	UserPassword string `json:"password"`
}

// funkcije za http zahteve
func getUsers(c *gin.Context) {
	rows, err := db.Query("select id, name, email from public.user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.UserID, &u.UserName, &u.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, users)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")

	var u User
	err := db.QueryRow("select id, name, email from public.user where id=$1", id).Scan(&u.UserID, &u.UserName, &u.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		} else {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, u)
}

func postUser(c *gin.Context) {
	var newUser User

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	insertDynStmt := `insert into "user"("name", "email", "password") values($1, $2, $3) returning id`
	err := db.QueryRow(insertDynStmt, newUser.UserName, newUser.Email, newUser.UserPassword).Scan(&newUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newUser)
}

func putUser(c *gin.Context) {
	id := c.Param("id")
	var updatedUser User

	if err := c.BindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	updateStmt := `update public.user set name=$1, email=$2, password=$3 where id=$4`
	res, err := db.Exec(updateStmt, updatedUser.UserName, updatedUser.Email, updatedUser.UserPassword, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "user is not found"})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func patchUser(c *gin.Context) {
	id := c.Param("id")
	var updatedFields map[string]interface{}

	if err := c.BindJSON(&updatedFields); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if len(updatedFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updateStmt := "update public.user set "
	args := []interface{}{}
	argID := 1

	for field, value := range updatedFields {
		updateStmt += fmt.Sprintf("%s=$%d, ", field, argID)
		args = append(args, value)
		argID++
	}
	updateStmt = updateStmt[:len(updateStmt)-2] + " where id=$" + fmt.Sprintf("%d", argID)
	args = append(args, id)

	_, err := db.Exec(updateStmt, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")

	deleteStmt := `delete from public.user where id = $1`
	res, err := db.Exec(deleteStmt, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted"})
}
