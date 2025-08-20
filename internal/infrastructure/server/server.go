package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/database"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	cfg    *config.Config
	db     *database.DB
	router *gin.Engine
	server *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, db *database.DB) *Server {
	// Gin 모드 설정
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Gin 기본 로거 비활성화 (slog 사용)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	// 라우터 생성 (기본 미들웨어 없이)
	router := gin.New()

	// Recovery 미들웨어 (패닉 복구)
	router.Use(gin.CustomRecovery(recoveryHandler))

	// slog 로깅 미들웨어
	router.Use(LoggerMiddleware(cfg))

	// DB 미들웨어
	router.Use(database.Middleware(db))

	return &Server{
		cfg:    cfg,
		db:     db,
		router: router,
	}
}

// recoveryHandler handles panics
func recoveryHandler(c *gin.Context, recovered interface{}) {
	if err, ok := recovered.(string); ok {
		slog.Error("Panic recovered",
			"error", err,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": "Internal server error",
	})
}

// SetupRoutes sets up all routes
func (s *Server) SetupRoutes() {
	// Health check endpoints
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/ready", s.readinessCheck)

	// API v1 routes
	v1 := s.router.Group("/api/v1")
	{
		// API routes will be added here
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// readinessCheck handles readiness check requests
func (s *Server) readinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	if err := s.db.HealthCheck(ctx); err != nil {
		slog.Error("Readiness check failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", s.cfg.App.Port),
		Handler:        s.router,
		ReadTimeout:    s.cfg.Server.ReadTimeout,
		WriteTimeout:   s.cfg.Server.WriteTimeout,
		IdleTimeout:    s.cfg.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	slog.Info("Starting server",
		"port", s.cfg.App.Port,
		"env", s.cfg.App.Env,
		"read_timeout", s.cfg.Server.ReadTimeout,
		"write_timeout", s.cfg.Server.WriteTimeout,
	)

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	slog.Info("Shutting down server...")
	return s.server.Shutdown(ctx)
}

// Router returns the gin router (for testing)
func (s *Server) Router() *gin.Engine {
	return s.router
}
