package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var (
	host = "localhost:8080"

	secretKey = []byte("secret-key")

	dbHost = "localhost"
	dbPort = 5432
	dbUser = "postgres"
	dbPass = "nikadpijan123"
	dbName = "projekat_prvi"
	db     *sql.DB
)

func main() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	var err error
	db, err = sql.Open("postgres", psqlconn)
	checkDbError(err)
	defer db.Close()
	err = db.Ping()
	checkDbError(err)
	fmt.Println("connected db")

	router := gin.Default()
	router.POST("/api/login", loginHandler)

	unprotected := router.Group("/api/organisational_unit")
	unprotected.Use()
	{
		unprotected.GET("", getNodes)
		unprotected.GET("/:id", getNodeById)
		unprotected.POST("", postNode)
		unprotected.PUT("/:id", putNode)
		unprotected.PATCH("/:id", patchNode)
		unprotected.DELETE("/:id", deleteNode)
	}

	protected := router.Group("/api/user")
	protected.Use(authMiddleware())
	{
		protected.GET("", getUsers)
		protected.GET("/:id", getUserByID)
		protected.POST("", postUser)
		protected.PUT("/:id", putUser)
		protected.PATCH("/:id", patchUser)
		protected.DELETE("/:id", deleteUser)

		protected.GET("/protected", ProtectedHandler)
	}

	router.Run(host)
}
