package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	var err error

	for i := 1; i <= 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("✅ Connected to database")
			return
		}

		log.Printf("⏳ Database not ready, retrying in 3s... (attempt %d/10)", i)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("❌ Failed to connect to database after retries: %v", err)
}
