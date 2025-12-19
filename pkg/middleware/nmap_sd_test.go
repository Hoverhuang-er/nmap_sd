package middleware

import (
	"testing"

	"github.com/Hoverhuang-er/nmap_sd/pkg/sd"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CIDR != "192.168.2.0/22" {
		t.Errorf("Expected default CIDR to be 192.168.2.0/22, got %s", cfg.CIDR)
	}

	if cfg.ScanPath != "/mgsd" {
		t.Errorf("Expected default ScanPath to be /mgsd, got %s", cfg.ScanPath)
	}

	if cfg.ScanInterval != 1 {
		t.Errorf("Expected default ScanInterval to be 1, got %d", cfg.ScanInterval)
	}

	if len(cfg.Ports) == 0 {
		t.Error("Expected Ports to have default values")
	}

	// Check for expected default ports
	expectedPorts := map[uint16]bool{
		9182: true,
		80:   true,
		443:  true,
	}

	foundPorts := make(map[uint16]bool)
	for _, ps := range cfg.Ports {
		foundPorts[ps.Port] = true
	}

	for port := range expectedPorts {
		if !foundPorts[port] {
			t.Errorf("Expected port %d to be in default Ports", port)
		}
	}
}

func TestCustomPorts(t *testing.T) {
	customPorts := []sd.PortService{
		{Port: 3000, Name: "custom-app", Job: "custom_services"},
		{Port: 5000, Name: "another-app", Job: "custom_services"},
	}

	cfg := Config{
		CIDR:         "10.0.0.0/24",
		ScanPath:     "/custom-scan",
		ScanInterval: 5,
		Ports:        customPorts,
	}

	if len(cfg.Ports) != 2 {
		t.Errorf("Expected 2 custom ports, got %d", len(cfg.Ports))
	}

	if cfg.Ports[0].Port != 3000 {
		t.Errorf("Expected first port to be 3000, got %d", cfg.Ports[0].Port)
	}
}
