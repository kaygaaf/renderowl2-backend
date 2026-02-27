package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kaygaaf/renderowl2/internal/handlers"
	"github.com/kaygaaf/renderowl2/internal/repository"
	"github.com/kaygaaf/renderowl2/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = gin.ReleaseMode
	}
	gin.SetMode(ginMode)

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := migrateDB(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize repositories
	timelineRepo := repository.NewTimelineRepository(db)

	// Initialize services
	timelineService := service.NewTimelineService(timelineRepo)

	// Initialize handlers
	timelineHandler := handlers.NewTimelineHandler(timelineService)

	// Setup router
	router := setupRouter(timelineHandler)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Renderowl 2.0 API starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// initDB initializes the database connection
func initDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default local development connection
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "postgres")
		dbname := getEnv("DB_NAME", "renderowl2")
		sslmode := getEnv("DB_SSLMODE", "disable")

		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	if os.Getenv("DB_DEBUG") == "true" {
		config.Logger = logger.Default.LogMode(logger.Info)
	}

	return gorm.Open(postgres.Open(dsn), config)
}

// migrateDB runs auto-migration for all models
func migrateDB(db *gorm.DB) error {
	// Import domain models for migration
	// This will be expanded as we add more models
	return db.AutoMigrate(
		// Timeline models will be added here
	)
}

// setupRouter configures the Gin router and routes
func setupRouter(timelineHandler *handlers.TimelineHandler) *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", healthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Timeline routes
		v1.POST("/timeline", timelineHandler.CreateTimeline)
		v1.GET("/timeline/:id", timelineHandler.GetTimeline)
		v1.PUT("/timeline/:id", timelineHandler.UpdateTimeline)
		v1.DELETE("/timeline/:id", timelineHandler.DeleteTimeline)
		v1.GET("/timelines", timelineHandler.ListTimelines)
		v1.GET("/timelines/me", timelineHandler.GetUserTimelines)
	}

	return router
}

// healthCheck handles the health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "renderowl2-api",
		"timestamp": time.Now().UTC(),
		"version":   "2.0.0",
	})
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
