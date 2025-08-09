package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/godoes/gorm-oracle"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Build DSN
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	service := os.Getenv("DB_SERVICE")

	encodedPassword := url.QueryEscape(password)
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%s/%s?SSL=true",
		user, encodedPassword, host, port, service)

	// Connect
	db, err := gorm.Open(oracle.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Tables in Database:")
	fmt.Println("===================")

	// Query all user tables
	var tables []struct {
		TableName string `gorm:"column:TABLE_NAME"`
	}

	db.Raw("SELECT TABLE_NAME FROM USER_TABLES ORDER BY TABLE_NAME").Scan(&tables)

	for _, t := range tables {
		// Get actual row count
		var count int64
		db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", t.TableName)).Scan(&count)
		fmt.Printf("- %-30s: %d rows\n", t.TableName, count)
	}
}
