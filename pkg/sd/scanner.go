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

// HostInfo represents detailed information about a scanned host
type HostInfo struct {
	IP       string     `json:"ip"`
	Hostname string     `json:"hostname,omitempty"`
	OS       string     `json:"os,omitempty"`
	Ports    []PortInfo `json:"ports"`
}

// PortInfo represents information about an open port
type PortInfo struct {
	Port    uint16 `json:"port"`
	State   string `json:"state"`
	Service string `json:"service,omitempty"`
}

// ScanNetworkRange scans the given CIDR range for active hosts and open ports
func ScanNetworkRange(cidr string, ports []PortService) ([]ServiceTarget, []HostInfo, error) {
	slog.Debug("ScanNetworkRange: Starting", "cidr", cidr, "port_count", len(ports))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	slog.Info("Starting nmap scan", "cidr", cidr)
	slog.Debug("ScanNetworkRange: Creating context with 10 minute timeout")

	// First scan: host discovery
	slog.Debug("ScanNetworkRange: Starting host discovery phase")
	hosts, err := discoverHosts(ctx, cidr)
	if err != nil {
		slog.Error("ScanNetworkRange: Host discovery failed", "error", err)
		return nil, nil, fmt.Errorf("host discovery failed: %w", err)
	}
	slog.Debug("ScanNetworkRange: Host discovery completed", "hosts_found", len(hosts))

	if len(hosts) == 0 {
		slog.Warn("No active hosts found")
		slog.Debug("ScanNetworkRange: Returning empty results")
		return []ServiceTarget{}, []HostInfo{}, nil
	}

	slog.Info("Found active hosts, scanning ports", "count", len(hosts))
	slog.Debug("ScanNetworkRange: Active hosts", "hosts", hosts)

	// Second scan: port detection on active hosts
	slog.Debug("ScanNetworkRange: Starting port scan phase")
	return scanPorts(ctx, hosts, ports)
}

// discoverHosts performs host discovery on the CIDR range
func discoverHosts(ctx context.Context, cidr string) ([]string, error) {
	slog.Debug("discoverHosts: Starting host discovery", "cidr", cidr)

	slog.Debug("discoverHosts: Creating nmap scanner with ping scan")
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets(cidr),
		nmap.WithPingScan(),
	)
	if err != nil {
		slog.Error("discoverHosts: Failed to create scanner", "error", err)
		return nil, fmt.Errorf("failed to create scanner: %w", err)
	}
	slog.Debug("discoverHosts: Scanner created successfully")

	slog.Debug("discoverHosts: Running nmap scan...")
	result, warnings, err := scanner.Run()
	if err != nil {
		slog.Error("discoverHosts: Scan execution failed", "error", err)
		return nil, fmt.Errorf("scan failed: %w", err)
	}
	slog.Debug("discoverHosts: Nmap scan completed", "hosts_scanned", len(result.Hosts))

	if len(*warnings) > 0 {
		slog.Debug("discoverHosts: Processing nmap warnings", "warning_count", len(*warnings))
		for _, w := range *warnings {
			slog.Warn("nmap warning", "message", w)
		}
	}

	slog.Debug("discoverHosts: Filtering active hosts")
	var activeHosts []string
	for _, host := range result.Hosts {
		if len(host.Addresses) > 0 && host.Status.State == "up" {
			ip := host.Addresses[0].String()
			activeHosts = append(activeHosts, ip)
			slog.Debug("discoverHosts: Found active host", "ip", ip, "status", host.Status.State)
		} else {
			if len(host.Addresses) > 0 {
				slog.Debug("discoverHosts: Skipping inactive host", "ip", host.Addresses[0].String(), "status", host.Status.State)
			}
		}
	}

	slog.Debug("discoverHosts: Host discovery completed", "active_hosts", len(activeHosts))
	return activeHosts, nil
}

// scanPorts scans common ports on active hosts
func scanPorts(ctx context.Context, hosts []string, ports []PortService) ([]ServiceTarget, []HostInfo, error) {
	slog.Debug("scanPorts: Starting port scan", "host_count", len(hosts), "port_count", len(ports))

	// Build port list
	slog.Debug("scanPorts: Building port list")
	var portList []string
	for _, ps := range ports {
		portList = append(portList, fmt.Sprintf("%d", ps.Port))
		slog.Debug("scanPorts: Adding port to scan", "port", ps.Port, "name", ps.Name, "job", ps.Job)
	}
	portStr := strings.Join(portList, ",")
	slog.Debug("scanPorts: Port list built", "ports", portStr)

	slog.Debug("scanPorts: Creating nmap scanner with service info and OS detection")
	scanner, err := nmap.NewScanner(
		ctx,
		nmap.WithTargets(hosts...),
		nmap.WithPorts(portStr),
		nmap.WithServiceInfo(),
		nmap.WithOSDetection(),
	)
	if err != nil {
		slog.Error("scanPorts: Failed to create port scanner", "error", err)
		return nil, nil, fmt.Errorf("failed to create port scanner: %w", err)
	}
	slog.Debug("scanPorts: Scanner created successfully")

	slog.Debug("scanPorts: Running nmap port scan...")
	result, warnings, err := scanner.Run()
	if err != nil {
		slog.Error("scanPorts: Port scan execution failed", "error", err)
		return nil, nil, fmt.Errorf("port scan failed: %w", err)
	}
	slog.Debug("scanPorts: Nmap port scan completed", "hosts_scanned", len(result.Hosts))

	if len(*warnings) > 0 {
		slog.Debug("scanPorts: Processing nmap warnings", "warning_count", len(*warnings))
		for _, w := range *warnings {
			slog.Warn("nmap warning", "message", w)
		}
	}

	slog.Debug("scanPorts: Building service targets from results")
	serviceTargets := buildServiceTargets(result, ports)
	slog.Debug("scanPorts: Service targets built", "target_groups", len(serviceTargets))

	slog.Debug("scanPorts: Building host information from results")
	hostInfos := buildHostInfos(result)
	slog.Debug("scanPorts: Host information built", "host_count", len(hostInfos))

	return serviceTargets, hostInfos, nil
}

