package main

import (
	"fmt"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	"github.com/gin-gonic/gin"
)


func main() {
	// Initialize the application
	// app := config.NewApp()

	// Connect to the database
	// database.ConnectDB()

	// Start the server
	// app.Start()
	config.LoadConfig()
	database.ConnectDB()
	// Initialize the server
	// server := config.NewServer()
	// Start the server
	// server.Start()

	r := gin.Default()
	r.LoadHTMLGlob("./views/*")
	port :="8080"

	// Define your routes here
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	// Setup tenant routes
	SetupTenantRoutes(r)

	r.Run(":" + port)
	fmt.Println("Server running on port", port)
}