package main

import (
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/api/v1"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/core"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/middleware"

	_ "github.com/GoogleCloudPlatform/golang-samples/run/helloworld/docs"
)

func main() {
	// Initialize the database
	// Replace with your actual database connection string
	if err := core.InitDB("postgres://user:password@localhost/dbname?sslmode=disable"); err != nil {
		log.Fatal(err)
	}

	// Setup the API routes
	v1.SetupRoutes()

	// Create a new ServeMux and apply the CORS middleware
	mux := http.NewServeMux()
	v1.SetupRoutes(mux) // Assuming SetupRoutes takes a mux

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
