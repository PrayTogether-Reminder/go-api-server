package router

import (
	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/meta"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"github.com/gin-gonic/gin"
)

// Setup configures all application-specific routes using dependency injection
func Setup(router *gin.Engine, cfg *config.Config, db *database.DB) {
	// Meta handler (health check, app version, legal documents)
	metaHandler := meta.NewHandler(cfg, db)
	router.GET("/health", metaHandler.Health)

	// API v1 routes
	// Domain routes will be added here when implementing features
	// Example:
	// v1 := router.Group("/api/v1")
	// v1.POST("/members", memberHandler.Create)
	// v1.POST("/rooms", roomHandler.Create)
}
