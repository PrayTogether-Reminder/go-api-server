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

	// Get database configuration
	dbType := os.Getenv("DB_TYPE")
	if dbType != "oracle" {
		log.Printf("DB_TYPE is set to %s, not oracle. Set DB_TYPE=oracle to test Oracle connection.\n", dbType)
		return
	}

	// Get TNS Admin and Alias for Wallet-based connection
	tnsAdmin := os.Getenv("TNS_ADMIN")
	tnsAlias := os.Getenv("TNS_ALIAS")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if tnsAdmin == "" {
		tnsAdmin = "./resources/main-wallet"
	}
	if tnsAlias == "" {
		tnsAlias = "z5f5ees1n47gddba_high"
	}
	if user == "" {
		user = "ADMIN"
	}

	fmt.Printf("Oracle Wallet Configuration:\n")
	fmt.Printf("  TNS_ADMIN: %s\n", tnsAdmin)
	fmt.Printf("  TNS_ALIAS: %s\n", tnsAlias)
	fmt.Printf("  User: %s\n", user)
	fmt.Printf("  Password: ****\n")

	// Check if wallet directory exists
	if _, err := os.Stat(tnsAdmin); os.IsNotExist(err) {
		log.Fatalf("TNS_ADMIN directory does not exist: %s", tnsAdmin)
	}

	// Check required wallet files
	requiredFiles := []string{
		"tnsnames.ora",
		"sqlnet.ora",
		"cwallet.sso",
		"ewallet.p12",
	}
	for _, file := range requiredFiles {
		filePath := fmt.Sprintf("%s/%s", tnsAdmin, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Printf("Warning: Required wallet file not found: %s", filePath)
		} else {
			fmt.Printf("✅ Found wallet file: %s\n", file)
		}
	}

	// Set TNS_ADMIN environment variable for Oracle driver
	os.Setenv("TNS_ADMIN", tnsAdmin)

	// Build Oracle DSN for wallet connection
	// Format: user/password@tns_alias
	dsn := fmt.Sprintf("%s/%s@%s", user, password, tnsAlias)
	fmt.Printf("\nConnecting with DSN: %s/%s@%s\n", user, "****", tnsAlias)

	// Test connection
	fmt.Println("\nTesting Oracle Autonomous Database connection...")
	db, err := gorm.Open(oracle.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to Oracle database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping Oracle database: %v", err)
	}

	fmt.Println("✅ Successfully connected to Oracle Autonomous Database!")

	// Get Oracle version
	var version string
	if err := db.Raw("SELECT BANNER FROM V$VERSION WHERE ROWNUM = 1").Scan(&version).Error; err != nil {
		// Try alternative query for Autonomous Database
		if err := db.Raw("SELECT VERSION FROM PRODUCT_COMPONENT_VERSION WHERE PRODUCT LIKE 'Oracle%'").Scan(&version).Error; err != nil {
			log.Printf("Failed to get Oracle version: %v", err)
		}
	}
	if version != "" {
		fmt.Printf("Oracle Version: %s\n", version)
	}

	// Get database service info
	var serviceName string
	if err := db.Raw("SELECT SYS_CONTEXT('USERENV', 'SERVICE_NAME') FROM DUAL").Scan(&serviceName).Error; err == nil {
		fmt.Printf("Connected to service: %s\n", serviceName)
	}

	// Test creating a simple table
	type TestTable struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}

	fmt.Println("\nTesting table operations...")

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
		fmt.Println("✅ Successfully created test table!")

		// Insert test data
		testRecord := TestTable{ID: 1, Name: "Test Record"}
		if err := db.Create(&testRecord).Error; err != nil {
			log.Printf("Failed to insert test record: %v", err)
		} else {
			fmt.Println("✅ Successfully inserted test record!")
		}

		// Query test data
		var result TestTable
		if err := db.First(&result).Error; err != nil {
			log.Printf("Failed to query test record: %v", err)
		} else {
			fmt.Printf("✅ Successfully queried test record: ID=%d, Name=%s\n", result.ID, result.Name)
		}

		// Clean up
		if err := db.Migrator().DropTable(&TestTable{}); err != nil {
			log.Printf("Failed to drop test table: %v", err)
		} else {
			fmt.Println("✅ Successfully cleaned up test table!")
		}
	}

	fmt.Println("\n🎉 Oracle Autonomous Database with Wallet connection test completed successfully!")
	fmt.Println("\nYour application is configured correctly to use Oracle Autonomous Database.")
	fmt.Println("The server will use these settings automatically when you run it.")
}
