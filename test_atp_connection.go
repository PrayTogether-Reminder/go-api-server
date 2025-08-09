package main

import (
	"fmt"
	"log"
	"os"

	"github.com/godoes/gorm-oracle"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if user == "" {
		user = "ADMIN"
	}

	fmt.Println("Oracle ATP Connection Test (Without Wallet)")
	fmt.Println("============================================")
	fmt.Printf("User: %s\n", user)

	// ATP connection string
	tnsDesc := "(description= (retry_count=20)(retry_delay=3)(address=(protocol=tcps)(port=1522)(host=adb.ap-chuncheon-1.oraclecloud.com))(connect_data=(service_name=g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com))(security=(ssl_server_dn_match=yes)))"

	// Build DSN
	dsn := fmt.Sprintf("%s/%s@%s", user, password, tnsDesc)

	fmt.Println("\nUsing ATP Connection String:")
	fmt.Println("Host: adb.ap-chuncheon-1.oraclecloud.com")
	fmt.Println("Port: 1522")
	fmt.Println("Service: g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com")
	fmt.Println("Protocol: TCPS (SSL)")

	// Test connection
	fmt.Println("\nConnecting to Oracle ATP...")
	db, err := gorm.Open(oracle.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to Oracle ATP: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping Oracle ATP: %v", err)
	}

	fmt.Println("✅ Successfully connected to Oracle ATP!")

	// Get database information
	fmt.Println("\n=== Database Information ===")

	// Get version
	var version string
	if err := db.Raw("SELECT VERSION FROM V$INSTANCE").Scan(&version).Error; err == nil {
		fmt.Printf("Database Version: %s\n", version)
	}

	// Get service name
	var serviceName string
	if err := db.Raw("SELECT SYS_CONTEXT('USERENV', 'SERVICE_NAME') FROM DUAL").Scan(&serviceName).Error; err == nil {
		fmt.Printf("Service Name: %s\n", serviceName)
	}

	// Get database name
	var dbName string
	if err := db.Raw("SELECT SYS_CONTEXT('USERENV', 'DB_NAME') FROM DUAL").Scan(&dbName).Error; err == nil {
		fmt.Printf("Database Name: %s\n", dbName)
	}

	// Test table operations
	fmt.Println("\n=== Testing Table Operations ===")

	type TestTable struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	// Drop table if exists
	if db.Migrator().HasTable(&TestTable{}) {
		if err := db.Migrator().DropTable(&TestTable{}); err != nil {
			log.Printf("Failed to drop existing test table: %v", err)
		}
	}

	// Create table
	if err := db.AutoMigrate(&TestTable{}); err != nil {
		log.Printf("Failed to create test table: %v", err)
	} else {
		fmt.Println("✅ Table created successfully")

		// Insert test data
		testRecord := TestTable{ID: 1, Name: "ATP Test"}
		if err := db.Create(&testRecord).Error; err != nil {
			log.Printf("Failed to insert test record: %v", err)
		} else {
			fmt.Println("✅ Data inserted successfully")

			// Query test data
			var result TestTable
			if err := db.First(&result).Error; err != nil {
				log.Printf("Failed to query test record: %v", err)
			} else {
				fmt.Printf("✅ Data queried successfully: ID=%d, Name=%s\n", result.ID, result.Name)
			}

			// Clean up
			if err := db.Migrator().DropTable(&TestTable{}); err != nil {
				log.Printf("Failed to drop test table: %v", err)
			} else {
				fmt.Println("✅ Table cleaned up successfully")
			}
		}
	}

	fmt.Println("\n🎉 Oracle ATP connection test completed successfully!")
	fmt.Println("\nYour application can now connect to Oracle ATP without wallet.")
	fmt.Println("The server will use this ATP connection string automatically.")
}
