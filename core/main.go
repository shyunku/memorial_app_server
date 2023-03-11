package main

import (
	"github.com/joho/godotenv"
	"memorial_app_server/controllers"
	"memorial_app_server/database"
	"memorial_app_server/log"
	"os"
)

func main() {
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

	// Run web server with gin
	controllers.RunGin()
}
