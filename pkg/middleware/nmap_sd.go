package middleware

import (
	"html/template"
	"log/slog"
	"os"
	"strings"
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
	hostInfo    []sd.HostInfo
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
	// Log level: "INFO", "ERROR", "DEBUG" (default: "INFO")
	LogLevel string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		CIDR:         "192.168.2.0/22",
		ScanPath:     "/mgsd",
		ScanInterval: 1,
		LogLevel:     "INFO",
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
	slog.Debug("New: Creating NmapSD middleware instance")
	cfg := DefaultConfig()
	if len(config) > 0 {
		slog.Debug("New: Using custom configuration")
		cfg = config[0]
		if cfg.CIDR == "" {
			slog.Debug("New: CIDR empty, using default", "default", "192.168.2.0/22")
			cfg.CIDR = "192.168.2.0/22"
		}
		if cfg.ScanPath == "" {
			slog.Debug("New: ScanPath empty, using default", "default", "/mgsd")
			cfg.ScanPath = "/mgsd"
		}
		if cfg.ScanInterval <= 0 {
			slog.Debug("New: ScanInterval invalid, using default", "default", 1)
			cfg.ScanInterval = 1
		}
		if len(cfg.Ports) == 0 {
			slog.Debug("New: Ports empty, using default ports")
			cfg.Ports = DefaultConfig().Ports
		}
		if cfg.LogLevel == "" {
			slog.Debug("New: LogLevel empty, using default", "default", "INFO")
			cfg.LogLevel = "INFO"
		}
	} else {
		slog.Debug("New: Using default configuration")
	}

	slog.Debug("New: Final configuration", "cidr", cfg.CIDR, "scanPath", cfg.ScanPath, "scanInterval", cfg.ScanInterval, "logLevel", cfg.LogLevel, "portCount", len(cfg.Ports))

	// Set log level
	slog.Debug("New: Setting log level", "level", cfg.LogLevel)
	setLogLevel(cfg.LogLevel)

	slog.Debug("New: Creating NmapSD instance")
	nsd := &NmapSD{
		cidr:     cfg.CIDR,
		scanPath: cfg.ScanPath,
		ports:    cfg.Ports,
		data:     []sd.ServiceTarget{},
	}
	slog.Debug("New: NmapSD instance created")

	// Start background scanner
	slog.Debug("New: Creating scheduler", "interval_minutes", cfg.ScanInterval)
	nsd.scheduler = gocron.NewScheduler(time.Local)
	nsd.scheduler.Every(cfg.ScanInterval).Minutes().Do(func() {
		nsd.performScan()
	})
	slog.Debug("New: Starting scheduler asynchronously")
	nsd.scheduler.StartAsync()

	// Perform initial scan
	slog.Debug("New: Launching initial scan in background")
	go nsd.performScan()

	slog.Debug("New: Middleware handler created successfully")
	return func(c *gin.Context) {
		slog.Debug("Middleware: Request received", "path", c.Request.URL.Path, "method", c.Request.Method)
		// Check if this is the scan result endpoint
		if c.Request.URL.Path == cfg.ScanPath && c.Request.Method == "GET" {
			slog.Debug("Middleware: Handling scan result request")
			nsd.handleScanResult(c)
			return
		}
		// Check if this is the info endpoint
		if c.Request.URL.Path == "/info" && c.Request.Method == "GET" {
			slog.Debug("Middleware: Handling info page request")
			nsd.handleInfo(c)
			return
		}
		slog.Debug("Middleware: Passing request to next handler")
		c.Next()
	}
}

// performScan executes the network scan and updates data
func (n *NmapSD) performScan() {
	slog.Debug("performScan: Starting network scan", "cidr", n.cidr, "port_count", len(n.ports))
	slog.Info("Starting network scan...")

	slog.Debug("performScan: Calling ScanNetworkRange")
	results, hostInfo, err := sd.ScanNetworkRange(n.cidr, n.ports)
	if err != nil {
		slog.Error("Failed to scan network", "error", err)
		slog.Debug("performScan: Scan failed, returning without updating data")
		return
	}
	slog.Debug("performScan: Scan completed successfully", "service_groups", len(results), "hosts", len(hostInfo))

	slog.Debug("performScan: Acquiring data mutex lock")
	n.dataMutex.Lock()
	slog.Debug("performScan: Updating scan results")
	n.data = results
	n.hostInfo = hostInfo
	n.initialized = true
	slog.Debug("performScan: Results updated, releasing mutex")
	n.dataMutex.Unlock()

	slog.Info("Scan completed", "service_groups", len(results), "hosts", len(hostInfo))
	slog.Debug("performScan: Network scan finished")
}

