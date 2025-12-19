package middleware

import (
	"nmap_sd/pkg/sd"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

// NmapSD Gin middleware for network service discovery
type NmapSD struct {
	cidr        string
	scanPath    string
	scheduler   *gocron.Scheduler
	data        []sd.ServiceTarget
	dataMutex   sync.RWMutex
	initialized bool
}

// Config for NmapSD middleware
type Config struct {
	// CIDR to scan (e.g., "192.168.2.0/22")
	CIDR string
	// API path to expose scan results (default: "/mgsd")
	ScanPath string
	// Scan interval in minutes (default: 1)
	ScanInterval int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		CIDR:         "192.168.2.0/22",
		ScanPath:     "/mgsd",
		ScanInterval: 1,
	}
}

// New creates a new NmapSD middleware instance
func New(config ...Config) gin.HandlerFunc {
	cfg := DefaultConfig()
	if len(config) > 0 {
		cfg = config[0]
		if cfg.CIDR == "" {
			cfg.CIDR = "192.168.2.0/22"
		}
		if cfg.ScanPath == "" {
			cfg.ScanPath = "/mgsd"
		}
		if cfg.ScanInterval <= 0 {
			cfg.ScanInterval = 1
		}
	}

	nsd := &NmapSD{
		cidr:     cfg.CIDR,
		scanPath: cfg.ScanPath,
		data:     []sd.ServiceTarget{},
	}

	// Start background scanner
	nsd.scheduler = gocron.NewScheduler(time.Local)
	nsd.scheduler.Every(cfg.ScanInterval).Minutes().Do(func() {
		nsd.performScan()
	})
	nsd.scheduler.StartAsync()

	// Perform initial scan
	go nsd.performScan()

	return func(c *gin.Context) {
		// Check if this is the scan result endpoint
		if c.Request.URL.Path == cfg.ScanPath && c.Request.Method == "GET" {
			nsd.handleScanResult(c)
			return
		}
		c.Next()
	}
}

// performScan executes the network scan and updates data
func (n *NmapSD) performScan() {
	log.Info("Starting network scan...")
	results, err := sd.ScanNetworkRange(n.cidr)
	if err != nil {
		log.Errorf("Failed to scan network: %v", err)
		return
	}

	n.dataMutex.Lock()
	n.data = results
	n.initialized = true
	n.dataMutex.Unlock()

	log.Infof("Scan completed, found %d service groups", len(results))
}

// handleScanResult returns the current scan results
func (n *NmapSD) handleScanResult(c *gin.Context) {
	n.dataMutex.RLock()
	data := n.data
	initialized := n.initialized
	n.dataMutex.RUnlock()

	if !initialized {
		c.JSON(200, []sd.ServiceTarget{})
		return
	}

	c.JSON(200, data)
}

// Stop gracefully stops the scheduler
func (n *NmapSD) Stop() {
	if n.scheduler != nil {
		n.scheduler.Stop()
	}
}
