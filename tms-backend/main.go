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
	port :="8080"
	r.Run(":" + port)
	fmt.Println("Server running on port", port)
}