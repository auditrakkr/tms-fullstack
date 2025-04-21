package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/auditrakkr/tms-fullstack/tms-backend/config"
	"github.com/auditrakkr/tms-fullstack/tms-backend/database"
	"github.com/auditrakkr/tms-fullstack/tms-backend/global"
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
	r.Static("/assets", "./assets/dist")
	r.Static("/js", "./assets/js") // Add this for the themeToggle.js
	LoadTemplates(r, "./views")
	port :="8080"

	// Define your routes here
    r.GET("/", func(c *gin.Context) {
    fmt.Println("Home route accessed!")

    // Use the API_VERSION from global package
    apiVersionPath := ""
    if global.USE_API_VERSION_IN_URL && global.API_VERSION != "" {
        apiVersionPath = "/" + global.API_VERSION
    }

    fmt.Println("About to render template with:", gin.H{
        "apiVersion": global.API_VERSION,
        "title": "Tenant Management System",
        "homeActive": "true",
        "currentUrlSlug": apiVersionPath,
    })

    // Render the home view with the necessary variables
    c.HTML(200, "guest-website/home.html", gin.H{
        "apiVersion": global.API_VERSION,
        "title": "Tenant Management System",
        "homeActive": "true",
        "currentUrlSlug": apiVersionPath,
    })

    fmt.Println("Template should have rendered")
})

	// Setup tenant routes
	SetupTenantRoutes(r)

	r.Run(":" + port)
	fmt.Println("Server running on port", port)
}

// LoadTemplates loads all HTML templates from a directory recursively
func LoadTemplates(router *gin.Engine, templatesDir string) {
    templ := template.New("")
    err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        // Skip directories
        if info.IsDir() {
            return nil
        }
        // Only process .html files
        if !strings.HasSuffix(path, ".html") {
            return nil
        }

        // Read the template file
        b, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        // Get the relative path from the templates directory
        name, err := filepath.Rel(templatesDir, path)
        if err != nil {
            return err
        }

        // Parse the template
        _, err = templ.New(name).Parse(string(b))
        return err
    })

    if err != nil {
        panic(err)
    }

    router.SetHTMLTemplate(templ)
}