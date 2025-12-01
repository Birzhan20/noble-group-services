package main

import (
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/api/v1"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/core"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/middleware"

	_ "github.com/GoogleCloudPlatform/golang-samples/run/helloworld/docs"
)

func main() {
	// Initialize the database
	// Read connection string from environment variable, fallback to default for local dev
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost/dbname?sslmode=disable"
		log.Println("DATABASE_URL not set, using default connection string")
	}

	if err := core.InitDB(connStr); err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}

	// Create a new ServeMux and apply the CORS middleware
	mux := http.NewServeMux()

	// Setup the API routes using the created mux
	v1.SetupRoutes(mux)

	httpserver := &http.Server{
		Addr:    ":3000",
		Handler: middleware.CORSMiddleware(mux),
	}

	// Start the server
	log.Println("Starting server on :3000")
	if err := httpserver.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
