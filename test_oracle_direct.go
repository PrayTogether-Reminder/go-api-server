package main

import (
	"database/sql"
	"fmt"
	"log"
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
	tnsAlias := os.Getenv("TNS_ALIAS")

	if user == "" {
		user = "ADMIN"
	}
	if tnsAdmin == "" {
		tnsAdmin = "./resources/main-wallet"
	}
	if tnsAlias == "" {
		tnsAlias = "z5f5ees1n47gddba_high"
	}

	fmt.Printf("Oracle Connection Configuration:\n")
	fmt.Printf("  User: %s\n", user)
	fmt.Printf("  TNS_ADMIN: %s\n", tnsAdmin)
	fmt.Printf("  TNS_ALIAS: %s\n", tnsAlias)

	// Get absolute path for TNS_ADMIN
	absPath, err := filepath.Abs(tnsAdmin)
	if err != nil {
		log.Fatalf("Failed to get absolute path for TNS_ADMIN: %v", err)
	}
	fmt.Printf("  TNS_ADMIN (absolute): %s\n", absPath)

	// Method 1: Using go-ora with wallet
	fmt.Println("\n=== Method 1: Using go-ora with wallet configuration ===")

	// Set TNS_ADMIN environment variable
	os.Setenv("TNS_ADMIN", absPath)

	// Build connection URL for go-ora with wallet
	// Format: oracle://user:password@host:port/service?WALLET=wallet_path
	walletPath := filepath.Join(absPath, "cwallet.sso")

	// Using the TNS descriptor directly
	connStr := fmt.Sprintf("oracle://%s:%s@adb.ap-chuncheon-1.oraclecloud.com:1522/g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com?SSL=true&SSL_VERIFY=false&WALLET=%s",
		user, password, walletPath)

	fmt.Printf("Connecting with URL (masked): oracle://%s:****@...?SSL=true&WALLET=%s\n", user, walletPath)

	db, err := sql.Open("oracle", connStr)
	if err != nil {
		log.Printf("Failed to open connection: %v", err)
	} else {
		defer db.Close()

		// Test the connection
		if err := db.Ping(); err != nil {
			log.Printf("Failed to ping database: %v", err)
		} else {
			fmt.Println("✅ Successfully connected using go-ora with wallet!")

			// Get version
			var version string
			err := db.QueryRow("SELECT BANNER FROM V$VERSION WHERE ROWNUM = 1").Scan(&version)
			if err != nil {
				// Try alternative query
				err = db.QueryRow("SELECT VERSION FROM PRODUCT_COMPONENT_VERSION WHERE PRODUCT LIKE 'Oracle%'").Scan(&version)
			}
			if err == nil {
				fmt.Printf("Oracle Version: %s\n", version)
			}
		}
	}

	// Method 2: Using TNS string directly
	fmt.Println("\n=== Method 2: Using full TNS descriptor ===")

	tnsString := "(description= (retry_count=20)(retry_delay=3)(address=(protocol=tcps)(port=1522)(host=adb.ap-chuncheon-1.oraclecloud.com))(connect_data=(service_name=g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com))(security=(ssl_server_dn_match=yes)))"

	// Build DSN with TNS string
	dsn := fmt.Sprintf("oracle://%s:%s@%s?WALLET=%s", user, password, tnsString, absPath)

	db2, err := sql.Open("oracle", dsn)
	if err != nil {
		log.Printf("Failed to open connection with TNS: %v", err)
	} else {
		defer db2.Close()

		if err := db2.Ping(); err != nil {
			log.Printf("Failed to ping with TNS: %v", err)
		} else {
			fmt.Println("✅ Successfully connected using full TNS descriptor!")
		}
	}

	// Method 3: Using simple TNS alias (requires proper TNS_ADMIN)
	fmt.Println("\n=== Method 3: Using TNS alias with TNS_ADMIN ===")

	// Ensure TNS_ADMIN is set
	os.Setenv("TNS_ADMIN", absPath)

	// Simple connection string with TNS alias
	simpleConn := fmt.Sprintf("%s/%s@%s", user, password, tnsAlias)
	fmt.Printf("Connecting with: %s/****@%s\n", user, tnsAlias)

	db3, err := sql.Open("oracle", simpleConn)
	if err != nil {
		log.Printf("Failed to open connection with TNS alias: %v", err)
	} else {
		defer db3.Close()

		if err := db3.Ping(); err != nil {
			log.Printf("Failed to ping with TNS alias: %v", err)
		} else {
			fmt.Println("✅ Successfully connected using TNS alias!")
		}
	}
}
