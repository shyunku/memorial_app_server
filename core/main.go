package main

import (
	"github.com/joho/godotenv"
	"memorial_app_server/controllers"
	"memorial_app_server/database"
	"memorial_app_server/log"
	"memorial_app_server/service"
	"os"
)

func main() {
	// Create Jwt secret key if needed
	//crypto.PrintNewJwtSecret()

	// Load environment variables
	log.Info("Initializing environments...")
	if err := godotenv.Load(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}

	// Initialize database
	log.Info("Initializing database...")
	if _, err := database.Initialize(); err != nil {
		log.Error(err)
		os.Exit(-2)
	}

	// Initialize in-memory database
	log.Info("Initializing in-memory database...")
	service.InMemoryDB = service.NewRedis()

	// Run web server with gin
	controllers.RunGin()
}
