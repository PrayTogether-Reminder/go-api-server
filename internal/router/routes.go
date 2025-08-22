package router

import (
	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"net/http"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

// Setup configures all application-specific routes using dependency injection
// This follows Clean Architecture principles where dependencies are injected
func Setup(router *gin.Engine, cfg *config.Config, db *database.DB) {
	// Initialize repositories

	// Initialize service

	// Initialize use case

	// Health check endpoints (moved from bootstrap to maintain Clean Architecture)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Example endpoint
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

	}
}
