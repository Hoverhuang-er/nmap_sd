package sd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Ullaakut/nmap/v3"
)

// ServiceTarget represents a group of discovered services
type ServiceTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

// PortService defines a port and its associated service info
type PortService struct {
	Port   uint16
	Name   string
	Job    string
	Labels map[string]string
}

// ScanNetworkRange scans the given CIDR range for active hosts and open ports
func ScanNetworkRange(cidr string, ports []PortService) ([]ServiceTarget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	slog.Info("Starting nmap scan", "cidr", cidr)

	// First scan: host discovery
	hosts, err := discoverHosts(ctx, cidr)
	if err != nil {
		return nil, fmt.Errorf("host discovery failed: %w", err)
	}

	if len(hosts) == 0 {
		slog.Warn("No active hosts found")
		return []ServiceTarget{}, nil
	}

	slog.Info("Found active hosts, scanning ports", "count", len(hosts))

	// Second scan: port detection on active hosts
	return scanPorts(ctx, hosts, ports)
}

// discoverHosts performs host discovery on the CIDR range
func discoverHosts(ctx context.Context, cidr string) ([]string, error) {
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets(cidr),
		nmap.WithPingScan(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner: %w", err)
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	if len(*warnings) > 0 {
		for _, w := range *warnings {
			slog.Warn("nmap warning", "message", w)
		}
	}

	var activeHosts []string
	for _, host := range result.Hosts {
		if len(host.Addresses) > 0 && host.Status.State == "up" {
			activeHosts = append(activeHosts, host.Addresses[0].String())
		}
	}

	return activeHosts, nil
}

// scanPorts scans common ports on active hosts
func scanPorts(ctx context.Context, hosts []string, ports []PortService) ([]ServiceTarget, error) {
	// Build port list
	var portList []string
	for _, ps := range ports {
		portList = append(portList, fmt.Sprintf("%d", ps.Port))
	}
	portStr := strings.Join(portList, ",")

	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets(hosts...),
		nmap.WithPorts(portStr),
		nmap.WithServiceInfo(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create port scanner: %w", err)
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, fmt.Errorf("port scan failed: %w", err)
	}

	if len(*warnings) > 0 {
		for _, w := range *warnings {
			slog.Warn("nmap warning", "message", w)
		}
	}

	return buildServiceTargets(result, ports), nil
}

// buildServiceTargets organizes scan results by service type
func buildServiceTargets(result *nmap.Run, ports []PortService) []ServiceTarget {
	// Group targets by job
	jobMap := make(map[string][]string)
	jobLabels := make(map[string]map[string]string)

	for _, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			continue
		}

		ip := host.Addresses[0].String()

		for _, port := range host.Ports {
			if port.State.State != "open" {
				continue
			}

			// Find matching port service
			for _, ps := range ports {
				if port.ID == ps.Port {
					target := fmt.Sprintf("%s:%d", ip, port.ID)
					jobMap[ps.Job] = append(jobMap[ps.Job], target)

					// Set labels for this job
					if _, exists := jobLabels[ps.Job]; !exists {
						labels := map[string]string{
							"job": ps.Job,
						}
						// Add custom labels
						for k, v := range ps.Labels {
							labels[k] = v
						}
						jobLabels[ps.Job] = labels
					}
					break
				}
			}
		}
	}

	// Convert to ServiceTarget slice
	var targets []ServiceTarget
	for job, targetList := range jobMap {
		if len(targetList) > 0 {
			targets = append(targets, ServiceTarget{
				Targets: targetList,
				Labels:  jobLabels[job],
			})
		}
	}

	return targets
}
