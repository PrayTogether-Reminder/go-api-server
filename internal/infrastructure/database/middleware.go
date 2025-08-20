package database

import (
	"github.com/gin-gonic/gin"
)

const (
	// ContextKey is the key for database in gin context
	ContextKey = "db"
)

// Middleware returns a gin middleware that injects database into context
// ORACLE_CLOUD_MINIMAL_SETUP.md의 Gin 통합 패턴 참고
func Middleware(db *DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextKey, db)
		c.Next()
	}
}

// GetDB retrieves database from gin context
func GetDB(c *gin.Context) (*DB, bool) {
	if db, exists := c.Get(ContextKey); exists {
		if database, ok := db.(*DB); ok {
			return database, true
		}
	}
	return nil, false
}
