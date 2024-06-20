package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
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
	//api
	users    []user
	jsonFile = "user.json"
	page     = "/users"
	pageid   = "/users/:id"
	host     = "localhost:8080"
	//jwt token credentials
	dummyUsername = "Aleksa"
	dummyPassword = "ryko123"
	secretKey     = []byte("secret-key")
	//database
	dbHost = "localhost"
	dbPort = 5432
	dbUser = "postgres"
	dbPass = "nikadpijan123"
	dbName = "projekat_prvi"
	db     *sql.DB
)

func main() {

	//otvaranje baze podataka
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	var err error
	db, err = sql.Open("postgres", psqlconn)
	checkDbError(err)
	defer db.Close()
	err = db.Ping()
	checkDbError(err)
	fmt.Println("connected db")

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

func checkDbError(err error) {
	if err != nil {
		panic(err)
	}
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
	rows, err := db.Query("select id, name, age, occupation from public.users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []user
	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Name, &u.Age, &u.Occupation); err != nil {
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

	var u user
	err := db.QueryRow("select id, name, age, occupation from public.users where id=$1", id).Scan(&u.ID, &u.Name, &u.Age, &u.Occupation)
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

func postUsers(c *gin.Context) {
	var newUser user

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	insertDynStmt := `insert into "users"("name", "age", "occupation") values($1, $2, $3) returning id`
	err := db.QueryRow(insertDynStmt, newUser.Name, newUser.Age, newUser.Occupation).Scan(&newUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newUser)
}

func putUser(c *gin.Context) {
	id := c.Param("id")
	var updatedUser user

	if err := c.BindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	updateStmt := `update public.users set name=$1, age=$2, occupation=$3 where id=$4`
	res, err := db.Exec(updateStmt, updatedUser.Name, updatedUser.Age, updatedUser.Occupation, id)
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

	c.IndentedJSON(http.StatusOK, updatedUser)
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

	updateStmt := "update public.users set "
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

	deleteStmt := `delete from public.users where id = $1`
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