// handleScanResult returns the current scan results
func (n *NmapSD) handleScanResult(c *gin.Context) {
	slog.Debug("handleScanResult: Acquiring read lock")
	n.dataMutex.RLock()
	data := n.data
	initialized := n.initialized
	n.dataMutex.RUnlock()
	slog.Debug("handleScanResult: Read lock released", "initialized", initialized, "data_count", len(data))

	if !initialized {
		slog.Debug("handleScanResult: Scan not initialized, returning empty array")
		c.JSON(200, []sd.ServiceTarget{})
		return
	}

	slog.Debug("handleScanResult: Returning scan results", "service_groups", len(data))
	c.JSON(200, data)
}

// handleInfo renders an HTML page with host information
func (n *NmapSD) handleInfo(c *gin.Context) {
	slog.Debug("handleInfo: Acquiring read lock")
	n.dataMutex.RLock()
	hostInfo := n.hostInfo
	initialized := n.initialized
	n.dataMutex.RUnlock()
	slog.Debug("handleInfo: Read lock released", "initialized", initialized, "host_count", len(hostInfo))

	if !initialized {
		slog.Debug("handleInfo: Scan not initialized, returning scanning message")
		c.Data(200, "text/html; charset=utf-8", []byte("<h1>Scanning in progress...</h1>"))
		return
	}

	slog.Debug("handleInfo: Parsing HTML template")
	tmpl := template.Must(template.New("info").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Network Scan Results</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        h1 {
            color: #333;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            background-color: white;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #4CAF50;
            color: white;
            font-weight: bold;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .port-list {
            margin: 0;
            padding-left: 20px;
        }
        .port-item {
            margin: 4px 0;
        }
        .no-data {
            text-align: center;
            padding: 40px;
            color: #666;
        }
        .timestamp {
            color: #666;
            font-size: 14px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <h1>Network Scan Results</h1>
    <div class="timestamp">Last updated: {{.Timestamp}}</div>
    {{if .Hosts}}
    <table>
        <thead>
            <tr>
                <th>IP Address</th>
                <th>Hostname</th>
                <th>Operating System</th>
                <th>Open Ports</th>
            </tr>
        </thead>
        <tbody>
            {{range .Hosts}}
            <tr>
                <td>{{.IP}}</td>
                <td>{{if .Hostname}}{{.Hostname}}{{else}}-{{end}}</td>
                <td>{{if .OS}}{{.OS}}{{else}}Unknown{{end}}</td>
                <td>
                    <ul class="port-list">
                        {{range .Ports}}
                        <li class="port-item">
                            <strong>{{.Port}}</strong> ({{.State}})
                            {{if .Service}} - {{.Service}}{{end}}
                        </li>
                        {{end}}
                    </ul>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{else}}
    <div class="no-data">No hosts found</div>
    {{end}}
</body>
</html>
`))

	slog.Debug("handleInfo: Template parsed successfully")
	data := map[string]interface{}{
		"Hosts":     hostInfo,
		"Timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}
	slog.Debug("handleInfo: Rendering template with data", "host_count", len(hostInfo))

	c.Header("Content-Type", "text/html; charset=utf-8")
	err := tmpl.Execute(c.Writer, data)
	if err != nil {
		slog.Error("handleInfo: Template execution failed", "error", err)
	} else {
		slog.Debug("handleInfo: Template rendered successfully")
	}
}

// Stop gracefully stops the scheduler
func (n *NmapSD) Stop() {
	slog.Debug("Stop: Stopping scheduler")
	if n.scheduler != nil {
		n.scheduler.Stop()
		slog.Debug("Stop: Scheduler stopped successfully")
	} else {
		slog.Debug("Stop: No scheduler to stop")
	}
}

// setLogLevel sets the global log level based on the configuration
func setLogLevel(level string) {
	slog.Debug("setLogLevel: Setting log level", "requested_level", level)
	var logLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
		slog.Debug("setLogLevel: DEBUG level selected")
	case "ERROR":
		logLevel = slog.LevelError
		slog.Debug("setLogLevel: ERROR level selected")
	case "INFO":
		logLevel = slog.LevelInfo
		slog.Debug("setLogLevel: INFO level selected")
	default:
		logLevel = slog.LevelInfo
		slog.Debug("setLogLevel: Unknown level, defaulting to INFO", "provided_level", level)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Log after setting the new level
	slog.Debug("setLogLevel: Log level configured", "level", level)
	slog.Info("Log level set", "level", strings.ToUpper(level))
}
