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
	Type     string // oracle, postgres
	Host     string
	Port     int
	User     string
	Password string
	Service  string // Oracle Service Name or PostgreSQL DB Name
	Charset  string
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
			Type:     getEnv("DB_TYPE", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Service:  getEnv("DB_NAME", "praytogether"),
			Charset:  getEnv("DB_CHARSET", "UTF8"),
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
		return fmt.Sprintf(`user="%s" password="%s" connectString="%s:%d/%s"`,
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
	default:
		return ""
	}
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
