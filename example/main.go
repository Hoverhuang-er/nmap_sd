package main

import (
	"nmap_sd/pkg/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Set log format
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

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
	log.Info("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
