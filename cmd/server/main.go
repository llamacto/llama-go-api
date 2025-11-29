package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/llamacto/llama-gin-kit/config"
	"github.com/llamacto/llama-gin-kit/pkg/container"
	"github.com/llamacto/llama-gin-kit/pkg/database"
	"github.com/llamacto/llama-gin-kit/pkg/email"
	"github.com/llamacto/llama-gin-kit/pkg/jwt"
	"github.com/llamacto/llama-gin-kit/routes"
)

// @title Llamabase API
// @version 1.0
// @description An elegant Go web framework inspired by Laravel
// @host localhost:6066
// @BasePath /v1

func main() {
	// Load configuration (uses cache when available)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	container.App().Set(container.ServiceConfig, cfg)

	// Initialize JWT service
	jwt.Init(cfg)
	container.App().Set(container.ServiceJWT, jwt.MustServiceInstance())

	// Initialize email service
	email.Init(cfg)
	container.App().Set(container.ServiceEmail, email.MustServiceInstance())

	// Initialize database when enabled
	if cfg.Database.Enabled {
		db, err := database.InitDB(cfg.Database)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		container.App().Set(container.ServiceDB, db)
	} else {
		log.Println("Database initialization skipped (DB_ENABLED=false)")
	}

	// Set Gin mode from configuration
	switch strings.ToLower(cfg.Server.Mode) {
	case "release", "prod", "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	r := gin.Default()

	// Enable CORS
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
	}
	r.Use(cors.New(corsConfig))

	// Register routes
	routes.RegisterRoutes(r)

	// Start server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting server on %s", serverAddr)

	go func() {
		if err := r.Run(serverAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}
