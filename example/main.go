package main

import (
	"log/slog"
	"os"

	"github.com/Hoverhuang-er/nmap_sd/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Create Gin router
	r := gin.Default()

	// Register NmapSD middleware
	r.Use(middleware.New(middleware.Config{
		CIDR:         "192.168.2.0/22", // Scan this CIDR range
		ScanPath:     "/mgsd",          // Expose results at this path
		ScanInterval: 1,                // Scan every 1 minute
	}))

	// Add other routes as needed
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Network Service Discovery API",
			"version": "v2.0",
			"endpoints": gin.H{
				"discovery": "/mgsd",
				"health":    "/health",
			},
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Start server
	slog.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
