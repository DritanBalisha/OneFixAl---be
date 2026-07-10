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
	// Railway provides DATABASE_URL directly — use it if available
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	// Fallback to individual vars
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname,
	)
}

func Connect() {
	dsn := getDBConnString()
	var err error

	for i := 1; i <= 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// ✅ Configure connection pool
			sqlDB, err := DB.DB()
			if err == nil {
				sqlDB.SetMaxOpenConns(10)
				sqlDB.SetMaxIdleConns(5)
				sqlDB.SetConnMaxLifetime(time.Hour)
			}
			fmt.Println("✅ Connected to database")
			return
		}
		log.Printf("⏳ DB not ready, retrying in 3s... (%d/10): %v", i, err)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("❌ Failed to connect after retries: %v", err)
}
