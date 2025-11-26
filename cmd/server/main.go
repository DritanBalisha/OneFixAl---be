package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"OneFixAL/internal/db"
	"OneFixAL/internal/models"
	rt "OneFixAL/internal/routes"
)

func main() {
	// Load .env only in LOCAL development
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️  Warning: .env file not found, using system environment variables")
		} else {
			log.Println("📄 .env file loaded successfully")
		}
	} else {
		log.Println("🚀 Running in Railway environment")
	}

	// Connect to DB
	db.Connect()

	// Auto-migrate models
	err := db.DB.AutoMigrate(
		&models.User{},
		&models.TechnicianProfile{},
		&models.Booking{},
		&models.Availability{},
		&models.Notification{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}
	log.Println("✅ Migrations complete")

	// Setup router
	r := rt.SetupRouter()

	// Railway assigns PORT dynamically → must use os.Getenv("PORT")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Local default
	}

	log.Printf("🚀 Server running on port %s", port)
	r.Run(":" + port)
}
