package main

import (
	"log"
	"net/http"

	"adventure-engine/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Setup routes
	server.SetupRoutes(r)

	// Start server
	log.Println("Starting Saga Engine server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
