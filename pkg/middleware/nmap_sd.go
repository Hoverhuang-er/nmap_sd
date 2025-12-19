package middleware

import (
	"log/slog"
	"sync"
	"time"

	"github.com/Hoverhuang-er/nmap_sd/pkg/sd"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
)

// NmapSD Gin middleware for network service discovery
type NmapSD struct {
	cidr        string
	scanPath    string
	ports       []sd.PortService
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
	// Ports to scan (default: common ports)
	Ports []sd.PortService
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		CIDR:         "192.168.2.0/22",
		ScanPath:     "/mgsd",
		ScanInterval: 1,
		Ports: []sd.PortService{
			{Port: 9182, Name: "windows_exporter", Job: "windows_exporter"},
			{Port: 80, Name: "http", Job: "http_services"},
			{Port: 443, Name: "https", Job: "http_services"},
			{Port: 8080, Name: "http-proxy", Job: "http_services"},
			{Port: 8083, Name: "http-alt", Job: "http_services"},
			{Port: 8089, Name: "http-alt", Job: "http_services"},
			{Port: 8888, Name: "http-alt", Job: "http_services"},
			{Port: 38089, Name: "custom", Job: "http_services"},
		},
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
		if len(cfg.Ports) == 0 {
			cfg.Ports = DefaultConfig().Ports
		}
	}

	nsd := &NmapSD{
		cidr:     cfg.CIDR,
		scanPath: cfg.ScanPath,
		ports:    cfg.Ports,
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
	slog.Info("Starting network scan...")
	results, err := sd.ScanNetworkRange(n.cidr, n.ports)
	if err != nil {
		slog.Error("Failed to scan network", "error", err)
		return
	}

	n.dataMutex.Lock()
	n.data = results
	n.initialized = true
	n.dataMutex.Unlock()

	slog.Info("Scan completed", "service_groups", len(results))
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
