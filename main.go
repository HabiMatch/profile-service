package main

import (
	"log"
	"net/http"
	"os"

	"github.com/HabiMatch/profile-service/router"
	"github.com/HabiMatch/profile-service/utils"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the database
	db, err := utils.InitDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func() {
		conn, _ := db.DB()
		_ = conn.Close()
	}()

	// Initialize the router
	r := router.InitRouter(db)

	// Get the port from environment variables or use the default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
