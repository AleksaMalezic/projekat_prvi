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

	protected := router.Group("/api")
	protected.Use(authMiddleware())
	{
		protected.GET(organisationalUnitPage, getNodes)
		protected.GET(organisationalUnitPageId, getNodeById)
		protected.POST(organisationalUnitPage, postNode)
		protected.PUT(organisationalUnitPageId, putNode)
		protected.PATCH(organisationalUnitPageId, patchNode)
		protected.DELETE(organisationalUnitPageId, deleteNode)

		protected.GET(userPage, getUsers)
		protected.GET(userPageId, getUserByID)
		protected.POST(userPage, postUser)
		protected.PUT(userPageId, putUser)
		protected.PATCH(userPageId, patchUser)
		protected.DELETE(userPageId, deleteUser)

		protected.GET("/protected", ProtectedHandler)
	}

	router.Run(host)
}
