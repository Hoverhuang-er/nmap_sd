package sd

import (
	"testing"
)

func TestServiceTarget(t *testing.T) {
	target := ServiceTarget{
		Targets: []string{"192.168.1.1:80", "192.168.1.2:80"},
		Labels: map[string]string{
			"job": "http_services",
		},
	}

	if len(target.Targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(target.Targets))
	}

	if target.Labels["job"] != "http_services" {
		t.Errorf("Expected job label to be http_services, got %s", target.Labels["job"])
	}
}

func TestPortService(t *testing.T) {
	port := PortService{
		Port: 9182,
		Name: "windows_exporter",
		Job:  "windows_exporter",
	}

	if port.Port != 9182 {
		t.Errorf("Expected port 9182, got %d", port.Port)
	}

	if port.Name != "windows_exporter" {
		t.Errorf("Expected name windows_exporter, got %s", port.Name)
	}

	if port.Job != "windows_exporter" {
		t.Errorf("Expected job windows_exporter, got %s", port.Job)
	}
}