// buildServiceTargets organizes scan results by service type
func buildServiceTargets(result *nmap.Run, ports []PortService) []ServiceTarget {
	slog.Debug("buildServiceTargets: Starting to build service targets", "total_hosts", len(result.Hosts))

	// Group targets by job
	jobMap := make(map[string][]string)
	jobLabels := make(map[string]map[string]string)

	for _, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			slog.Debug("buildServiceTargets: Skipping host with no addresses")
			continue
		}

		ip := host.Addresses[0].String()
		slog.Debug("buildServiceTargets: Processing host", "ip", ip, "port_count", len(host.Ports))

		for _, port := range host.Ports {
			if port.State.State != "open" {
				slog.Debug("buildServiceTargets: Skipping non-open port", "ip", ip, "port", port.ID, "state", port.State.State)
				continue
			}

			// Find matching port service
			for _, ps := range ports {
				if port.ID == ps.Port {
					target := fmt.Sprintf("%s:%d", ip, port.ID)
					jobMap[ps.Job] = append(jobMap[ps.Job], target)
					slog.Debug("buildServiceTargets: Matched port to job", "target", target, "job", ps.Job, "service", ps.Name)

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
						slog.Debug("buildServiceTargets: Created labels for job", "job", ps.Job, "labels", labels)
					}
					break
				}
			}
		}
	}

	slog.Debug("buildServiceTargets: Converting job map to service targets", "job_count", len(jobMap))
	// Convert to ServiceTarget slice
	var targets []ServiceTarget
	for job, targetList := range jobMap {
		if len(targetList) > 0 {
			targets = append(targets, ServiceTarget{
				Targets: targetList,
				Labels:  jobLabels[job],
			})
			slog.Debug("buildServiceTargets: Added service target", "job", job, "target_count", len(targetList))
		}
	}

	slog.Debug("buildServiceTargets: Completed building service targets", "total_service_groups", len(targets))
	return targets
}

// buildHostInfos extracts detailed host information from scan results
func buildHostInfos(result *nmap.Run) []HostInfo {
	slog.Debug("buildHostInfos: Starting to build host information", "total_hosts", len(result.Hosts))
	var hostInfos []HostInfo

	for _, host := range result.Hosts {
		if len(host.Addresses) == 0 {
			slog.Debug("buildHostInfos: Skipping host with no addresses")
			continue
		}

		ip := host.Addresses[0].String()
		slog.Debug("buildHostInfos: Processing host", "ip", ip)

		// Get hostname if available
		hostname := ""
		if len(host.Hostnames) > 0 {
			hostname = host.Hostnames[0].String()
			slog.Debug("buildHostInfos: Found hostname", "ip", ip, "hostname", hostname)
		} else {
			slog.Debug("buildHostInfos: No hostname found", "ip", ip)
		}

		// Get OS information if available
		osInfo := ""
		if len(host.OS.Matches) > 0 {
			osInfo = host.OS.Matches[0].Name
			slog.Debug("buildHostInfos: Found OS info", "ip", ip, "os", osInfo, "accuracy", host.OS.Matches[0].Accuracy)
		} else {
			slog.Debug("buildHostInfos: No OS information detected", "ip", ip)
		}

		// Collect port information
		slog.Debug("buildHostInfos: Collecting port information", "ip", ip, "total_ports", len(host.Ports))
		var portInfos []PortInfo
		for _, port := range host.Ports {
			if port.State.State != "closed" && port.State.State != "filtered" {
				portInfos = append(portInfos, PortInfo{
					Port:    port.ID,
					State:   port.State.State,
					Service: port.Service.Name,
				})
				slog.Debug("buildHostInfos: Added port info", "ip", ip, "port", port.ID, "state", port.State.State, "service", port.Service.Name)
			} else {
				slog.Debug("buildHostInfos: Skipping port", "ip", ip, "port", port.ID, "state", port.State.State)
			}
		}

		if len(portInfos) > 0 {
			hostInfos = append(hostInfos, HostInfo{
				IP:       ip,
				Hostname: hostname,
				OS:       osInfo,
				Ports:    portInfos,
			})
			slog.Debug("buildHostInfos: Added host info", "ip", ip, "port_count", len(portInfos))
		} else {
			slog.Debug("buildHostInfos: No open ports found for host, skipping", "ip", ip)
		}
	}

	slog.Debug("buildHostInfos: Completed building host information", "total_hosts_with_ports", len(hostInfos))
	return hostInfos
}
