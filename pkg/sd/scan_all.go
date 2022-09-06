package sd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Ullaakut/nmap/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

func ScanNetworkWNmap() error {
	fs, err := os.OpenFile("mg_sd_dip.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Errorf("failed to open file: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	log.Info("Init nmap")
	scaner, err := nmap.NewScanner(
		nmap.WithTargets("192.168.2.3", "192.168.2.5", "192.168.2.6", "192.168.2.8", "192.168.2.9",
			"192.168.2.10", "192.168.2.11", "192.168.2.13", "192.168.2.14", "192.168.2.15", "192.168.2.16",
			"192.168.2.17", "192.168.2.18", "192.168.2.20", "192.168.2.21", "192.168.2.23", "192.168.2.26",
			"192.168.2.27", "192.168.2.28", "192.168.2.33", "192.168.2.35", "192.168.2.41", "192.168.2.43",
			"192.168.2.74", "192.168.2.92", "192.168.2.99", "192.168.2.131", "192.168.2.135", "192.168.2.145"),
		nmap.WithContext(ctx),
		nmap.WithOSDetection(),
		nmap.WithPorts("9182"))
	if err != nil {
		log.Errorf("failed to create nmap scanner: %v", err)
		return nil
	}
	log.Info("start to scan network")
	result, warnings, err := scaner.Run()
	if err != nil {
		log.Errorf("failed to run nmap scan: %v", err)
		return nil
	}
	log.Info("scan network finished")
	if warnings != nil {
		for _, w := range warnings {
			log.Warnf("nmap warning: %v", w)
		}
	}
	log.Info("loop through result")
	var wmitarget []string
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			log.Warn("host has no ports or addresses")
			continue
		}
		for _, port := range host.Ports {
			if port.ID == 9143 || port.Protocol == "tcp" {
				log.Infof("Finding port %d/%s with wmi_exporter", port.ID, port.Protocol)
				wmitarget = append(wmitarget, fmt.Sprintf("%s:%d", host.Addresses[0].String(), port.ID))
			}
		}
	}
	log.Info("capacity of host addr and port:%d", len(wmitarget))
	wmijson, err := jsoniter.MarshalToString([]RegisterTarget{
		{
			Targets: wmitarget,
			Labels: Labels{
				Env: "prod",
				Job: "mg_machine",
			},
		},
	})
	if err != nil {
		log.Errorf("failed to marshal to json: %v", err)
		return nil
	}
	log.Infof("wmi_exporter json: %s", wmijson)
	if _, err := fs.WriteString(wmijson); err != nil {
		log.Errorf("failed to write to file: %v", err)
		return nil
	}
	log.Info("write to file finished")
	fs.Close()
	return nil
}
