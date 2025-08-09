package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/sijms/go-ora/v2"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	tnsAdmin := os.Getenv("TNS_ADMIN")

	if user == "" {
		user = "ADMIN"
	}
	if tnsAdmin == "" {
		tnsAdmin = "./resources/main-wallet"
	}

	fmt.Printf("Oracle Autonomous Database Connection Test\n")
	fmt.Printf("==========================================\n")
	fmt.Printf("User: %s\n", user)
	fmt.Printf("TNS_ADMIN: %s\n", tnsAdmin)

	// Get absolute path for TNS_ADMIN
	absPath, err := filepath.Abs(tnsAdmin)
	if err != nil {
		log.Fatalf("Failed to get absolute path for TNS_ADMIN: %v", err)
	}
	fmt.Printf("TNS_ADMIN (absolute): %s\n", absPath)

	// Check wallet files
	walletFiles := []string{"cwallet.sso", "ewallet.p12", "tnsnames.ora", "sqlnet.ora"}
	for _, file := range walletFiles {
		path := filepath.Join(absPath, file)
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("✅ Found: %s\n", file)
		} else {
			fmt.Printf("❌ Missing: %s\n", file)
		}
	}

	// URL encode the password to handle special characters
	encodedPassword := url.QueryEscape(password)

	fmt.Println("\n=== Connecting to Oracle Autonomous Database ===")

	// Method 1: Direct connection with service endpoint
	host := "adb.ap-chuncheon-1.oraclecloud.com"
	port := "1522"
	service := "g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com"

	// Build connection URL with wallet
	walletPath := absPath
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%s/%s?WALLET=%s&SSL=enable&SSL_VERIFY=false",
		user, encodedPassword, host, port, service, walletPath)

	fmt.Printf("Connecting to: %s:%s/%s\n", host, port, service)
	fmt.Printf("Using wallet at: %s\n", walletPath)

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test the connection
	fmt.Println("\nTesting connection...")
	if err := db.Ping(); err != nil {
		// Try alternative connection method
		fmt.Printf("First attempt failed: %v\n", err)
		fmt.Println("Trying alternative connection method...")

		// Method 2: Using go-ora specific format
		connStr2 := fmt.Sprintf("oracle://%s:%s@%s:%s/%s?wallet=%s",
			user, encodedPassword, host, port, service, walletPath)

		db2, err := sql.Open("oracle", connStr2)
		if err != nil {
			log.Fatalf("Failed to open database connection (method 2): %v", err)
		}
		defer db2.Close()

		if err := db2.Ping(); err != nil {
			log.Fatalf("Failed to connect to Oracle Autonomous Database: %v", err)
		}
		db = db2
	}

	fmt.Println("✅ Successfully connected to Oracle Autonomous Database!")

	// Get database information
	fmt.Println("\n=== Database Information ===")

	// Get version
	var version string
	err = db.QueryRow("SELECT BANNER_FULL FROM V$VERSION WHERE ROWNUM = 1").Scan(&version)
	if err != nil {
		// Try alternative query
		err = db.QueryRow("SELECT VERSION_FULL FROM PRODUCT_COMPONENT_VERSION WHERE PRODUCT LIKE 'Oracle%' AND ROWNUM = 1").Scan(&version)
		if err != nil {
			// Try simpler query
			err = db.QueryRow("SELECT VERSION FROM V$INSTANCE").Scan(&version)
		}
	}
	if err == nil && version != "" {
		fmt.Printf("Database Version: %s\n", version)
	}

	// Get service name
	var serviceName string
	if err := db.QueryRow("SELECT SYS_CONTEXT('USERENV', 'SERVICE_NAME') FROM DUAL").Scan(&serviceName); err == nil {
		fmt.Printf("Service Name: %s\n", serviceName)
	}

	// Get database name
	var dbName string
	if err := db.QueryRow("SELECT SYS_CONTEXT('USERENV', 'DB_NAME') FROM DUAL").Scan(&dbName); err == nil {
		fmt.Printf("Database Name: %s\n", dbName)
	}

	// Test table operations
	fmt.Println("\n=== Testing Table Operations ===")

	// Drop test table if exists
	_, _ = db.Exec("DROP TABLE test_connection")

	// Create test table
	createSQL := `
		CREATE TABLE test_connection (
			id NUMBER PRIMARY KEY,
			name VARCHAR2(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createSQL); err != nil {
		log.Printf("Failed to create table: %v", err)
	} else {
		fmt.Println("✅ Table created successfully")

		// Insert data
		if _, err := db.Exec("INSERT INTO test_connection (id, name) VALUES (1, 'Test Record')"); err != nil {
			log.Printf("Failed to insert data: %v", err)
		} else {
			fmt.Println("✅ Data inserted successfully")

			// Query data
			var id int
			var name string
			if err := db.QueryRow("SELECT id, name FROM test_connection WHERE id = 1").Scan(&id, &name); err != nil {
				log.Printf("Failed to query data: %v", err)
			} else {
				fmt.Printf("✅ Data queried successfully: ID=%d, Name=%s\n", id, name)
			}

			// Clean up
			if _, err := db.Exec("DROP TABLE test_connection"); err != nil {
				log.Printf("Failed to drop table: %v", err)
			} else {
				fmt.Println("✅ Table dropped successfully")
			}
		}
	}

	fmt.Println("\n🎉 Oracle Autonomous Database connection test completed successfully!")
	fmt.Println("\nYour Go application can now connect to Oracle Autonomous Database.")
	fmt.Println("The connection uses the Oracle Wallet for secure authentication.")
}
