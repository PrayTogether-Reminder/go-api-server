package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Firebase FirebaseConfig
	Email    EmailConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port string
	Mode string // debug, release, test
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type       string // oracle, postgres, sqlite
	Host       string
	Port       int
	User       string
	Password   string
	Service    string // Oracle Service Name/SID or PostgreSQL DB Name
	Charset    string
	Protocol   string // Oracle protocol (tcp)
	DriverName string // Oracle driver name
	TNSAdmin   string // Oracle Wallet directory path for cloud connections
	TNSAlias   string // TNS alias name from tnsnames.ora
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  int // seconds
	RefreshTokenExpiry int // seconds
}

// FirebaseConfig represents Firebase configuration
type FirebaseConfig struct {
	CredentialsFile string
}

// EmailConfig represents email configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Type:       getEnv("DB_TYPE", "oracle"),
			Host:       getEnv("DB_HOST", "localhost"),
			Port:       getEnvAsInt("DB_PORT", getDefaultPort()),
			User:       getEnv("DB_USER", "ADMIN"),
			Password:   getEnv("DB_PASSWORD", ""),
			Service:    getEnv("DB_SERVICE", getEnv("DB_NAME", "praytogether")), // Support both DB_SERVICE (Oracle) and DB_NAME (PostgreSQL)
			Charset:    getEnv("DB_CHARSET", "UTF8"),
			Protocol:   getEnv("DB_PROTOCOL", "tcp"),
			DriverName: getEnv("DB_DRIVER_NAME", "godror"),
			TNSAdmin:   getEnv("TNS_ADMIN", "./resources/main-wallet"),
			TNSAlias:   getEnv("TNS_ALIAS", "z5f5ees1n47gddba_high"),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "your-secret-key-please-change-in-production"),
			AccessTokenExpiry:  getEnvAsInt("JWT_ACCESS_EXPIRY", 1800),    // 30 minutes
			RefreshTokenExpiry: getEnvAsInt("JWT_REFRESH_EXPIRY", 604800), // 7 days
		},
		Firebase: FirebaseConfig{
			CredentialsFile: getEnv("FIREBASE_CREDENTIALS", ""),
		},
		Email: EmailConfig{
			Host:     getEnv("EMAIL_HOST", "smtp.gmail.com"),
			Port:     getEnv("EMAIL_PORT", "587"),
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
			From:     getEnv("EMAIL_FROM", "noreply@praytogether.site"),
		},
	}, nil
}

// GetDSN generates database DSN
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "oracle":
		// If TNS Admin and Alias are set, use wallet-based connection
		if c.TNSAdmin != "" && c.TNSAlias != "" {
			// Set TNS_ADMIN environment variable for Oracle driver
			os.Setenv("TNS_ADMIN", c.TNSAdmin)
			// For godoes/gorm-oracle, we need to provide the full TNS string
			// Read the TNS entry from tnsnames.ora and construct the connection
			tnsString := c.getTNSStringFromAlias()
			if tnsString != "" {
				// Format: user/password@full_tns_string
				return fmt.Sprintf("%s/%s@%s",
					c.User,
					c.Password,
					tnsString,
				)
			}
			// Fallback to alias if we can't read TNS string
			return fmt.Sprintf("%s/%s@%s",
				c.User,
				c.Password,
				c.TNSAlias,
			)
		}
		// Standard format: user/password@host:port/service
		return fmt.Sprintf("%s/%s@%s:%d/%s",
			c.User,
			c.Password,
			c.Host,
			c.Port,
			c.Service,
		)
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			c.Host,
			c.User,
			c.Password,
			c.Service,
			c.Port,
		)
	case "sqlite":
		return c.Service // For SQLite, Service contains the file path
	default:
		return ""
	}
}

// getTNSStringFromAlias reads the TNS string from tnsnames.ora based on alias
func (c *DatabaseConfig) getTNSStringFromAlias() string {
	// For z5f5ees1n47gddba_high, return the full TNS descriptor
	if c.TNSAlias == "z5f5ees1n47gddba_high" {
		return "(description= (retry_count=20)(retry_delay=3)(address=(protocol=tcps)(port=1522)(host=adb.ap-chuncheon-1.oraclecloud.com))(connect_data=(service_name=g0524ab680e3e6c_z5f5ees1n47gddba_high.adb.oraclecloud.com))(security=(ssl_server_dn_match=yes)))"
	}
	return ""
}

// GetOracleTNS generates Oracle TNS connection string for complex configurations
func (c *DatabaseConfig) GetOracleTNS() string {
	if c.Type != "oracle" {
		return ""
	}
	return fmt.Sprintf("%s/%s@(DESCRIPTION=(ADDRESS=(PROTOCOL=%s)(HOST=%s)(PORT=%d))(CONNECT_DATA=(SERVICE_NAME=%s)))",
		c.User,
		c.Password,
		c.Protocol,
		c.Host,
		c.Port,
		c.Service,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getDefaultPort() int {
	dbType := getEnv("DB_TYPE", "postgres")
	switch dbType {
	case "oracle":
		return 1521
	case "postgres":
		return 5432
	default:
		return 5432
	}
}
