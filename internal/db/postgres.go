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

func getDBConnString() string {
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE") 

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func Connect() {
	dsn := getDBConnString()

	var err error
	// Retry up to 10 times (30s total)
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
