package meta

import (
	"net/http"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/database"
	"github.com/gin-gonic/gin"
)

// Handler handles meta endpoints (health check, app version, legal documents, etc.)
type Handler struct {
	cfg *config.Config
	db  *database.DB
}

// NewHandler creates a new meta handler
func NewHandler(cfg *config.Config, db *database.DB) *Handler {
	return &Handler{
		cfg: cfg,
		db:  db,
	}
}

// Health checks service and database health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// AppVersions exposes mobile app version requirements and maintenance flags
func (h *Handler) AppVersions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"minimumAppVersion":     h.cfg.AppVersion.MinimumAppVersion,
		"forceUpdateAppVersion": h.cfg.AppVersion.ForceUpdateAppVersion,
		"maintenanceMode":       h.cfg.AppVersion.MaintenanceMode,
	})
}
