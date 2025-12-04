package main

import (
	"log"
	"net/http"
	"os"

	v1 "noble-group-services/api/v1"
	"noble-group-services/core"
	"noble-group-services/crud"
	_ "noble-group-services/docs" // Swagger docs
)

// @title Noble Group Services API
// @version 1.0
// @description API for Noble Group Services e-commerce platform.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

func main() {
	// Database connection string
	// Default to the one in comments or environment variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/noble"
	}

	// Initialize Database
	if err := core.InitDB(dsn); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer core.CloseDB()

	// Set DB for CRUD operations
	crud.SetDB(core.DB)

	// Setup Router
	mux := http.NewServeMux()
	v1.SetupRoutes(mux)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
